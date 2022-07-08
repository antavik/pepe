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

	log "github.com/go-pkgz/lgr"
	"github.com/jessevdk/go-flags"

	"github.com/antibantique/pepe/src/api"
	"github.com/antibantique/pepe/src/proc"
	"github.com/antibantique/pepe/src/providers"
	"github.com/antibantique/pepe/src/discovery"
)

var opts struct {
	Port             int    `short:"p" long:"port" default:"9393" description:"port to listen"`
	MaxSize          string `long:"max" env:"MAX_SIZE" default:"128K" description:"max request size"`
	StdOutLogEnbaled bool   `long:"stdout" env:"STDOUT" description:"stdout log"`
	AllowAll         bool    `long:"all" env:"ALLOWALL" description:"allow all logs"`

	Docker struct {
		Host    string `long:"host" env:"HOST" default:"unix:///var/run/docker.sock" description:"docker host"`
		Network string `long:"network" env:"NETWORK" default:"" description:"docker network"`
	} `group:"docker" namespace:"docker" env-namespace:"DOCKER"`

	Tg struct {
		Token     string        `long:"token" env:"TOKEN" default:"" description:"telegram token"`
		Server    string        `long:"server" env:"SERVER" default:"https://api.telegram.org" description:"telebram bot api server"`
		Timeout   time.Duration `long:"timeout" env:"TIMEOUT" default:"1m" description:"telegram timeout"`
		ChatId string           `long:"chat" env:"CHAT" description:"telegram chat id"`
	} `group:"telegram" namespace:"telegram" env-namespace:"TELEGRAM"`

	Sl struct {
		Token     string        `long:"token" env:"TOKEN" default:"" description:"slack token"`
		Server    string        `long:"server" env:"SERVER" default:"https://slack.com/api/" description:"slack bot api server"`
		Timeout   time.Duration `long:"timeout" env:"TIMEOUT" default:"1m" description:"slack timeout"`
		ChatId string           `long:"chat" env:"CHAT" description:"slack chat id"`
	} `group:"slack" namespace:"slack" env-namespace:"SLACK"`

	Dbg bool `long:"debug" env:"DEBUG" description:"debug mode"`
}

var version = "development"

func main() {
	fmt.Printf("pepe %s\n", version)

	p := flags.NewParser(&opts, flags.PrintErrors|flags.HelpFlag)
	if _, err := p.Parse(); err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	setupLog()

	maxSize, err := parseSize(opts.MaxSize)
	if err != nil {
		log.Printf("[FATAL] invalid max body size value, %v", err)
	}

	provs, err := setupProviders()
	if err != nil {
		log.Printf("[FATAL] provider configuration error, %v", err)
	}

	// setup docker
	docker := discovery.Docker{
		DockerClient:    discovery.NewDockerClient(opts.Docker.Host, opts.Docker.Network),
		RefreshInterval: time.Second * 10,
	}
	containersCh := docker.Listen(context.Background())

	// run manager to process running containers to services
	serviceMan := discovery.NewServiceManager()
	go serviceMan.Run(containersCh)

	// run task processor
	processor := proc.Processor{
		Services:  serviceMan,
		Providers: provs,
		AllowAll:  opts.AllowAll,
	}
	tasksCh := processor.Run()

	// setup and run api server
	server := api.Server{
		Port:             opts.Port,
		MaxBodySize:      int64(maxSize),
		StdOutLogEnbaled: opts.StdOutLogEnbaled,
		Version:          version,
		TasksCh:          tasksCh,
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

func setupProviders() (ps []*proc.Provider, err error) {
	// setup telegram client
	if opts.Tg.Token != "" {
		tgClient, err := providers.NewTelegramClient(context.Background(), opts.Tg.Token, opts.Tg.Server, opts.Tg.ChatId, opts.Tg.Timeout)
		if err != nil {
			log.Printf("[ERROR] setup telegram client error, %v", err)
		} else {
			ps = append(
				ps,
				&proc.Provider{
					Client: tgClient,
					Accept: func(s *discovery.Service) bool { return s.Config.TgEnabled },
				},
			)
		}
	}

	// setup slack client
	if opts.Sl.Token != "" {
		ps = append(
			ps,
			&proc.Provider{
				Client: providers.NewSlackClient(opts.Sl.Token, opts.Sl.Server, opts.Sl.ChatId, opts.Sl.Timeout),
				Accept: func(s *discovery.Service) bool { return s.Config.SlEnabled },
			},
		)
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
				return 0, fmt.Errorf("parse error %s: %v", size, err)
			}
			return uint64(float64(val) * math.Pow(float64(1024), float64(i+1))), nil
		}
	}

	return strconv.ParseUint(size, 10, 64)
}