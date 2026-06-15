package search

import (
	"bufio"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/sourcegraph/conc/pool"
)

const (
	indexMagic   = "BOLT"
	indexVersion = 1
	blockSize    = 128
)

type IndexMeta struct {
	Root      string
	FileCount uint64
	BuildTime time.Time
}

type Index struct {
	meta    IndexMeta
	blocks  [][]string
	posting map[uint32][]uint32
	mu      sync.RWMutex
	added   []string
	removed map[string]struct{}
}

func trigramKey(a, b, c byte) uint32 {
	return uint32(a)<<16 | uint32(b)<<8 | uint32(c)
}

func trigrams(s string) []uint32 {
	s = strings.ToLower(s)
	if len(s) < 3 {
		return nil
	}
	seen := make(map[uint32]struct{})
	result := make([]uint32, 0, len(s)-2)
	for i := 0; i <= len(s)-3; i++ {
		k := trigramKey(s[i], s[i+1], s[i+2])
		if _, ok := seen[k]; !ok {
			seen[k] = struct{}{}
			result = append(result, k)
		}
	}
	return result
}

func IndexPath(root string) string {
	h := sha256.Sum256([]byte(root))
	dir, _ := os.UserConfigDir()
	return filepath.Join(dir, "bolt", "index", fmt.Sprintf("%x.idx", h[:8]))
}

func BuildIndex(root string, progressFn func(n int)) (*Index, error) {
	idx := &Index{
		posting: make(map[uint32][]uint32),
		removed: make(map[string]struct{}),
		meta:    IndexMeta{Root: root, BuildTime: time.Now()},
	}

	var (
		mu    sync.Mutex
		paths []string
	)

	p := pool.New().WithMaxGoroutines(runtime.NumCPU() * 2)
	var collect func(dir string)
	collect = func(dir string) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		var local []string
		for _, e := range entries {
			full := filepath.Join(dir, e.Name())
			local = append(local, full)
			if e.IsDir() {
				d := full
				p.Go(func() { collect(d) })
			}
		}
		mu.Lock()
		paths = append(paths, local...)
		if progressFn != nil {
			progressFn(len(paths))
		}
		mu.Unlock()
	}
	p.Go(func() { collect(root) })
	p.Wait()

	idx.meta.FileCount = uint64(len(paths))

	// group into blocks
	for i := 0; i < len(paths); i += blockSize {
		end := min(i+blockSize, len(paths))
		blockID := uint32(len(idx.blocks))
		block := paths[i:end]
		idx.blocks = append(idx.blocks, block)

		for _, p := range block {
			name := filepath.Base(p)
			for _, tri := range trigrams(name) {
				idx.posting[tri] = append(idx.posting[tri], blockID)
			}
		}
	}

	return idx, nil
}

func (idx *Index) Query(pattern string, matcher *Matcher, limit int, out chan<- Result) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	tris := trigrams(pattern)
	if len(tris) == 0 {
		// pattern too short for trigrams — scan all blocks
		idx.scanAll(matcher, limit, out)
		return
	}

	// find smallest posting list
	var candidate []uint32
	for _, tri := range tris {
		list := idx.posting[tri]
		if len(list) == 0 {
			return // no results
		}
		if candidate == nil || len(list) < len(candidate) {
			candidate = list
		}
	}

	// intersect with remaining lists
	for _, tri := range tris {
		list := idx.posting[tri]
		candidate = intersect(candidate, list)
		if len(candidate) == 0 {
			return
		}
	}

	count := 0
	seen := make(map[uint32]struct{})
	for _, blockID := range candidate {
		if _, ok := seen[blockID]; ok {
			continue
		}
		seen[blockID] = struct{}{}
		if int(blockID) >= len(idx.blocks) {
			continue
		}
		for _, path := range idx.blocks[blockID] {
			if _, removed := idx.removed[path]; removed {
				continue
			}
			name := filepath.Base(path)
			ok, score := matcher.Match(name)
			if ok {
				info, _ := os.Lstat(path)
				out <- Result{Path: path, Name: name, Info: info, Score: score}
				count++
				if limit > 0 && count >= limit {
					return
				}
			}
		}
	}

	// merge delta
	for _, path := range idx.added {
		name := filepath.Base(path)
		ok, score := matcher.Match(name)
		if ok {
			info, _ := os.Lstat(path)
			out <- Result{Path: path, Name: name, Info: info, Score: score}
			count++
			if limit > 0 && count >= limit {
				return
			}
		}
	}
}

func (idx *Index) scanAll(matcher *Matcher, limit int, out chan<- Result) {
	count := 0
	for _, block := range idx.blocks {
		for _, path := range block {
			name := filepath.Base(path)
			ok, score := matcher.Match(name)
			if ok {
				info, _ := os.Lstat(path)
				out <- Result{Path: path, Name: name, Info: info, Score: score}
				count++
				if limit > 0 && count >= limit {
					return
				}
			}
		}
	}
}

func intersect(a, b []uint32) []uint32 {
	set := make(map[uint32]struct{}, len(b))
	for _, v := range b {
		set[v] = struct{}{}
	}
	out := a[:0]
	for _, v := range a {
		if _, ok := set[v]; ok {
			out = append(out, v)
		}
	}
	return out
}

// SaveIndex writes the index to disk with zstd-compressed blocks.
func SaveIndex(idx *Index, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w, err := zstd.NewWriter(f)
	if err != nil {
		return err
	}
	defer w.Close()

	bw := bufio.NewWriter(w)
	// header
	bw.WriteString(indexMagic)
	bw.WriteByte(indexVersion)
	_ = binary.Write(bw, binary.LittleEndian, uint32(len(idx.blocks)))

	// blocks
	for _, block := range idx.blocks {
		_ = binary.Write(bw, binary.LittleEndian, uint32(len(block)))
		for _, p := range block {
			bw.WriteString(p)
			bw.WriteByte(0)
		}
	}

	// posting lists
	_ = binary.Write(bw, binary.LittleEndian, uint32(len(idx.posting)))
	for tri, list := range idx.posting {
		_ = binary.Write(bw, binary.LittleEndian, tri)
		_ = binary.Write(bw, binary.LittleEndian, uint32(len(list)))
		for _, id := range list {
			_ = binary.Write(bw, binary.LittleEndian, id)
		}
	}

	return bw.Flush()
}

// LoadIndex reads the index from disk.
func LoadIndex(path, root string) (*Index, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r, err := zstd.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	br := bufio.NewReader(r)

	magic := make([]byte, 4)
	if _, err := io.ReadFull(br, magic); err != nil {
		return nil, fmt.Errorf("invalid index")
	}
	if string(magic) != indexMagic {
		return nil, fmt.Errorf("invalid index magic")
	}
	ver, _ := br.ReadByte()
	if ver != indexVersion {
		return nil, fmt.Errorf("unsupported index version %d", ver)
	}

	var numBlocks uint32
	_ = binary.Read(br, binary.LittleEndian, &numBlocks)

	idx := &Index{
		blocks:  make([][]string, numBlocks),
		posting: make(map[uint32][]uint32),
		removed: make(map[string]struct{}),
		meta:    IndexMeta{Root: root},
	}

	for i := range idx.blocks {
		var n uint32
		_ = binary.Read(br, binary.LittleEndian, &n)
		block := make([]string, 0, n)
		for j := uint32(0); j < n; j++ {
			s, _ := br.ReadString(0)
			block = append(block, strings.TrimRight(s, "\x00"))
		}
		idx.blocks[i] = block
	}

	var numTris uint32
	_ = binary.Read(br, binary.LittleEndian, &numTris)
	for i := uint32(0); i < numTris; i++ {
		var tri, n uint32
		_ = binary.Read(br, binary.LittleEndian, &tri)
		_ = binary.Read(br, binary.LittleEndian, &n)
		list := make([]uint32, n)
		for j := range list {
			_ = binary.Read(br, binary.LittleEndian, &list[j])
		}
		idx.posting[tri] = list
	}

	return idx, nil
}
