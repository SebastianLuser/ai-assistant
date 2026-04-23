package service

import (
	"math"
	"time"

	"jarvis/pkg/domain"
)

func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func timeDecay(created time.Time) float64 {
	hours := time.Since(created).Hours()
	if hours < 0 {
		hours = 0
	}
	return 1.0 / (1.0 + math.Log1p(hours/24.0))
}

func mergeSearchResults(vecResults, ftsResults []domain.Memory, vecErr, ftsErr error, vecWeight, ftsWeight float64) []domain.Memory {
	seen := make(map[int64]int)
	var merged []domain.Memory

	if vecErr == nil {
		for _, m := range vecResults {
			m.Score *= vecWeight
			seen[m.ID] = len(merged)
			merged = append(merged, m)
		}
	}

	if ftsErr == nil {
		for _, m := range ftsResults {
			if idx, ok := seen[m.ID]; ok {
				merged[idx].Score += m.Score * ftsWeight
			} else {
				m.Score *= ftsWeight
				merged = append(merged, m)
			}
		}
	}

	return merged
}
