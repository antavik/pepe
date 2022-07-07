package discovery

import (
	"sync"
	"strings"
	"text/template"
	"fmt"
	"strconv"

	log "github.com/go-pkgz/lgr"
)

type ServiceManger struct {
	mapping map[string]*Service
	mu      sync.RWMutex
}

type Service struct {
	Id     string
	Ip     string
	Name   string
	Config *Config
}

type Config struct {
	Template  *template.Template
	TgEnabled bool
	SlEnabled bool
}

func DefaultConfig() *Config {
	conf := Config{}

	conf.TgEnabled = true
	conf.SlEnabled = true

	return &conf
}

func NewServiceManager() *ServiceManger {
	return &ServiceManger{
		mapping: make(map[string]*Service),
	}
}

func (sm *ServiceManger) Run(containers chan ContainerInfo) {
	for c := range containers {

		if c.State != "running" {
			sm.mu.Lock()
			delete(sm.mapping, c.Ip)
			sm.mu.Unlock()
			continue
		}

		if c.Ip == "" {
			continue
		}

		config := parseConfig(c.Labels, c.Name)

		sm.mu.Lock()
		sm.mapping[c.Ip] = &Service{
			Id:     c.Id,
			Ip:     c.Ip,
			Name:   c.Name,
			Config: config,
		}
		sm.mu.Unlock()
	}
}

func (sm *ServiceManger) Get(ip string) (svc *Service, exists bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	svc, exists = sm.mapping[ip]
	if !exists {
		svc = &Service{
			Ip:     ip,
			Config: DefaultConfig(),
		}
	}

	return svc, exists
}

func parseConfig(labels map[string]string, svc string) *Config{
	c := DefaultConfig()

	for k, v := range labels {
		if strings.HasPrefix(k, "pepe") {
			parts := strings.Split(k, `.`)
			conf := parts[1]

			switch conf {
			case "format":
				name := fmt.Sprintf("%s_template", svc)
				template, err := template.New(name).Parse(v)
				if err != nil {
					log.Printf("[WARNING] template parse error for %s, %v", svc, err)
					continue
				}
				c.Template = template
			case "telegram":
				enabled, err := strconv.ParseBool(v)
				if err != nil {
					log.Printf("[WARNING] telegram enabled parse error for %s, %v", svc, err)
					continue
				}
				c.TgEnabled = enabled
			case "slack":
				enabled, err := strconv.ParseBool(v)
				if err != nil {
					log.Printf("[WARNING] slack enabled parse error for %s, %v", svc, err)
					continue
				}
				c.SlEnabled = enabled
			}
		}
	}

	return c
}