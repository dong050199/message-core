package xhttp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPromConfig(t *testing.T) {
	prom := NewBasePromConfig("test", "test")
	assert.Equal(t, "test", prom.Subsystem)
}
