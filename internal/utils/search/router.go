package search

import (
	"os"
	"time"
)

const indexMaxAge = 24 * time.Hour

type Options struct {
	Root          string
	Pattern       string
	Mode          MatchMode
	CaseSensitive bool
	Recursive     bool
	MaxDepth      int
	Hidden        bool
	Type          string
	Exts          []string
	Ignore        []string
	Limit         int
	ForceIndex    bool
	ForceWalk     bool
	SearchContent bool
}

func Search(opts Options, out chan<- Result) error {
	m, err := NewMatcher(opts.Pattern, opts.Mode, opts.CaseSensitive)
	if err != nil {
		return err
	}

	useIndex := opts.ForceIndex || shouldUseIndex(opts.Root)
	if opts.ForceWalk {
		useIndex = false
	}

	if useIndex {
		idxPath := IndexPath(opts.Root)
		idx, err := LoadIndex(idxPath, opts.Root)
		if err != nil {
			// fall back to walk
			useIndex = false
		} else {
			idx.Query(opts.Pattern, m, opts.Limit, out)
			return nil
		}
	}

	if !useIndex {
		Walk(WalkOptions{
			Root:          opts.Root,
			Recursive:     opts.Recursive,
			MaxDepth:      opts.MaxDepth,
			Hidden:        opts.Hidden,
			Type:          opts.Type,
			Exts:          opts.Exts,
			Ignore:        opts.Ignore,
			Matcher:       m,
			Limit:         opts.Limit,
			SearchContent: opts.SearchContent,
		}, out)
	}
	return nil
}

func shouldUseIndex(root string) bool {
	idxPath := IndexPath(root)
	info, err := os.Stat(idxPath)
	if err != nil {
		return false
	}
	return time.Since(info.ModTime()) < indexMaxAge
}
