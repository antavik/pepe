package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetupProvider_emptyTokens(t *testing.T) {
	opts.Tg.Token = ""
	opts.Sl.Token = ""

	_, err := setupProviders()

	assert.Error(t, err, "should be error because no provider")
}

func TestParseSize(t *testing.T) {
	// empty string input
	{
		size, err := parseSize("")

		assert.Error(t, err)
		assert.Equal(t, uint64(0), size)
	}
	// valid inputs
	{
		tests := []struct{
			size string
			want uint64
		}{
			{"1k", uint64(1024)},
			{"2k", uint64(2048)},
			{"1M", uint64(1048576)},
			{"1G", uint64(1073741824)},
			{"10", uint64(10)},
		}

		for _, tt := range tests {
			size, err := parseSize(tt.size)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, size)
		}
	}
}