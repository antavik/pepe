package proc

import (
	"strings"
	"sync"

	log "github.com/go-pkgz/lgr"

	"github.com/antibantique/pepe/src/discovery"
)

type provider interface {
	Send(string) error
}

type Task struct {
	RemoteAddr string
	LogData    map[string][]string
}

type Provider struct {
	Client   provider
	Supports func(*discovery.Service) bool
}

type Processor struct {
	Services  *discovery.ServiceManger
	Providers []*Provider
	AllowAll  bool
}

func (p *Processor) Run() chan *Task {
	tasksCh := make(chan *Task)
	errorsCh := make(chan error)

	go p.run(tasksCh, errorsCh)
	go p.trackErr(errorsCh)

	return tasksCh
}

func (p *Processor) run(tasksCh chan *Task, errorsCh chan error) {
	defer close(errorsCh)

	var wg sync.WaitGroup

	for task := range tasksCh {
		parts := strings.Split(task.RemoteAddr, `:`)
		ip := parts[0]
		svc, exists := p.Services.Get(ip)
		if !exists && !p.AllowAll {
			log.Printf("[WARN] service not recognized by ip %s", ip)
			continue
		}

		msg, err := format(svc, task.LogData)
		if err != nil {
			log.Printf("[WARN] message format error, %v", err)
			continue
		}

		for _, prov := range p.Providers {
			if prov.Supports(svc) {
				wg.Add(1)
				go func(p *Provider) {
					defer wg.Done()
					errorsCh <- p.Client.Send(*msg)
				}(prov)
			}
		}

		wg.Wait()
	}
}

func (p *Processor) trackErr(errorsCh chan error) {
	for err := range errorsCh {
		if err != nil {
			log.Printf("[ERROR] error send message, %v", err)
		}
	}
}