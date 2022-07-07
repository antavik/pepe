package discovery

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"fmt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListContainers(t *testing.T) {
	{
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, `{"message": "error"}`, http.StatusInternalServerError)
		}))
		defer srv.Close()

		addr := fmt.Sprintf("tcp://%s", strings.TrimPrefix(srv.URL, "http://"))

		client := NewDockerClient(addr, "bridge")
		_, err := client.listContainers()

		assert.Error(t, err, "demon error response should return error")
	}
	{ // docker demon success response
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/v1.22/containers/json", r.URL.Path)

			resp, err := os.ReadFile("testdata/containers.json")
			require.NoError(t, err)

			w.Write(resp)
		}))
		defer srv.Close()

		addr := fmt.Sprintf("tcp://%s", strings.TrimPrefix(srv.URL, "http://"))

		client := NewDockerClient(addr, "bridge")
		containers, err := client.listContainers()

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