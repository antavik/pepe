package docker

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"fmt"
	"io"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListContainers(t *testing.T) {
	{
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/v1.22/containers/json", r.URL.Path)

			http.Error(w, `{"message": "error"}`, http.StatusInternalServerError)
		}))
		defer srv.Close()

		addr := fmt.Sprintf("tcp://%s", strings.TrimPrefix(srv.URL, "http://"))

		client := NewDockerClient(addr, "bridge")
		_, err := client.ListContainers()

		assert.Error(t, err, "demon error response should return error")
	}
	{
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/v1.22/containers/json", r.URL.Path)

			resp, err := os.ReadFile("testdata/containers.json")
			require.NoError(t, err)

			w.Write(resp)
		}))
		defer srv.Close()

		addr := fmt.Sprintf("tcp://%s", strings.TrimPrefix(srv.URL, "http://"))

		client := NewDockerClient(addr, "bridge")
		containers, err := client.ListContainers()

		assert.NoError(t, err, "demon valid response should no error")
		assert.Len(t, containers, 2)
		assert.Equal(t, "4835daea7d471171ab81f4548c1da3de2eb35d4869707b1a1cd91c47cec1993c", containers[0].Id)
		assert.Equal(t, "pepe_serv0_run_c285d257cc02", containers[0].Name)
		assert.Equal(t, "running", containers[0].State)
		assert.Equal(t, "Alert", containers[0].Labels["pepe.format"])
		assert.Equal(t, []int{8080}, containers[0].Ports)
		assert.NotEmpty(t, containers[0].Labels)
	}
}

func TestLogs(t *testing.T) {
	{
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/v1.22/containers/test/logs", r.URL.Path)
			require.True(t, strings.Contains(r.URL.RawQuery, "follow=true&stdout=true&stderr=true&since="))

			http.Error(w, `{"message": "error"}`, http.StatusInternalServerError)
		}))
		defer srv.Close()

		addr := fmt.Sprintf("tcp://%s", strings.TrimPrefix(srv.URL, "http://"))

		client := NewDockerClient(addr, "")
		_, err := client.Logs("test", true, true, true)

		assert.Error(t, err, "demon error response should return error")
	}
	{
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/v1.22/containers/test/logs", r.URL.Path)
			require.True(t, strings.Contains(r.URL.RawQuery, "follow=true&stdout=true&stderr=true&since="))

			resp, err := os.ReadFile("testdata/log")
			require.NoError(t, err)

			w.Write(resp)
		}))
		defer srv.Close()

		addr := fmt.Sprintf("tcp://%s", strings.TrimPrefix(srv.URL, "http://"))

		client := NewDockerClient(addr, "")
		body, err := client.Logs("test", true, true, true)
		defer body.Close()

		data, err := io.ReadAll(body)
		if err != nil {
			require.Error(t, err)
		}

		assert.NoError(t, err, "demon valid response should no error")
		assert.NotEmpty(t, data, "should be not empty body")
	}
}