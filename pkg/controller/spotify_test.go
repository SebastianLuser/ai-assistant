package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpotifyController_New(t *testing.T) {
	ctrl := NewSpotifyController(nil)

	assert.NotNil(t, ctrl)
}
