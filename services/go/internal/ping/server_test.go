package ping

import (
	"context"
	"testing"

	pingv1 "github.com/snowaa-desigram/backend/services/go/gen/desigram/ping/v1"
)

func TestPing(t *testing.T) {
	resp, err := NewServer().Ping(context.Background(), &pingv1.PingRequest{Message: "hi"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := resp.GetMessage(), "pong: hi"; got != want {
		t.Errorf("message = %q, want %q", got, want)
	}
	if got, want := resp.GetService(), "ping"; got != want {
		t.Errorf("service = %q, want %q", got, want)
	}
}
