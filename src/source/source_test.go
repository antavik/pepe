package source

import (
	"testing"
	"fmt"

	"github.com/stretchr/testify/assert"
)

func TestStringer(t *testing.T) {
	tests := []struct{
		src  S
		want string
	}{
		{ S{ Name: "test_name" }, "test_name" },
		{ S{ Id:   "test_id" },   "test_id" },
		{ S{ Ip:   "test_ip" },   "test_ip" },
	}

	for _, test := range tests {
		assert.Equal(t, test.want, fmt.Sprint(test.src))
	}
}

func TestMap(t *testing.T) {
	testSrc := S{
		Name: "test_name",
		Ip:   "test_ip",
		Id:   "test_id",
	}

	m := testSrc.Map()

	assert.Equal(t, testSrc.Name, m["name"])
	assert.Equal(t, testSrc.Ip,   m["ip"])
	assert.Equal(t, testSrc.Id,   m["id"])
}