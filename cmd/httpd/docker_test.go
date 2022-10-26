package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/353solutions/unter/cache"
)

func TestIntegrationCache(t *testing.T) {
	require := require.New(t)
	redisPort := runDocker(t, "redis:7-alpine")
	t.Logf("redis on %s", redisPort)

	addr := fmt.Sprintf("localhost:%s", redisPort)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	cache, err := cache.Connect(ctx, addr, time.Second)
	require.NoError(err, "new cache")
	defer cache.Close()
}

// Homework: Run postgres image. You'll need to add environment variables to runDocker

// runs docker image, returns post on host
func runDocker(t *testing.T, image string, env ...string) string {
	// env is []string
	require := require.New(t)
	cmd := exec.Command("docker", "run", "-P", "-d", image)
	out, err := cmd.CombinedOutput()
	require.NoErrorf(err, "run docker %s", image)
	cid := string(out[:len(out)-1]) // trim \n
	t.Logf("image: %q, container: %q", image, cid)

	t.Cleanup(func() {
		cmd := exec.Command("docker", "rm", "-f", cid)
		if err := cmd.Run(); err != nil {
			t.Logf("warning: can't delete %q - %s", cid, err)
		}
	})

	port, err := exposedPort(cid)
	require.NoError(err, "getting port")
	return port
}

// returns the first exposed port from a container
func exposedPort(cid string) (string, error) {
	cmd := exec.Command("docker", "inspect", "-f", "{{ .NetworkSettings.Ports | json }}", cid)
	data, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	var ports map[string][]map[string]string
	if err := json.Unmarshal(data, &ports); err != nil {
		return "", err
	}

	for _, mappings := range ports {
		for _, addr := range mappings {
			return addr["HostPort"], nil
		}
	}

	return "", fmt.Errorf("no exposed ports")
}
