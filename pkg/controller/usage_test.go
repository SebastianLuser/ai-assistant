package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsageController_New(t *testing.T) {
	ctrl := NewUsageController(nil)

	assert.NotNil(t, ctrl)
}
