package manager

import (
	"context"

	"github.com/antibantique/pepe/src/proc"
	"github.com/antibantique/pepe/src/config"
	"github.com/antibantique/pepe/src/source"
	"github.com/antibantique/pepe/src/discovery/docker"
)

type Manager struct {
	Docker     *docker.Docker
	TaskCh     chan *proc.Task
	CommonConf config.C

	harvestersReg *Registry
}

func New(d *docker.Docker, tCh chan *proc.Task, commonC config.C) *Manager {
	return &Manager{
		Docker:     d,
		TaskCh:     tCh,
		CommonConf: commonC,

		harvestersReg: NewRegistry(),
	}
}

func (m *Manager) Run(ctx context.Context) {
	go m.listenDocker(ctx)
}

func (m *Manager) listenDocker(ctx context.Context) {
	for container := range m.Docker.Listen(ctx) {
		_, harvesting := m.harvestersReg.Get(container.Id)
		if harvesting {
			continue
		}

		src := source.S{
			Id:     container.Id,
			Ip:     container.Ip,
			Name:   container.Name,
			Config: config.MakeContainerConfig(container.Name, m.CommonConf, container.Labels),
		}

		if src.Config.Stdout || src.Config.Stderr {
			go m.harvest(src)
		}
	}
}

func (m *Manager) harvest(s source.S) {
	m.harvestersReg.Put(s.Id, &s)
	defer m.harvestersReg.Del(s.Id)

	for log := range m.Docker.FollowLogs(s.Id, s.Config.Stdout, s.Config.Stderr) {
		if s.Config.Re != nil && !s.Config.Re.MatchString(log) {
			continue
		}

		m.TaskCh <- &proc.Task{
			Src:    &s,
			RawLog: map[string]string{"": log},
		}
	}
}

func (m *Manager) List() []*source.S {
	return m.harvestersReg.List()
}