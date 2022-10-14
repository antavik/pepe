package main

import (
	"os"
	"fmt"
	"context"
	"time"
	"errors"
	"strings"
	"strconv"
	"math"
	"text/template"
	"regexp"

	log "github.com/go-pkgz/lgr"
	"github.com/jessevdk/go-flags"

	"github.com/antibantique/pepe/src/api"
	"github.com/antibantique/pepe/src/config"
	"github.com/antibantique/pepe/src/proc"
	"github.com/antibantique/pepe/src/providers"
	"github.com/antibantique/pepe/src/discovery/docker"
	"github.com/antibantique/pepe/src/discovery/manager"
	"github.com/antibantique/pepe/src/source"
)

var opts struct {
	Template string `long:"template" env:"MSG_TEMPLATE" default:"🔴 Alert\n{{ . }}" description:"provider message template"`
	Regex    string `long:"regex" env:"REGEX" default:"ERROR|CRITICAL|FATAL" description:"log regex pattern"`
	Dbg bool `long:"debug" env:"DEBUG" description:"debug mode"`

	Api struct {
		Port             int    `short:"p" long:"port" default:"9393" description:"api port to listen"`
		MaxSize          string `long:"max" env:"MAX_SIZE" default:"128K" description:"api max request size"`
		StdOutLogEnbaled bool   `long:"stdout" env:"STDOUT_LOG" description:"api stdout log"`
	} `group:"api" namespace:"api" env-namespace:"API"`

	Docker struct {
		Host    string `long:"host" env:"HOST" default:"unix:///var/run/docker.sock" description:"docker host"`
		Network string `long:"network" env:"NETWORK" default:"" description:"docker network"`
	} `group:"docker" namespace:"docker" env-namespace:"DOCKER"`

	Tg struct {
		Token   string        `long:"token" env:"TOKEN" default:"" description:"telegram token"`
		Server  string        `long:"server" env:"SERVER" default:"https://api.telegram.org" description:"telebram bot api server"`
		Timeout time.Duration `long:"timeout" env:"TIMEOUT" default:"1m" description:"telegram timeout"`
		ChatId  string        `long:"chat" env:"CHAT" description:"telegram chat id"`
	} `group:"telegram" namespace:"telegram" env-namespace:"TELEGRAM"`

	Sl struct {
		Token   string        `long:"token" env:"TOKEN" default:"" description:"slack token"`
		Server  string        `long:"server" env:"SERVER" default:"https://slack.com/api/" description:"slack bot api server"`
		Timeout time.Duration `long:"timeout" env:"TIMEOUT" default:"1m" description:"slack timeout"`
		ChatId  string        `long:"chat" env:"CHAT" description:"slack chat id"`
	} `group:"slack" namespace:"slack" env-namespace:"SLACK"`
}

var version = "development"

func main() {
	fmt.Printf("pepe %s\n", version)

	p := flags.NewParser(&opts, flags.PrintErrors|flags.HelpFlag)
	if _, err := p.Parse(); err != nil {
		os.Exit(2)
	}

	setupLog()

	maxSize, err := parseSize(opts.Api.MaxSize)
	if err != nil {
		log.Printf("[FATAL] invalid max body size value: %v", err)
	}

	template, err := template.New("default_template").Parse(opts.Template)
	if err != nil {
		log.Printf("[FATAL] invalid message template: %v", err)
	}

	regex, err := regexp.Compile(opts.Regex)
	if err != nil {
		log.Printf("[FATAL] invalid regex pattern: %v", err)
	}

	provs, err := setupProviders()
	if err != nil {
		log.Printf("[FATAL] provider configuration error: %v", err)
	}

	config := config.C{
		Template:  template,
		Re:        regex,
		TgEnabled: func() bool {
			_, ok := provs["telegram"]
			return ok
		}(),
		SlEnabled: func() bool {
			_, ok := provs["slack"]
			return ok
		}(),
	}

	processor := proc.Proc{ Providers: provs }
	taskCh := processor.Run()

	docker := docker.New(opts.Docker.Host, opts.Docker.Network)

	srcMan := manager.New(docker, taskCh, config)
	srcMan.Run(context.Background())

	// setup and run api server
	server := api.Server{
		Port:             opts.Api.Port,
		MaxBodySize:      int64(maxSize),
		StdOutLogEnbaled: opts.Api.StdOutLogEnbaled,
		Version:          version,
		TaskCh:           taskCh,
		CommonConf:       config,
		SrcManager:       srcMan,
	}
	server.Run(context.Background())
}

func setupLog() {
	if opts.Dbg {
		log.Setup(log.Debug, log.CallerFile, log.Msec, log.LevelBraces)
		return
	}
	log.Setup(log.Msec, log.LevelBraces)
}

func setupProviders() (map[string]*proc.Provider, error) {
	ps := make(map[string]*proc.Provider)

	// setup telegram client
	if opts.Tg.Token != "" {
		tgProvider, err := providers.NewTelegramProvider(context.Background(), opts.Tg.Token, opts.Tg.Server, opts.Tg.ChatId, opts.Tg.Timeout)
		if err != nil {
			log.Printf("[ERROR] setup telegram client error: %v", err)
		} else {
			ps["telegram"] = &proc.Provider{
				Client: tgProvider,
				Accept: func(s *source.S) bool { return s.Config.TgEnabled },
			}
		}
	}

	// setup slack client
	if opts.Sl.Token != "" {
		ps["slack"] = &proc.Provider{
			Client: providers.NewSlackProvider(opts.Sl.Token, opts.Sl.Server, opts.Sl.ChatId, opts.Sl.Timeout),
			Accept: func(s *source.S) bool { return s.Config.SlEnabled },
		}
	}

	if len(ps) == 0 {
		return ps, fmt.Errorf("at least one provider required")
	}

	return ps, nil
}

func parseSize(size string) (uint64, error) {
	if size == "" {
		return 0, errors.New("empty value")
	}

	size = strings.ToLower(size)

	for i, sfx := range []string{"k", "m", "g"} {
		if strings.HasSuffix(size, sfx) {
			val, err := strconv.Atoi(size[:len(size)-1])
			if err != nil {
				return 0, fmt.Errorf("parse error %s: %w", size, err)
			}
			return uint64(float64(val) * math.Pow(float64(1024), float64(i+1))), nil
		}
	}

	return strconv.ParseUint(size, 10, 64)
}