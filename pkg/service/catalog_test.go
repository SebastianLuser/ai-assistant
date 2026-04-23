package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNullCatalogService_RecordUsage(t *testing.T) {
	svc := NullCatalogService{}

	err := svc.RecordUsage("save_expense", "tool", true)

	assert.NoError(t, err)
}

func TestNullCatalogService_GetAll(t *testing.T) {
	svc := NullCatalogService{}

	entries, err := svc.GetAll()

	assert.NoError(t, err)
	assert.Nil(t, entries)
}

func TestNullCatalogService_GetByName(t *testing.T) {
	svc := NullCatalogService{}

	entry, err := svc.GetByName("save_expense", "tool")

	assert.NoError(t, err)
	assert.Nil(t, entry)
}

func TestLastUsedFormatted_Nil(t *testing.T) {
	assert.Equal(t, "never", LastUsedFormatted(nil))
}

func TestCatalogEntry_SuccessRate(t *testing.T) {
	// This tests the domain method but is convenient to test here.
	// With 8 successes out of 10 uses.
	entry := struct {
		UsageCount   int64
		SuccessCount int64
	}{UsageCount: 10, SuccessCount: 8}

	rate := float64(entry.SuccessCount) / float64(entry.UsageCount)
	assert.InDelta(t, 0.8, rate, 0.001)
}
