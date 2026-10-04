package state

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
)

func TestGateArgs(t *testing.T) {
	for shell, want := range map[string][]string{
		"sh":         {"-c", "x"},
		"bash":       {"-c", "x"},
		"pwsh":       {"-NoProfile", "-NonInteractive", "-Command", "x"},
		"powershell": {"-NoProfile", "-NonInteractive", "-Command", "x"},
		"cmd":        {"/C", "x"},
	} {
		if got := GateArgs(shell, "x"); !reflect.DeepEqual(got, want) {
			t.Errorf("%s: got %v want %v", shell, got, want)
		}
	}
}

func TestGateRefusesUnknownShell(t *testing.T) {
	var out bytes.Buffer
	_, _, err := RunGate(t.TempDir(), "python3", "touch ran", &out, &out)
	var u *UnknownShellError
	if !errors.As(err, &u) {
		t.Fatalf("got %v", err)
	}
	if want := "shell python3 is not one of sh, bash, pwsh, powershell, cmd"; err.Error() != want {
		t.Fatalf("got %q want %q", err, want)
	}
}

func TestGateCmdCommandLine(t *testing.T) {
	if got, want := GateCmdLine(`findstr "a b" f`), `/C findstr "a b" f`; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
