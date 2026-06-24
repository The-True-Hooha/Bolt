package diff

import (
	"fmt"
	"strings"
)

// Unified returns a simple unified diff of two texts.
func Unified(nameA, nameB, textA, textB string) string {
	linesA := splitLines(textA)
	linesB := splitLines(textB)

	n, m := len(linesA), len(linesB)
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if linesA[i] == linesB[j] {
				dp[i][j] = dp[i+1][j+1] + 1
			} else if dp[i+1][j] >= dp[i][j+1] {
				dp[i][j] = dp[i+1][j]
			} else {
				dp[i][j] = dp[i][j+1]
			}
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "--- %s\n+++ %s\n", nameA, nameB)

	i, j := 0, 0
	for i < n || j < m {
		if i < n && j < m && linesA[i] == linesB[j] {
			fmt.Fprintf(&b, "  %s\n", linesA[i])
			i++
			j++
		} else if j < m && (i >= n || dp[i][j+1] >= dp[i+1][j]) {
			fmt.Fprintf(&b, "+ %s\n", linesB[j])
			j++
		} else {
			fmt.Fprintf(&b, "- %s\n", linesA[i])
			i++
		}
	}
	return b.String()
}

func splitLines(s string) []string {
	lines := strings.Split(s, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}
