package proc

import (
	"sync"

	log "github.com/go-pkgz/lgr"
)

type Proc struct {
	Providers map[string]*Provider
}

func (p *Proc) Run() chan *Task {
	taskCh := make(chan *Task)
	errorCh := make(chan error)

	go p.trackErr(errorCh)
	go p.run(taskCh, errorCh)

	return taskCh
}

func (p *Proc) run(taskCh chan *Task, errorsCh chan error) {
	defer close(errorsCh)

	var wg sync.WaitGroup

	for task := range taskCh {
		msg, err := Format(task)
		if err != nil {
			log.Printf("[WARN] message format error, %v", err)
			continue
		}

		for _, prov := range p.Providers {
			if !prov.Accept(task.Src) {
				continue
			}

			wg.Add(1)

			go func(p *Provider) {
				defer wg.Done()
				errorsCh <- p.Client.Send(msg)
			}(prov)
		}

		wg.Wait()
	}
}

func (p *Proc) trackErr(errorCh chan error) {
	for err := range errorCh {
		if err != nil {
			log.Printf("[ERROR] error send message, %v", err)
		}
	}
}