package ping_test

import (
	"testing"

	"github.com/snowaa-desigram/backend/services/go/internal/ping/service"
)

func TestServicePing(t *testing.T) {
	if got, want := service.New().Ping("hi"), "pong: hi"; got != want {
		t.Errorf("Ping = %q, want %q", got, want)
	}
}
