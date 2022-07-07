package discovery

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	manager := NewServiceManager()
	containersCh := make(chan ContainerInfo)
	defer close(containersCh)

	go manager.Run(containersCh)

	containers := []ContainerInfo{
		ContainerInfo{Name: "svc0", Ip: "0.0.0.0", State: "running"},
		ContainerInfo{Name: "svc1", Ip: "0.0.0.1", State: "running"},
	}
	for _, c := range containers {
		containersCh <- c
	}

	time.Sleep(time.Millisecond)

	assert.Equal(t, len(containers), len(manager.mapping))
}

func TestGet(t *testing.T) {
	{
		manager := NewServiceManager()
		containersCh := make(chan ContainerInfo)
		defer close(containersCh)

		go manager.Run(containersCh)

		containers := []ContainerInfo{
			ContainerInfo{Name: "svc0", Ip: "0.0.0.0", State: "running"},
			ContainerInfo{Name: "svc1", Ip: "0.0.0.1", State: "running"},
		}
		for _, c := range containers {
			containersCh <- c

			time.Sleep(time.Millisecond)
			svc, exists := manager.Get(c.Ip)

			assert.True(t, exists, "service should exist")
			assert.NotNil(t, svc)
			assert.Equal(t, c.Id, svc.Id)
			assert.Equal(t, c.Ip, svc.Ip)
			assert.Equal(t, c.Name, svc.Name)
		}
	}
	{
		manager := NewServiceManager()
		containersCh := make(chan ContainerInfo)
		defer close(containersCh)

		go manager.Run(containersCh)

		time.Sleep(time.Millisecond)
		svc, exists := manager.Get("0.0.0.0")

		assert.False(t, exists, "service should not exist")
		assert.NotNil(t, svc)
	}
}

func TestDefaultConfig(t *testing.T) {
	c := DefaultConfig()

	assert.NotNil(t, c)
	assert.True(t, c.TgEnabled)
	assert.True(t, c.SlEnabled)
}

func TestParseConfig(t *testing.T) {
	l := map[string]string{
		"pepe.format":   "format",
		"pepe.telegram": "false",
		"pepe.slack":    "false",
	}

	c := parseConfig(l, "svc")

	assert.NotNil(t, c)
	assert.NotNil(t, c.Template)
	assert.False(t, c.TgEnabled)
	assert.False(t, c.SlEnabled)
}