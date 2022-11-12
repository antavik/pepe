package manager

import (
	"context"
	"fmt"
	"sync"

	"github.com/antibantique/pepe/src/proc"
	"github.com/antibantique/pepe/src/config"
	"github.com/antibantique/pepe/src/source"
	"github.com/antibantique/pepe/src/discovery/docker"
)

type Manager struct {
	Docker     *docker.Docker
	TaskCh     chan *proc.Task
	CommonConf config.C

	wg  sync.WaitGroup
	reg *Registry
}

func New(d *docker.Docker, tCh chan *proc.Task, commonC config.C) *Manager {
	return &Manager{
		Docker:     d,
		TaskCh:     tCh,
		CommonConf: commonC,

		wg:  sync.WaitGroup{},
		reg: NewRegistry(),
	}
}

func (m *Manager) List() []Harvester {
	return m.reg.List()
}

func (m *Manager) Run(ctx context.Context) {
	go m.listenDocker(ctx)
}

func (m *Manager) Stop() {
	for _, harv := range m.reg.List() {
		harv.Cancel()
	}

	m.wg.Wait()
}

func (m *Manager) listenDocker(ctx context.Context) {
	for container := range m.Docker.Listen(ctx) {
		_, exists := m.reg.Get(container.Id)
		if exists {
			continue
		}

		src := source.S{
			Id:     container.Id,
			Ip:     container.Ip,
			Name:   container.Name,
			Config: config.MakeContainerConfig(container.Name, m.CommonConf, container.Labels),
		}

		if src.Config.Stdout || src.Config.Stderr {
			go m.harvest(ctx, src)
		}
	}
}

func (m *Manager) harvest(ctx context.Context, s source.S) {
	m.wg.Add(1)
	defer m.wg.Done()

	cancelCtx, cancel := context.WithCancel(ctx)

	m.reg.Put(s.Id, Harvester{ s, cancel })
	defer func() {
		harv := m.reg.Del(s.Id)
		harv.Cancel()
	}()

	for l := range m.Docker.FollowLogs(cancelCtx, s.Id, s.Config.Stdout, s.Config.Stderr) {
		if l.Err != nil {
			m.TaskCh <- &proc.Task{
				Src:    &s,
				RawLog: map[string]string{
					"pepe": fmt.Sprintf("Error while streaming docker logs: %v", l.Err),
				},
			}
		}

		if s.Config.Re != nil && !s.Config.Re.MatchString(l.Text) {
			continue
		}

		m.TaskCh <- &proc.Task{
			Src:    &s,
			RawLog: map[string]string{"": l.Text},
		}
	}
}