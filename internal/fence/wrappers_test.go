package fence

import (
	"fmt"
	"strings"
	"testing"
)

func TestPowerShellWrappers(t *testing.T) {
	j, _, _ := newJudge(t, PhaseWork)
	push := "git push is denied by the work fence (Bash(git push:*))"
	cases := []struct {
		name, command, want string
	}{
		{"short", `pwsh -c 'git push'`, push},
		{"long and options", `PowerShell -NoProfile -NonInteractive -NoLogo -ExecutionPolicy Bypass -Command 'git push origin main'`, "git push origin main is denied by the work fence (Bash(git push:*))"},
		{"abbreviations", `PWSh.EXE -NOP -NONI -NOL -EP Bypass -C 'go test ./...'`, ""},
		{"path", `C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe -c 'git push'`, push},
		{"case-insensitive inner command", `pwsh -c 'GIT PUSH'`, "GIT PUSH is denied by the work fence (Bash(git push:*))"},
		{"remaining words", `pwsh -c git push origin main`, "git push origin main is denied by the work fence (Bash(git push:*))"},
		{"missing", `pwsh -c`, "-Command has no command string"},
		{"expanded", `pwsh -c $command`, "the -Command string expands and cannot be judged"},
		{"stdin", `pwsh -Command -`, "pwsh: runs a script, stdin or an encoded command, not a literal -Command string"},
		{"file", `pwsh -File x.ps1`, "pwsh: runs a script, stdin or an encoded command, not a literal -Command string"},
		{"short file", `pwsh -f x.ps1`, "pwsh: runs a script, stdin or an encoded command, not a literal -Command string"},
		{"bare script", `pwsh x.ps1`, "pwsh: runs a script, stdin or an encoded command, not a literal -Command string"},
		{"no command", `pwsh`, "pwsh: runs a script, stdin or an encoded command, not a literal -Command string"},
		{"encoded", `powershell -EncodedCommand Z2l0`, "powershell: runs a script, stdin or an encoded command, not a literal -Command string"},
		{"unsupported option", `pwsh -Sta -c 'git status'`, "pwsh: runs a script, stdin or an encoded command, not a literal -Command string"},
		{"expanded policy value", `pwsh -ExecutionPolicy $policy -c 'git status'`, "pwsh: $policy expands before the command string and cannot be judged"},
		{"expanded policy abbreviation", `powershell -ep $policy -c 'git status'`, "powershell: $policy expands before the command string and cannot be judged"},
	}
	for _, tool := range []string{"bash", "powershell"} {
		for _, tc := range cases {
			t.Run(tool+"/"+tc.name, func(t *testing.T) {
				d := j.Decide(tool, "", tc.command)
				switch {
				case tc.want == "" && !d.Allow:
					t.Errorf("denied: %s", d.Reason)
				case tc.want != "" && (d.Allow || d.Reason != tc.want):
					t.Errorf("got allow=%v reason=%q, want %q", d.Allow, d.Reason, tc.want)
				}
			})
		}
	}
}

func TestPowerShellDynamicCommands(t *testing.T) {
	j, _, _ := newJudge(t, PhaseWork)
	names := []string{
		"Invoke-Expression", "iex", "Invoke-Command", "icm", "Start-Process",
		"saps", "start", "Start-Job", "sajb", "Start-ThreadJob", "Invoke-Item", "ii",
	}
	for _, name := range names {
		forms := []string{
			name + " git push",
			"Get-Content a | " + name + " git push",
			"Get-Content a; " + strings.ToUpper(name) + " git push",
			`C:\Windows\System32\` + name + `.exe git push`,
		}
		for _, command := range forms {
			t.Run(command, func(t *testing.T) {
				d := j.Decide("powershell", "", command)
				written := command
				for _, separator := range []string{"| ", "; "} {
					if i := strings.LastIndex(command, separator); i >= 0 {
						written = command[i+len(separator):]
						break
					}
				}
				written = strings.Fields(written)[0]
				if d.Allow || d.Reason != written+" is not allowed" {
					t.Errorf("got allow=%v reason=%q, want %q", d.Allow, d.Reason, written+" is not allowed")
				}
			})
		}
	}

	for _, tc := range []struct {
		name, command, want string
	}{
		{"bash wrapper", `pwsh -c 'INVOKE-ITEM a'`, "INVOKE-ITEM is not allowed"},
		{"powershell path", `C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe -c "IEX 'git push'"`, "IEX is not allowed"},
		{"parenthesized", `(Get-Content a | iex 'git push').Length`, "iex is not allowed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tool := "powershell"
			if tc.name == "bash wrapper" {
				tool = "bash"
			}
			d := j.Decide(tool, "", tc.command)
			if d.Allow || d.Reason != tc.want {
				t.Errorf("got allow=%v reason=%q, want %q", d.Allow, d.Reason, tc.want)
			}
		})
	}
}

func TestPowerShellScriptBlocks(t *testing.T) {
	j, _, _ := newJudge(t, PhaseWork)
	reason := scriptBlockRefusal
	for _, command := range []string{
		"iex 'git status'; & {git push}",
		"Get-Content a | Invoke-Item b; @{ Name = 'value' }",
		"pwsh -c 'Start-Process git status; ForEach-Object { $_.Name }'",
	} {
		d := j.Decide("powershell", "", command)
		if d.Allow || d.Reason != reason {
			t.Errorf("%q after a denied command: got allow=%v reason=%q, want %q", command, d.Allow, d.Reason, reason)
		}
	}
	for _, command := range []string{
		`Get-ChildItem | ForEach-Object { $_.Name }`,
		`& {git push}`,
		`@{ Name = "value" }`,
		`{git status}`,
		`pwsh -c '& {git push}'`,
	} {
		d := j.Decide("powershell", "", command)
		if d.Allow || d.Reason != reason {
			t.Errorf("%q: got allow=%v reason=%q, want %q", command, d.Allow, d.Reason, reason)
		}
	}
	for _, command := range []string{
		`Write-Output '{ literal }'`,
		`Write-Output "{ literal }"`,
		`Get-ChildItem | ForEach-Object -MemberName Name`,
		`Get-ChildItem | Where-Object Name -like '*.go'`,
		`Get-ChildItem | Select-Object -ExpandProperty Name`,
	} {
		if d := j.Decide("powershell", "", command); !d.Allow {
			t.Errorf("%q denied: %s", command, d.Reason)
		}
	}

	for _, tc := range []struct {
		tool, command, want string
	}{
		{"bash", `{ git push; }`, "git push is denied by the work fence (Bash(git push:*))"},
		{"powershell", `cmd /c "echo {ok}"`, ""},
	} {
		d := j.Decide(tc.tool, "", tc.command)
		if tc.want == "" && !d.Allow {
			t.Errorf("[%s] %q denied: %s", tc.tool, tc.command, d.Reason)
		}
		if tc.want != "" && (d.Allow || d.Reason != tc.want) {
			t.Errorf("[%s] %q: got allow=%v reason=%q, want %q", tc.tool, tc.command, d.Allow, d.Reason, tc.want)
		}
	}
}

func TestPowerShellPropertyStatements(t *testing.T) {
	j, _, _ := newJudge(t, PhaseWork)
	push := "git push is denied by the work fence (Bash(git push:*))"
	parentheses := powerShellParenthesesRefusal
	cases := []struct {
		name, command, want string
	}{
		{"pipeline", `(Get-Content README.md | Measure-Object -Line).Lines`, ""},
		{"leading stage", `(Get-Content a).Name | Measure-Object`, ""},
		{"property path", `(Get-Item a).Name.Value2 | Select-Object -First 1`, ""},
		{"nested", `((Get-Item a).Name).Length`, ""},
		{"denied inner command", `(git push).Length`, push},
		{"denied inner pipeline", `(Get-Content a | git push).Length`, push},
		{"dynamic inner command", `(Get-Content a | iex 'git push').Length`, "iex is not allowed"},
		{"script block inner command", `(Get-Content a | ForEach-Object { $_.Name }).Length`, scriptBlockRefusal},
		{"non-leading pipeline stage", `Get-Content a | (Get-Content b).Name`, parentheses},
		{"non-leading conditional stage", `Get-Content a || (Get-Content b).Name`, parentheses},
		{"method", `(Get-Content a).Trim()`, parentheses},
		{"argument", `Write-Output (Get-Content a)`, parentheses},
		{"array", `@(git status)`, parentheses},
		{"trailing argument", `(Get-Content a).Name suffix`, parentheses},
		{"empty inner", `().Name`, parentheses},
		{"invalid path", `(Get-Content a).Name-Thing`, parentheses},
		{"bash unchanged", `(Get-Content a).Count`, subshellRefusal},
		{"wrapper", `pwsh -c '(Get-Content a).Count'`, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tool := "powershell"
			if tc.name == "bash unchanged" {
				tool = "bash"
			}
			d := j.Decide(tool, "", tc.command)
			if tc.want == "" && !d.Allow {
				t.Errorf("%q denied: %s", tc.command, d.Reason)
			}
			if tc.want != "" && (d.Allow || d.Reason != tc.want) {
				t.Errorf("%q: got allow=%v reason=%q, want %q", tc.command, d.Allow, d.Reason, tc.want)
			}
		})
	}
}

func TestPowerShellEnvironmentAssignments(t *testing.T) {
	j, _, _ := newJudge(t, PhaseWork)
	cases := []struct {
		name, command, want string
	}{
		{"single quoted", `$env:GOOS='linux'; go build ./...`, ""},
		{"double quoted", `$env:GOOS = "linux"; go vet ./...`, ""},
		{"bare", `$env:_NAME=linux; go test ./x`, ""},
		{"case insensitive drive", `$Env:GOOS=linux; go test ./x`, ""},
		{"empty literal", `$env:NAME=''; go test ./x`, ""},
		{"escaped single quote", `$env:NAME='can''t'; go test ./x`, ""},
		{"invalid first character", `$env:1GOOS=linux; go test ./x`, "command word $env:1GOOS=linux starts with $ and cannot be judged"},
		{"expanded compact", `$env:GOOS=$x; go build ./...`, "command word $env:GOOS=$x starts with $ and cannot be judged"},
		{"expanded compact quoted space", `$env:GOOS="$x value"; go build ./...`, `command word $env:GOOS="$x value" starts with $ and cannot be judged`},
		{"expanded spaced", `$env:GOOS = $x; go build ./...`, "command word $env:GOOS starts with $ and cannot be judged"},
		{"double with backtick", "$env:GOOS=\"lin`ux\"; go build ./...", "command word $env:GOOS=\"lin`ux\" starts with $ and cannot be judged"},
		{"parenthesized compact", `$env:GOOS=(linux); go test ./x`, "command word $env:GOOS=(linux) starts with $ and cannot be judged"},
		{"array compact", `$env:GOOS=@(linux); go test ./x`, "command word $env:GOOS=@(linux) starts with $ and cannot be judged"},
		{"extra word", `$env:GOOS=linux extra; go test ./x`, "command word $env:GOOS=linux starts with $ and cannot be judged"},
		{"pipeline position", `Get-Content a | $env:GOOS=linux`, "command word $env:GOOS=linux starts with $ and cannot be judged"},
		{"pipeline suffix", `$env:GOOS=linux | Get-Content`, "command word $env:GOOS=linux starts with $ and cannot be judged"},
		{"bash unchanged", `$env:GOOS=linux; go build ./...`, "command word $env:GOOS=linux starts with $ and cannot be judged"},
		{"wrapper", `pwsh -c '$env:GOOS=linux; go build ./...'`, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tool := "powershell"
			if tc.name == "bash unchanged" {
				tool = "bash"
			}
			d := j.Decide(tool, "", tc.command)
			if tc.want == "" && !d.Allow {
				t.Errorf("%q denied: %s", tc.command, d.Reason)
			}
			if tc.want != "" && (d.Allow || d.Reason != tc.want) {
				t.Errorf("%q: got allow=%v reason=%q, want %q", tc.command, d.Allow, d.Reason, tc.want)
			}
		})
	}
}

func TestPowerShellUnreadableSyntaxGuidance(t *testing.T) {
	j, _, _ := newJudge(t, PhaseWork)
	for _, tc := range []struct {
		tool, command, want string
	}{
		{"powershell", `Write-Output "$(git push)"`, powerShellSubstitutionRefusal},
		{"powershell", "Write-Output a`tb", powerShellBacktickRefusal},
		{"powershell", `pwsh -c 'Write-Output "$(git push)"'`, powerShellSubstitutionRefusal},
		{"bash", `Write-Output "$(git push)"`, "command substitution cannot be judged"},
		{"bash", "Write-Output a`tb", "command substitution cannot be judged"},
	} {
		t.Run(tc.tool+"/"+tc.command, func(t *testing.T) {
			d := j.Decide(tc.tool, "", tc.command)
			if d.Allow || d.Reason != tc.want {
				t.Errorf("got allow=%v reason=%q, want %q", d.Allow, d.Reason, tc.want)
			}
		})
	}
}

func TestCMDWrapper(t *testing.T) {
	j, _, _ := newJudge(t, PhaseWork)
	push := "git push is denied by the work fence (Bash(git push:*))"
	cases := []struct {
		name, command, want string
	}{
		{"single", `cmd /c git push`, push},
		{"keep", `CMD.EXE /D /Q /S /K GIT PUSH`, "GIT PUSH is denied by the work fence (Bash(git push:*))"},
		{"and", `cmd /c git status && git push`, push},
		{"or", `cmd /c 'git status || git push'`, push},
		{"pipe", `cmd /c "git status | git push"`, push},
		{"group", `cmd /c "(git push)"`, push},
		{"nested group", `cmd /c "((git push))"`, push},
		{"group and command", `cmd /c "(git status) && git push"`, push},
		{"semicolon is literal", "cmd /c \"echo ok; git push\"", ""},
		{"newline is literal", "cmd /c \"echo ok\ngit push\"", ""},
		{"percent", `cmd /c echo %PATH%`, "cmd: % and ^ cannot be judged"},
		{"caret", `cmd /c echo a^&git push`, "cmd: % and ^ cannot be judged"},
		{"path", `C:\Windows\System32\cmd.exe /c git push`, push},
		{"missing", `cmd`, "cmd: runs stdin or a batch file, not a literal /c string"},
		{"missing string", `cmd /c`, "cmd: runs stdin or a batch file, not a literal /c string"},
		{"file", `cmd script.cmd`, "cmd: runs stdin or a batch file, not a literal /c string"},
		{"unsupported", `cmd /u /c git status`, "cmd: runs stdin or a batch file, not a literal /c string"},
	}
	for _, tool := range []string{"bash", "powershell"} {
		for _, tc := range cases {
			t.Run(tool+"/"+tc.name, func(t *testing.T) {
				d := j.Decide(tool, "", tc.command)
				switch {
				case tc.want == "" && !d.Allow:
					t.Errorf("denied: %s", d.Reason)
				case tc.want != "" && (d.Allow || d.Reason != tc.want):
					t.Errorf("got allow=%v reason=%q, want %q", d.Allow, d.Reason, tc.want)
				}
			})
		}
	}
}

func TestWrappersPreserveFenceAndDepth(t *testing.T) {
	for _, phase := range []string{PhaseWork, PhaseVerify} {
		j, _, _ := newJudge(t, phase)
		d := j.Decide("powershell", "", `pwsh -c 'echo x > .vloop/state/x'`)
		if d.Allow || !strings.Contains(d.Reason, "not allowed in a "+phase+" session") && phase == PhaseVerify {
			t.Errorf("[%s] wrapper lost phase/root for write: %+v", phase, d)
		}
		if phase == PhaseWork && (d.Allow || !strings.Contains(d.Reason, "write to .vloop/state/x is not allowed in a work session")) {
			t.Errorf("[%s] wrapper lost phase/root for write: %+v", phase, d)
		}
	}

	j, _, _ := newJudge(t, PhaseWork)
	nested := "git status"
	for i := 0; i < 8; i++ {
		nested = fmt.Sprintf("pwsh -c %q", nested)
	}
	if d := j.Decide("bash", "", nested); !d.Allow {
		t.Errorf("eight wrappers denied: %s", d.Reason)
	}
	nested = fmt.Sprintf("pwsh -c %q", nested)
	if d := j.Decide("bash", "", nested); d.Allow || d.Reason != "shell wrappers nest too deep to judge" {
		t.Errorf("ninth wrapper = %+v", d)
	}
}
