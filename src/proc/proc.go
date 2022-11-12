package proc

import (
	"sync"
	"context"

	log "github.com/go-pkgz/lgr"

	"github.com/antibantique/pepe/src/providers"
)

type Proc struct {
	Provs map[string]providers.P
	F     func(*Task) (string, error)
}

func NewProc(provs map[string]providers.P) *Proc {
	return &Proc{ provs, Format }
}

func (p *Proc) Run(ctx context.Context) chan *Task {
	taskCh := make(chan *Task)

	go p.run(ctx, taskCh)

	return taskCh
}

func (p *Proc) run(ctx context.Context, taskCh chan *Task) {
	var wg sync.WaitGroup

	for {
		select {
		case task := <- taskCh:
			msg, err := p.F(task)
			if err != nil {
				log.Printf("[WARN] message format error, %v", err)
				continue
			}

			for _, prov := range p.Provs {
				if !prov.Accepted(task.Src) {
					continue
				}

				wg.Add(1)

				go func(pr providers.P) {
					defer wg.Done()

					if err := pr.Send(msg); err != nil {
						log.Printf("[ERROR] error send message, %v", err)
					}
				}(prov)
			}

			wg.Wait()
		case <- ctx.Done():
			return
		}
	}
}