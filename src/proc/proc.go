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

	go p.run(taskCh)

	return taskCh
}

func (p *Proc) run(taskCh chan *Task) {
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

				if err := p.Client.Send(msg); err != nil {
					log.Printf("[ERROR] error send message, %v", err)
				}
			}(prov)
		}

		wg.Wait()
	}
}