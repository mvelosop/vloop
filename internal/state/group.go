package state

import (
	"io"
	"os"
	"os/exec"
	"time"
)

// drainWait bounds how long RunGroup waits for the output to be read once the
// command's group has been killed.
const drainWait = 2 * time.Second

// RunGroup runs c in a process group of its own (a Job object on Windows) and
// copies its output to c.Stdout and c.Stderr. After timeout (none when it is
// zero or less) the whole group is killed and timedOut is true; when the
// command itself exits, whatever is left in its group is killed too, so a
// background process cannot hold the output open. The read of the output is
// bounded either way. The error is that of c.Wait, or of starting it.
func RunGroup(c *exec.Cmd, timeout time.Duration) (timedOut bool, err error) {
	stdout, stderr := c.Stdout, c.Stderr
	var readers []*os.File
	var done []chan struct{}
	var writers []*os.File
	pipe := func(dst io.Writer) (*os.File, error) {
		if dst == nil {
			return nil, nil
		}
		r, w, err := os.Pipe()
		if err != nil {
			return nil, err
		}
		readers, writers = append(readers, r), append(writers, w)
		ch := make(chan struct{})
		done = append(done, ch)
		go func() {
			defer close(ch)
			_, _ = io.Copy(dst, r)
		}()
		return w, nil
	}
	closeAll := func() {
		for _, w := range writers {
			w.Close()
		}
		for _, r := range readers {
			r.Close()
		}
	}
	ow, err := pipe(stdout)
	if err != nil {
		closeAll()
		return false, err
	}
	ew := ow
	if stderr != nil && !sameWriter(stdout, stderr) {
		if ew, err = pipe(stderr); err != nil {
			closeAll()
			return false, err
		}
	}
	if ow != nil {
		c.Stdout = ow
	} else {
		c.Stdout = nil
	}
	if ew != nil {
		c.Stderr = ew
	} else {
		c.Stderr = nil
	}
	defer func() { c.Stdout, c.Stderr = stdout, stderr }()

	g, err := startGroup(c)
	for _, w := range writers {
		w.Close()
	}
	writers = nil
	if err != nil {
		closeAll()
		return false, err
	}
	defer g.close()

	var timer *time.Timer
	fired := make(chan struct{})
	if timeout > 0 {
		timer = time.AfterFunc(timeout, func() {
			g.kill()
			close(fired)
		})
	}
	err = c.Wait()
	if timer != nil && !timer.Stop() {
		<-fired
		timedOut = true
	}
	g.kill()

	deadline := time.After(drainWait)
	for _, ch := range done {
		select {
		case <-ch:
		case <-deadline:
			closeAll()
			return timedOut, err
		}
	}
	closeAll()
	return timedOut, err
}

// sameWriter reports whether a and b are the same writer, so that they share
// one pipe and their output keeps its order. Writers that cannot be compared
// are different.
func sameWriter(a, b io.Writer) (same bool) {
	defer func() {
		if recover() != nil {
			same = false
		}
	}()
	return a == b
}
