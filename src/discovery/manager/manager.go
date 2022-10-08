package manager

import (
	"context"

	"github.com/antibantique/pepe/src/proc"
	"github.com/antibantique/pepe/src/config"
	"github.com/antibantique/pepe/src/source"
	"github.com/antibantique/pepe/src/discovery"
)

type Manager struct {
	Docker     *discovery.Docker
	TaskCh     chan *proc.Task
	CommonConf config.C

	registry   *Registry
}

func New(d *discovery.Docker, tCh chan *proc.Task, commonC config.C) *Manager {
	return &Manager{
		Docker:     d,
		TaskCh:     tCh,
		CommonConf: commonC,

		registry:   NewRegistry(),
	}
}

func (m *Manager) Run(ctx context.Context) {
	go m.listenDocker(ctx)
}

func (m *Manager) listenDocker(ctx context.Context) {
	for container := range m.Docker.Listen(ctx) {
		if container.State != "running" {
			m.registry.Del(container.Id)
			continue
		}

		if container.Id == "" {
			continue
		}

		src, exists := m.registry.Get(container.Id)
		if exists {
			continue
		}

		src = &source.S{
			Id:     container.Id,
			Ip:     container.Ip,
			Name:   container.Name,
			Config: config.MakeContainerConfig(container.Name, m.CommonConf, container.Labels),
		}

		m.registry.Put(src.Id, src)

		if src.Config.Stdout || src.Config.Stderr {
			go m.harvest(src)
		}
	}
}

func (m *Manager) harvest(s *source.S) {
	for log := range m.Docker.FollowLogs(s.Id, s.Config.Stdout, s.Config.Stderr) {
		if s.Config.Re != nil && !s.Config.Re.MatchString(log) {
			continue
		}

		m.TaskCh <- &proc.Task{
			Src: s,
			Log: map[string]string{"Log": log},
		}
	}
}