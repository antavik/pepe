package docker

import (
	"io"
	"net/http"
	"net"
	"regexp"
	"context"
	"strings"
	"time"
	"fmt"
	"encoding/json"
	"bufio"

	log "github.com/go-pkgz/lgr"
)

// source from https://github.com/umputun/reproxy
const apiVer = "v1.22"

type Docker struct {
	client  *dockerClient
	refresh time.Duration
}

type ContainerInfo struct {
	Id     string
	Name   string
	State  string
	Labels map[string]string
	Ts     time.Time
	Ip     string
	Ports  []int
}

func New(host, net string) *Docker {
	return &Docker{
		client:  NewDockerClient(host, net),
		refresh: time.Second * 5,
	}
}

func (d *Docker) Listen(ctx context.Context) chan *ContainerInfo {
	events := make(chan *ContainerInfo)

	go func() {
		if err := d.listen(ctx, events); err != nil {
			log.Printf("[ERROR] unexpected docker client exit, %v", err)
		}
	}()

	return events
}

func (d *Docker) listen(ctx context.Context, events chan *ContainerInfo) error {
	ticker := time.NewTicker(d.refresh)

	defer ticker.Stop()
	defer close(events)

	update := func() {
		containers, err := d.client.ListContainers()
		if err != nil {
			log.Printf("[WARN] failed to fetch containers info, %v", err)
			return
		}

		for _, c := range containers {
			events <- c
		}
	}

	update()
	for {
		select {
		case <-ticker.C:
			update()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (d *Docker) FollowLogs(contId string, stdout, stderr bool) chan string {
	follow := true

	stream, err := d.client.Logs(contId, follow, stdout, stderr)
	if err != nil {
		log.Printf("[ERROR] docker api error, %v", err)
		return nil
	}

	logCh := make(chan string)

	go func() {
		log.Printf("[DEBUG] start following logs of container %s", contId)

		defer stream.Close()
		defer close(logCh)

		s := bufio.NewScanner(NewLogReader(stream))
		for s.Scan() {
			logCh <- s.Text()
		}
		if err := s.Err(); err != nil {
			log.Printf("[ERROR] docker logs streaming, %v", err)
		}

		log.Printf("[DEBUG] stop following logs of container %s", contId)
	}()

	return logCh
}

// low level docker client
type dockerClient struct {
	client  http.Client
	network string
}

func NewDockerClient(host, network string) *dockerClient {
	re := regexp.MustCompile(`^(?:([a-z0-9]+)://)?(.*)$`)
	parts := re.FindStringSubmatch(host)
	proto, addr := parts[1], parts[2]

	log.Printf("[DEBUG] configuring docker client to connect to %s via %s", addr, proto)

	client := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial(proto, addr)
			},
		},
	}

	return &dockerClient{client, network}
}

func (dc *dockerClient) ListContainers() ([]*ContainerInfo, error) {
	url := fmt.Sprintf("http://localhost/%s/containers/json", apiVer)
	resp, err := dc.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to docker socket to fetch containers info: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		e := struct {
			Message string `json:"message"`
		}{}

		if err := json.NewDecoder(resp.Body).Decode(&e); err != nil {
			return nil, fmt.Errorf("failed to parse error from docker daemon: %v", err)
		}

		return nil, fmt.Errorf("unexpected error from docker daemon: %s", e.Message)
	}

	var response []struct {
		Id              string
		Name            string
		State           string
		Labels          map[string]string
		Created         int64
		NetworkSettings struct {
			Networks map[string]struct { IPAddress string }
		}
		Names           []string
		Ports           []struct{ PrivatePort int }
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to parse response from docker daemon: %v", err)
	}

	containers := make([]*ContainerInfo, len(response))

	for i, resp := range response {
		c := ContainerInfo{}

		c.Id = resp.Id
		c.Name = strings.TrimPrefix(resp.Names[0], "/")
		c.State = resp.State
		c.Labels = resp.Labels
		c.Ts = time.Unix(resp.Created, 0)

		for k, v := range resp.NetworkSettings.Networks {
			if dc.network == "" || k == dc.network { // match on network name if defined
				c.Ip = v.IPAddress
				break
			}
		}

		for _, p := range resp.Ports {
			c.Ports = append(c.Ports, p.PrivatePort)
		}

		containers[i] = &c
	}

	return containers, nil
}

func (dc *dockerClient) Logs(contId string, follow, stdout, stderr bool) (io.ReadCloser, error) {
	url := fmt.Sprintf(
		"http://localhost/%s/containers/%s/logs?follow=%t&stdout=%t&stderr=%t&since=%d",
		apiVer, contId, follow, stdout, stderr, time.Now().Unix(),
	)
	resp, err := dc.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to docker socket to follow logs: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		e := struct {
			Message string `json:"message"`
		}{}

		if err := json.NewDecoder(resp.Body).Decode(&e); err != nil {
			return nil, fmt.Errorf("failed to parse error from docker daemon: %v", err)
		}

		return nil, fmt.Errorf("unexpected error from docker daemon: %s", e.Message)
	}

	return resp.Body, nil
}