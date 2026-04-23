package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCatalogEntry_SuccessRate(t *testing.T) {
	e := CatalogEntry{UsageCount: 10, SuccessCount: 8}
	assert.InDelta(t, 0.8, e.SuccessRate(), 0.001)
}

func TestCatalogEntry_SuccessRate_Zero(t *testing.T) {
	e := CatalogEntry{UsageCount: 0}
	assert.Equal(t, float64(0), e.SuccessRate())
}

func TestCatalogEntry_SuccessRate_AllFail(t *testing.T) {
	e := CatalogEntry{UsageCount: 5, SuccessCount: 0, ErrorCount: 5}
	assert.Equal(t, float64(0), e.SuccessRate())
}

func TestCatalogEntry_SuccessRate_AllSuccess(t *testing.T) {
	e := CatalogEntry{UsageCount: 100, SuccessCount: 100}
	assert.Equal(t, float64(1), e.SuccessRate())
}
