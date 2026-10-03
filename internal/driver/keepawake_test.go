package driver

import (
	"errors"
	"reflect"
	"testing"
)

func TestKeepAwakeCommands(t *testing.T) {
	yes := func(string) bool { return true }
	no := func(string) bool { return false }
	got, err := KeepAwakeCommand("darwin", 42, no)
	if want := []string{"caffeinate", "-i", "-w", "42"}; err != nil || !reflect.DeepEqual(got, want) {
		t.Errorf("darwin = %v, %v", got, err)
	}
	got, err = KeepAwakeCommand("linux", 42, yes)
	if err != nil || len(got) < 4 || got[0] != "systemd-inhibit" || got[1] != "--what=idle" || got[2] != "--why=vloop run" {
		t.Errorf("linux = %v, %v", got, err)
	}
	if _, err = KeepAwakeCommand("linux", 42, no); err == nil {
		t.Error("linux without systemd-inhibit must be an error")
	}
	if got, err = KeepAwakeCommand("windows", 42, no); got != nil || err != nil {
		t.Errorf("windows = %v, %v: the hold is an API call", got, err)
	}
	if _, err = KeepAwakeCommand("plan9", 42, yes); err == nil {
		t.Error("an unknown OS must be an error")
	}
	if AwakeWarning("x") != "vloop: could not keep the machine awake (x) — a sleeping machine stalls the run" {
		t.Error(AwakeWarning("x"))
	}
}

func TestKeepAwakeOff(t *testing.T) {
	tried := false
	start := func([]string) error { tried = true; return nil }
	native := func() error { tried = true; return nil }
	for _, goos := range []string{"darwin", "linux", "windows"} {
		if err := keepAwake(false, goos, 1, func(string) bool { return true }, start, native); err != nil || tried {
			t.Fatalf("%s: off attempted a hold (%v, %v)", goos, tried, err)
		}
	}
	if err := keepAwake(true, "darwin", 1, nil, start, native); err != nil || !tried {
		t.Errorf("on did not start the helper: %v", err)
	}
	boom := errors.New("boom")
	if err := keepAwake(true, "darwin", 1, nil, func([]string) error { return boom }, native); err != boom {
		t.Errorf("start failure = %v", err)
	}
}
