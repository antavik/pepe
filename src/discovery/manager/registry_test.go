package manager

import (
	"testing"
	"context"

	"github.com/stretchr/testify/assert"

	"github.com/antibantique/pepe/src/source"
	"github.com/antibantique/pepe/src/config"
)

func TestGet(t *testing.T) {
	// get by valid key
	{
		_, cancel := context.WithCancel(context.Background())
		h := Harvester{ source.S{ Ip: "0.0.0.0", Config: &config.C{} }, cancel }

		r := NewRegistry()
		r.Put("key", h)

		testHarv, exists := r.Get("key")

		assert.True(t, exists)
		assert.NotNil(t, testHarv)
	}
	// get by invalid key
	{
		r := NewRegistry()

		testSrc, exists := r.Get("key")

		assert.False(t, exists)
		assert.Empty(t, testSrc, "should return nil")
	}
}

func TestDel(t *testing.T) {
	// del valid key
	{
		_, cancel := context.WithCancel(context.Background())
		h := Harvester{ source.S{ Ip: "0.0.0.0", Config: &config.C{} }, cancel }

		r := NewRegistry()
		r.Put("key", h)

		testHarv := r.Del("key")

		assert.NotNil(t, testHarv)

		_, exists := r.Get("key")

		assert.False(t, exists)
	}
	// del from empty registry
	{
		r := NewRegistry()

		testHarv := r.Del("key")

		assert.Empty(t, testHarv)
	}
}

func TestList(t *testing.T) {
	_, cancel := context.WithCancel(context.Background())
	h := Harvester{ source.S{ Ip: "0.0.0.0", Config: &config.C{} }, cancel }

	r := NewRegistry()
	r.Put("key", h)

	testList := r.List()

	assert.Equal(t, len(testList), 1)
}