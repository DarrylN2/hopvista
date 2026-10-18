//go:build windows && amd64

package appsmoke

import (
	"bytes"
	"debug/pe"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func copyPackagedExecutable(t *testing.T) string {
	t.Helper()
	source := os.Getenv("HOPVISTA_PACKAGED_EXE")
	if source == "" {
		t.Skip("set HOPVISTA_PACKAGED_EXE to an absolute built executable")
	}
	if !filepath.IsAbs(source) {
		t.Fatal("HOPVISTA_PACKAGED_EXE must be absolute")
	}
	image, err := pe.Open(source)
	if err != nil {
		t.Fatal(err)
	}
	if image.FileHeader.Machine != pe.IMAGE_FILE_MACHINE_AMD64 {
		image.Close()
		t.Fatal("packaged executable must target Windows x64")
	}
	image.Close()
	executable := filepath.Join(t.TempDir(), "hopvista.exe")
	from, err := os.Open(source)
	if err != nil {
		t.Fatal(err)
	}
	defer from.Close()
	to, err := os.Create(executable)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(to, from); err != nil {
		to.Close()
		t.Fatal(err)
	}
	if err := to.Close(); err != nil {
		t.Fatal(err)
	}
	return executable
}

// exec.Cmd drains both output streams before Wait returns. Tests inspect these
// buffers only after stop, so they include all output without concurrent reads.
type packagedProcess struct {
	cmd     *exec.Cmd
	base    string
	stdout  bytes.Buffer
	stderr  listenerOutput
	stopped bool
}

type listenerOutput struct {
	buffer   bytes.Buffer
	pending  string
	listener chan string
}

func (out *listenerOutput) Write(data []byte) (int, error) {
	n, err := out.buffer.Write(data)
	out.pending += string(data)
	for {
		line, rest, found := strings.Cut(out.pending, "\n")
		if !found {
			break
		}
		out.pending = rest
		if at := strings.Index(line, "HopVista: http://127.0.0.1:"); at >= 0 {
			base := strings.TrimSuffix(strings.TrimSpace(strings.TrimPrefix(line[at:], "HopVista: ")), "/")
			select {
			case out.listener <- base:
			default:
			}
		}
	}
	return n, err
}

func startPackagedProcess(t *testing.T, executable string) *packagedProcess {
	t.Helper()
	p := &packagedProcess{stderr: listenerOutput{listener: make(chan string, 1)}}
	p.cmd = exec.Command(executable)
	p.cmd.Dir = filepath.Dir(executable)
	p.cmd.Env = systemOnlyPATH(os.Environ())
	p.cmd.Stdout, p.cmd.Stderr = &p.stdout, &p.stderr
	if err := p.cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { p.stop(t) })
	select {
	case p.base = <-p.stderr.listener:
	case <-time.After(10 * time.Second):
		t.Fatal("packaged executable did not announce its listener")
	}
	return p
}

func (p *packagedProcess) stop(t *testing.T) {
	t.Helper()
	if p.stopped {
		return
	}
	p.stopped = true
	if err := p.cmd.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
		t.Errorf("stop packaged executable: %v", err)
	}
	var exit *exec.ExitError
	if err := p.cmd.Wait(); err != nil && !errors.As(err, &exit) {
		t.Errorf("wait for packaged executable: %v", err)
	}
}
