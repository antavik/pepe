package discovery

import (
	"net/http"
	"net"
	"regexp"
	"context"
	"strings"
	"time"
	"fmt"
	"encoding/json"

	log "github.com/go-pkgz/lgr"
)

// source from https://github.com/umputun/reproxy

type Docker struct {
	DockerClient    DockerClient
	RefreshInterval time.Duration
}

type DockerClient interface {
	listContainers() ([]ContainerInfo, error)
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

func (d *Docker) Listen(ctx context.Context) (containersCh chan ContainerInfo) {
	containersCh = make(chan ContainerInfo)

	go func() {
		if err := d.listen(ctx, containersCh); err != nil {
			log.Printf("[ERROR] unexpected docker client exit, %v", err)
		}
	}()

	return containersCh
}

func (d *Docker) listen(ctx context.Context, containersCh chan ContainerInfo) error {
	ticker := time.NewTicker(d.RefreshInterval)
	defer ticker.Stop()
	defer close(containersCh)

	saved := make(map[string]ContainerInfo)

	update := func() {
		containers, err := d.DockerClient.listContainers()
		if err != nil {
			log.Printf("[WARN] failed to list containers, %v", err)
			return
		}

		refresh := false

		for _, c := range containers {
			old, exists := saved[c.Id]

			if !exists || c.Ip != old.Ip || c.State != old.State || !c.Ts.Equal(old.Ts) {
				refresh = true
				break
			}
		}

		if refresh {
			log.Printf("[INFO] found changes in running containers")

			for k := range saved {
				delete(saved, k)
			}
			for _, c := range containers {
				saved[c.Id] = c
				containersCh <- c
			}
		}
	}

	update()
	for {
		select {
		case <- ticker.C:
			update()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

type dockerClient struct {
	client  http.Client
	network string
}

func NewDockerClient(host, network string) *dockerClient {
	schemaRegex := regexp.MustCompile("^(?:([a-z0-9]+)://)?(.*)$")
	parts := schemaRegex.FindStringSubmatch(host)
	proto, addr := parts[1], parts[2]

	log.Printf("[DEBUG] configuring docker client to connect to %s via %s", addr, proto)

	client := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial(proto, addr)
			},
		},
		Timeout: time.Second * 5,
	}

	return &dockerClient{client, network}
}

func (dc *dockerClient) listContainers() ([]ContainerInfo, error) {
	// Minimum API version that returns attached networks
	// docs.docker.com/engine/api/version-history/#v122-api-changes
	const APIVer = "v1.22"

	resp, err := dc.client.Get(fmt.Sprintf("http://localhost/%s/containers/json", APIVer))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to docker socket, %v", err)
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

	containers := make([]ContainerInfo, len(response))

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

		containers[i] = c
	}

	return containers, nil
}