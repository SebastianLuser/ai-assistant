package skills

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDependencyChecker_AllAvailable_True(t *testing.T) {
	dc := NewDependencyChecker(map[string]bool{"calendar": true, "sheets": true})

	assert.True(t, dc.AllAvailable([]string{"calendar", "sheets"}))
}

func TestDependencyChecker_AllAvailable_False(t *testing.T) {
	dc := NewDependencyChecker(map[string]bool{"calendar": true})

	assert.False(t, dc.AllAvailable([]string{"calendar", "sheets"}))
}

func TestDependencyChecker_AllAvailable_EmptyDeps(t *testing.T) {
	dc := NewDependencyChecker(map[string]bool{})

	assert.True(t, dc.AllAvailable(nil))
	assert.True(t, dc.AllAvailable([]string{}))
}

func TestDependencyChecker_Unavailable(t *testing.T) {
	dc := NewDependencyChecker(map[string]bool{"calendar": true})

	missing := dc.Unavailable([]string{"calendar", "sheets", "github"})

	assert.Equal(t, []string{"sheets", "github"}, missing)
}

func TestDependencyChecker_Unavailable_NoneUnavailable(t *testing.T) {
	dc := NewDependencyChecker(map[string]bool{"calendar": true})

	assert.Nil(t, dc.Unavailable([]string{"calendar"}))
}
