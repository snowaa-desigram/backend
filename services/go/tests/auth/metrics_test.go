package auth_test

import (
	"errors"
	"testing"

	"github.com/snowaa-desigram/backend/services/go/internal/auth"
)

// Метка result у auth_operations_total — ограниченное множество: ok, коды из common.yaml, internal.
func TestResultLabel(t *testing.T) {
	cases := []struct {
		err  error
		want string
	}{
		{nil, "ok"},
		{auth.ErrInvalidCredentials, "invalid_credentials"},
		{auth.ErrTooManyAttempts, "too_many_attempts"},
		{auth.ValidationError(map[string]string{"email": "bad"}), "validation"},
		{errors.New("db exploded"), "internal"},
	}
	for _, c := range cases {
		if got := auth.ResultLabel(c.err); got != c.want {
			t.Errorf("ResultLabel(%v) = %q, want %q", c.err, got, c.want)
		}
	}
}
