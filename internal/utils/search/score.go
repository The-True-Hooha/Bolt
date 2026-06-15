package search

import (
	"os"
	"sort"
	"strings"
	"time"
)

type Result struct {
	Path         string
	Name         string
	Info         os.FileInfo
	Score        int
	IsDir        bool
	ContentMatch bool // true when match was inside file content, not name
}

func rank(results []Result) {
	now := time.Now()
	for i := range results {
		r := &results[i]
		// prefer name-only matches over deep path matches
		if strings.Contains(strings.ToLower(r.Name), strings.ToLower(r.Name)) {
			r.Score += 30
		}
		// recency bonus
		if r.Info != nil && now.Sub(r.Info.ModTime()) < 7*24*time.Hour {
			r.Score += 10
		}
		// shallower path = better
		depth := strings.Count(r.Path, string(os.PathSeparator))
		r.Score -= depth * 2
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
}
