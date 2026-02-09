package errors

import (
	"fmt"
	"testing"
)

func TestRuntime(t *testing.T) {
	cause := fmt.Errorf("disk full")
	err := Runtime("write failed", cause)

	if err.Message != "write failed" {
		t.Errorf("Message = %q, want %q", err.Message, "write failed")
	}

	if err.ExitCode != ExitRuntime {
		t.Errorf("ExitCode = %d, want %d", err.ExitCode, ExitRuntime)
	}

	if err.Cause != cause {
		t.Errorf("Cause = %v, want %v", err.Cause, cause)
	}
}

func TestConfig(t *testing.T) {
	err := Config("bad flag")

	if err.Message != "bad flag" {
		t.Errorf("Message = %q, want %q", err.Message, "bad flag")
	}

	if err.ExitCode != ExitConfig {
		t.Errorf("ExitCode = %d, want %d", err.ExitCode, ExitConfig)
	}

	if err.Cause != nil {
		t.Errorf("Cause = %v, want nil", err.Cause)
	}
}

func TestEnvironment(t *testing.T) {
	cause := fmt.Errorf("not found")
	err := Environment("git missing", cause)

	if err.ExitCode != ExitEnvironment {
		t.Errorf("ExitCode = %d, want %d", err.ExitCode, ExitEnvironment)
	}
}

func TestUserAbort(t *testing.T) {
	err := UserAbort()

	if err.ExitCode != ExitUserAbort {
		t.Errorf("ExitCode = %d, want %d", err.ExitCode, ExitUserAbort)
	}

	if err.Message != "operation cancelled by user" {
		t.Errorf("Message = %q, want %q", err.Message, "operation cancelled by user")
	}
}

func TestError_with_cause(t *testing.T) {
	cause := fmt.Errorf("underlying")
	err := Runtime("top-level", cause)

	want := "top-level: underlying"
	if got := err.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestError_without_cause(t *testing.T) {
	err := Config("bad input")

	if got := err.Error(); got != "bad input" {
		t.Errorf("Error() = %q, want %q", got, "bad input")
	}
}

func TestUnwrap(t *testing.T) {
	cause := fmt.Errorf("root cause")
	err := Runtime("wrapper", cause)

	if got := err.Unwrap(); got != cause {
		t.Errorf("Unwrap() = %v, want %v", got, cause)
	}
}

func TestUnwrap_nil(t *testing.T) {
	err := Config("no cause")

	if got := err.Unwrap(); got != nil {
		t.Errorf("Unwrap() = %v, want nil", got)
	}
}
