package client

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type errTripper struct{}

// implement http.RoundTripper
func (e errTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, fmt.Errorf("no connection")
}

func TestHealth(t *testing.T) {
	c := New("http://example.com")
	c.client.Transport = errTripper{} // mock

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := c.Health(ctx)
	require.Error(t, err, "health")
}
