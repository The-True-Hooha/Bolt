package search

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
)

type WalkOptions struct {
	Root          string
	Recursive     bool
	MaxDepth      int // 0 = unlimited
	Hidden        bool
	Type          string // "f", "d", "l", ""
	Exts          []string
	Ignore        []string
	Matcher       *Matcher
	Limit         int
	SearchContent bool // also search inside text files
}

const contentSizeLimit = 512 * 1024 // 512KB max for content search

func Walk(opts WalkOptions, out chan<- Result) {
	var emitted atomic.Int64

	emit := func(r Result) bool {
		if opts.Limit > 0 && emitted.Add(1) > int64(opts.Limit) {
			return false
		}
		out <- r
		return true
	}

	// bounded work queue — avoids semaphore deadlock
	type job struct {
		dir   string
		depth int
	}

	workers := runtime.NumCPU() * 2
	jobs := make(chan job, 4096)
	var wg sync.WaitGroup

	process := func(j job) {
		defer wg.Done()

		if opts.MaxDepth > 0 && j.depth > opts.MaxDepth {
			return
		}
		entries, err := os.ReadDir(j.dir)
		if err != nil {
			return
		}

		for _, e := range entries {
			name := e.Name()
			if !opts.Hidden && strings.HasPrefix(name, ".") {
				continue
			}
			fullPath := filepath.Join(j.dir, name)

			// ignore patterns
			ignored := false
			for _, pat := range opts.Ignore {
				if ok, _ := filepath.Match(pat, name); ok {
					ignored = true
					break
				}
			}
			if ignored {
				continue
			}

			isDir := e.IsDir()

			// recurse into subdirs
			if isDir && opts.Recursive {
				wg.Add(1)
				jobs <- job{dir: fullPath, depth: j.depth + 1}
			}

			// type filter
			switch opts.Type {
			case "f":
				if isDir {
					continue
				}
			case "d":
				if !isDir {
					continue
				}
			}

			// extension filter (files only)
			if len(opts.Exts) > 0 && !isDir {
				ext := strings.ToLower(filepath.Ext(name))
				if !slices.Contains(opts.Exts, ext) {
					continue
				}
			}

			// match on name
			nameOk, score := opts.Matcher.Match(name)
			if nameOk {
				info, _ := e.Info()
				if !emit(Result{Path: fullPath, Name: name, Info: info, Score: score + 30, IsDir: isDir}) {
					return
				}
				continue
			}

			// content search for non-dir text files
			if opts.SearchContent && !isDir {
				if hit, score := contentMatch(fullPath, opts.Matcher); hit {
					info, _ := e.Info()
					emit(Result{Path: fullPath, Name: name, Info: info, Score: score, IsDir: false, ContentMatch: true})
				}
			}
		}
	}

	// start workers
	for range workers {
		go func() {
			for j := range jobs {
				process(j)
			}
		}()
	}

	wg.Add(1)
	jobs <- job{dir: opts.Root, depth: 0}
	wg.Wait()
	close(jobs)
}

// contentMatch reads the file and checks if the matcher hits anywhere in the content.
func contentMatch(path string, m *Matcher) (bool, int) {
	info, err := os.Stat(path)
	if err != nil || info.Size() == 0 || info.Size() > contentSizeLimit {
		return false, 0
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return false, 0
	}
	// skip binary files
	if bytes.IndexByte(data, 0) >= 0 {
		return false, 0
	}
	// search line by line so we can report score
	for line := range strings.SplitSeq(string(data), "\n") {
		if ok, score := m.Match(line); ok {
			return true, score
		}
	}
	return false, 0
}
