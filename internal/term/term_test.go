package term

import "testing"

func TestApply_disabled(t *testing.T) {
	orig := colorEnabled
	colorEnabled = false

	defer func() { colorEnabled = orig }()

	if got := Bold("hello"); got != "hello" {
		t.Errorf("Bold() = %q, want %q", got, "hello")
	}

	if got := BoldRed("err"); got != "err" {
		t.Errorf("BoldRed() = %q, want %q", got, "err")
	}
}

func TestApply_enabled(t *testing.T) {
	orig := colorEnabled
	colorEnabled = true

	defer func() { colorEnabled = orig }()

	got := Green("ok")
	want := "\033[32mok\033[0m"

	if got != want {
		t.Errorf("Green() = %q, want %q", got, want)
	}
}

func TestBoldCyan_enabled(t *testing.T) {
	orig := colorEnabled
	colorEnabled = true

	defer func() { colorEnabled = orig }()

	got := BoldCyan("herald")
	want := "\033[1;36mherald\033[0m"

	if got != want {
		t.Errorf("BoldCyan() = %q, want %q", got, want)
	}
}

func TestDim_enabled(t *testing.T) {
	orig := colorEnabled
	colorEnabled = true

	defer func() { colorEnabled = orig }()

	got := Dim("faint")
	want := "\033[2mfaint\033[0m"

	if got != want {
		t.Errorf("Dim() = %q, want %q", got, want)
	}
}
