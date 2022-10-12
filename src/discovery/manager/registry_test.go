package manager

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/antibantique/pepe/src/source"
	"github.com/antibantique/pepe/src/config"
)

func TestGet(t *testing.T) {
	// get by valid key
	{
		src := &source.S{ Ip: "0.0.0.0", Config: &config.C{} }

		r := NewRegistry()
		r.Put("key", src)

		testSrc, exists := r.Get("key")

		assert.True(t, exists)
		assert.Same(t, src, testSrc, "should return pointer to saved service")
	}
	// get by invalid key
	{
		r := NewRegistry()

		testSrc, exists := r.Get("key")

		assert.False(t, exists)
		assert.Nil(t, testSrc, "should return nil")
	}
}

func TestDel(t *testing.T) {
	// del valid key
	{
		src := &source.S{ Ip: "0.0.0.0", Config: &config.C{} }

		r := NewRegistry()
		r.Put("key", src)

		testSrc := r.Del("key")

		assert.NotNil(t, testSrc, "should return nil")

		_, exists := r.Get("key")

		assert.False(t, exists)
	}
	// del from empty registry
	{
		r := NewRegistry()

		testSrc := r.Del("key")

		assert.Nil(t, testSrc)
	}
}