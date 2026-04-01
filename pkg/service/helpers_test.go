package service

import (
	"errors"
	"testing"
	"time"

	"jarvis/pkg/domain"

	"github.com/stretchr/testify/assert"
)

// --------------- cosineSimilarity ---------------

func TestCosineSimilarity_IdenticalVectors_NoTag(t *testing.T) {
	v := []float64{1.0, 0.0, 0.0}

	result := cosineSimilarity(v, v)

	assert.InDelta(t, 1.0, result, 0.0001)
}

func TestCosineSimilarity_OrthogonalVectors_NoTag(t *testing.T) {
	a := []float64{1.0, 0.0, 0.0}
	b := []float64{0.0, 1.0, 0.0}

	result := cosineSimilarity(a, b)

	assert.InDelta(t, 0.0, result, 0.0001)
}

func TestCosineSimilarity_DifferentLengths_NoTag(t *testing.T) {
	a := []float64{1.0, 0.0}
	b := []float64{1.0, 0.0, 0.0}

	result := cosineSimilarity(a, b)

	assert.Equal(t, 0.0, result)
}

func TestCosineSimilarity_EmptyVectors_NoTag(t *testing.T) {
	result := cosineSimilarity([]float64{}, []float64{})

	assert.Equal(t, 0.0, result)
}

func TestCosineSimilarity_ZeroVector_NoTag(t *testing.T) {
	a := []float64{0.0, 0.0, 0.0}
	b := []float64{1.0, 0.0, 0.0}

	result := cosineSimilarity(a, b)

	assert.Equal(t, 0.0, result)
}

func TestCosineSimilarity_OppositeVectors(t *testing.T) {
	a := []float64{1.0, 0.0, 0.0}
	b := []float64{-1.0, 0.0, 0.0}

	result := cosineSimilarity(a, b)

	assert.InDelta(t, -1.0, result, 0.0001)
}

func TestCosineSimilarity_PartialOverlap(t *testing.T) {
	a := []float64{1.0, 1.0, 0.0}
	b := []float64{1.0, 0.0, 0.0}

	result := cosineSimilarity(a, b)

	assert.InDelta(t, 0.7071, result, 0.001)
}

func TestCosineSimilarity_NilVectors(t *testing.T) {
	result := cosineSimilarity(nil, nil)

	assert.Equal(t, 0.0, result)
}

func TestCosineSimilarity_OneNil(t *testing.T) {
	result := cosineSimilarity([]float64{1.0}, nil)

	assert.Equal(t, 0.0, result)
}

func TestCosineSimilarity_LargeVectors(t *testing.T) {
	a := make([]float64, 64)
	b := make([]float64, 64)
	for i := range a {
		a[i] = float64(i) / 64.0
		b[i] = float64(i) / 64.0
	}

	result := cosineSimilarity(a, b)

	assert.InDelta(t, 1.0, result, 0.0001)
}

// --------------- timeDecay ---------------

func TestTimeDecay_RecentTime_NearOne(t *testing.T) {
	result := timeDecay(time.Now())

	assert.InDelta(t, 1.0, result, 0.01)
}

func TestTimeDecay_OldTime_DecaysTowardZero(t *testing.T) {
	old := time.Now().AddDate(0, 0, -100)

	result := timeDecay(old)

	assert.Less(t, result, 0.1)
	assert.Greater(t, result, 0.0)
}

func TestTimeDecay_MonotoneDecreasing(t *testing.T) {
	t1 := time.Now()
	t2 := time.Now().AddDate(0, 0, -7)
	t3 := time.Now().AddDate(0, 0, -30)

	d1 := timeDecay(t1)
	d2 := timeDecay(t2)
	d3 := timeDecay(t3)

	assert.Greater(t, d1, d2)
	assert.Greater(t, d2, d3)
}

func TestTimeDecay_OneDayAgo(t *testing.T) {
	yesterday := time.Now().Add(-24 * time.Hour)

	result := timeDecay(yesterday)

	// e^(-0.05*1) ≈ 0.9512
	assert.InDelta(t, 0.9512, result, 0.01)
}

func TestTimeDecay_OneWeekAgo(t *testing.T) {
	weekAgo := time.Now().Add(-7 * 24 * time.Hour)

	result := timeDecay(weekAgo)

	// e^(-0.05*7) ≈ 0.7047
	assert.InDelta(t, 0.7047, result, 0.01)
}

func TestTimeDecay_AlwaysPositive(t *testing.T) {
	veryOld := time.Now().AddDate(-10, 0, 0)

	result := timeDecay(veryOld)

	assert.Greater(t, result, 0.0)
}

// --------------- mergeSearchResults ---------------

func TestMergeSearchResults_BothSucceed(t *testing.T) {
	vec := []domain.Memory{
		{ID: 1, Content: "vec result", Score: 0.9},
		{ID: 2, Content: "vec only", Score: 0.5},
	}
	fts := []domain.Memory{
		{ID: 1, Content: "fts result", Score: 0.8},
		{ID: 3, Content: "fts only", Score: 0.6},
	}

	results := mergeSearchResults(vec, fts, nil, nil, 0.6, 0.4)

	assert.Len(t, results, 3)

	scoreByID := make(map[int64]float64)
	for _, r := range results {
		scoreByID[r.ID] = r.Score
	}

	assert.InDelta(t, 0.6*0.9+0.4*0.8, scoreByID[1], 0.001)
	assert.InDelta(t, 0.6*0.5, scoreByID[2], 0.001)
	assert.InDelta(t, 0.4*0.6, scoreByID[3], 0.001)
}

func TestMergeSearchResults_VecFails(t *testing.T) {
	fts := []domain.Memory{
		{ID: 1, Content: "fts only", Score: 0.8},
	}

	results := mergeSearchResults(nil, fts, errors.New("vec fail"), nil, 0.6, 0.4)

	assert.Len(t, results, 1)
	assert.InDelta(t, 0.4*0.8, results[0].Score, 0.001)
}

func TestMergeSearchResults_FTSFails(t *testing.T) {
	vec := []domain.Memory{
		{ID: 1, Content: "vec only", Score: 0.9},
	}

	results := mergeSearchResults(vec, nil, nil, errors.New("fts fail"), 0.6, 0.4)

	assert.Len(t, results, 1)
	assert.InDelta(t, 0.6*0.9, results[0].Score, 0.001)
}

func TestMergeSearchResults_BothEmpty(t *testing.T) {
	results := mergeSearchResults(nil, nil, nil, nil, 0.6, 0.4)

	assert.Empty(t, results)
}

func TestMergeSearchResults_BothFail_ReturnsNil(t *testing.T) {
	results := mergeSearchResults(nil, nil, errors.New("vec"), errors.New("fts"), 0.6, 0.4)

	assert.Empty(t, results)
}

func TestMergeSearchResults_Deduplication(t *testing.T) {
	vec := []domain.Memory{
		{ID: 1, Content: "shared", Score: 0.5},
	}
	fts := []domain.Memory{
		{ID: 1, Content: "shared", Score: 0.5},
	}

	results := mergeSearchResults(vec, fts, nil, nil, 0.6, 0.4)

	assert.Len(t, results, 1)
	assert.Equal(t, int64(1), results[0].ID)
}

func TestMergeSearchResults_EqualWeights(t *testing.T) {
	vec := []domain.Memory{
		{ID: 1, Score: 0.8},
	}
	fts := []domain.Memory{
		{ID: 1, Score: 0.6},
	}

	results := mergeSearchResults(vec, fts, nil, nil, 0.5, 0.5)

	assert.Len(t, results, 1)
	assert.InDelta(t, 0.5*0.8+0.5*0.6, results[0].Score, 0.001)
}

func TestMergeSearchResults_MultipleOverlapping(t *testing.T) {
	vec := []domain.Memory{
		{ID: 1, Score: 0.9},
		{ID: 2, Score: 0.7},
		{ID: 3, Score: 0.3},
	}
	fts := []domain.Memory{
		{ID: 2, Score: 0.8},
		{ID: 3, Score: 0.6},
		{ID: 4, Score: 0.4},
	}

	results := mergeSearchResults(vec, fts, nil, nil, 0.6, 0.4)

	assert.Len(t, results, 4)

	scoreByID := make(map[int64]float64)
	for _, r := range results {
		scoreByID[r.ID] = r.Score
	}

	assert.InDelta(t, 0.6*0.9, scoreByID[1], 0.001)
	assert.InDelta(t, 0.6*0.7+0.4*0.8, scoreByID[2], 0.001)
	assert.InDelta(t, 0.6*0.3+0.4*0.6, scoreByID[3], 0.001)
	assert.InDelta(t, 0.4*0.4, scoreByID[4], 0.001)
}

func TestMergeSearchResults_ZeroWeights(t *testing.T) {
	vec := []domain.Memory{{ID: 1, Score: 0.9}}
	fts := []domain.Memory{{ID: 2, Score: 0.8}}

	results := mergeSearchResults(vec, fts, nil, nil, 0.0, 0.0)

	assert.Len(t, results, 2)
	for _, r := range results {
		assert.Equal(t, 0.0, r.Score)
	}
}

// --------------- SheetsFinanceService ---------------

func TestNewSheetsFinanceService_ImplementsFinanceService(t *testing.T) {
	var _ FinanceService = (*SheetsFinanceService)(nil)
}

func TestNewSheetsFinanceService_SetsFields(t *testing.T) {
	svc := NewSheetsFinanceService(nil, "Gastos2024")

	assert.Equal(t, "Gastos2024", svc.sheetName)
	assert.Nil(t, svc.client)
}

func TestNewSheetsFinanceService_DifferentSheetNames(t *testing.T) {
	svc1 := NewSheetsFinanceService(nil, "Sheet1")
	svc2 := NewSheetsFinanceService(nil, "Sheet2")

	assert.Equal(t, "Sheet1", svc1.sheetName)
	assert.Equal(t, "Sheet2", svc2.sheetName)
}
