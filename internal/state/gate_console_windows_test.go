//go:build windows

package state

import (
	"bytes"
	"os/exec"
	"testing"

	"golang.org/x/sys/windows"
)

var (
	kernel32        = windows.NewLazySystemDLL("kernel32.dll")
	getConsoleCP    = kernel32.NewProc("GetConsoleCP")
	getConsoleOutCP = kernel32.NewProc("GetConsoleOutputCP")
	setConsoleCP    = kernel32.NewProc("SetConsoleCP")
	setConsoleOutCP = kernel32.NewProc("SetConsoleOutputCP")
)

// A pwsh gate sets UTF-8 for itself. It must not leave the console vloop runs
// in on another code page: the operator's terminal keeps its own.
func TestPowerShellGateLeavesTheConsoleCodePages(t *testing.T) {
	in, _, _ := getConsoleCP.Call()
	out, _, _ := getConsoleOutCP.Call()
	if in == 0 || out == 0 {
		t.Skip("no console attached")
	}
	defer setConsoleCP.Call(in)
	defer setConsoleOutCP.Call(out)
	setConsoleCP.Call(437)
	setConsoleOutCP.Call(437)
	for _, shell := range []string{"pwsh", "powershell"} {
		if _, err := exec.LookPath(shell); err != nil {
			continue
		}
		var buf bytes.Buffer
		code, _, err := RunGate(t.TempDir(), shell, "Write-Output 'café'", &buf, &buf)
		if err != nil || code != 0 {
			t.Fatalf("%s: code %d, err %v, output %q", shell, code, err, buf.String())
		}
		if !bytes.HasPrefix(buf.Bytes(), []byte("caf\xc3\xa9")) {
			t.Errorf("%s: output % x, want UTF-8 café", shell, buf.Bytes())
		}
		gotIn, _, _ := getConsoleCP.Call()
		gotOut, _, _ := getConsoleOutCP.Call()
		if gotIn != 437 || gotOut != 437 {
			t.Errorf("%s: the console is on input %d, output %d after the gate, want 437 and 437", shell, gotIn, gotOut)
		}
	}
}
