package docker

import (
	"testing"
	"encoding/binary"
	"bytes"
	"io"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newLog(msg string) []byte {
	logLen := make([]byte, 4)
	binary.BigEndian.PutUint32(logLen, uint32(len(msg)))

	log := []byte{0x1, 0x0, 0x0, 0x0,}

	log = append(log, logLen...)
	log = append(log, []byte(msg)...)

	return log
}

func TestRead(t *testing.T) {
	msg := "test"

	testLog, err := io.ReadAll(NewLogReader(bytes.NewReader(newLog(msg))))
	if err != nil {
		require.Error(t, err)
	}

	assert.Equal(t, msg, string(testLog))
}