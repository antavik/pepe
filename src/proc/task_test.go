package proc

import (
	"testing"
	"fmt"

	"github.com/stretchr/testify/assert"

	"github.com/antibantique/pepe/src/source"
)

func TestString(t *testing.T) {
	log := map[string]string{
		"log":     "test",
		"message": "42",
	}

	tasks := []Task{
		Task{ Src: &source.S{ Name: "test_name", }, Log: log, },
		Task{ Src: &source.S{ Name: "test_id", }, Log: log, },
		Task{ Src: &source.S{ Name: "test_ip", }, Log: log, },
	}

	for _, task := range tasks {
		assert.NotEmpty(t, fmt.Sprint(task))
	}
}