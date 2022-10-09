package source

import (
	"testing"
	"fmt"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	tests := []struct{
		src  S
		want string
	}{
		{ S{ Name: "test_name", }, "test_name", },
		{ S{ Id:   "test_id", },   "test_id", },
		{ S{ Ip:   "test_ip", },   "test_ip", },
	}

	for _, test := range tests {
		assert.Equal(t, test.want, fmt.Sprint(test.src))
	}
}