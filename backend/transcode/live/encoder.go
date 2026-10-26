package live

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"sen1or/letslive/shared/pkg/logger"
)

const stopGracePeriod = 3 * time.Second

// Encoder is a running ffmpeg process. FLV goes in through Input; one MPEG-TS
// stream per encoded rendition comes out through Outputs.
type Encoder struct {
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	outputs []io.ReadCloser

	done    chan struct{} // closed once the process has been reaped
	waitErr error         // valid after done is closed

	ctx       context.Context // only used to attach request fields to log lines
	closeOnce sync.Once
}

// StartEncoder runs ffmpegPath with args, giving it stdin plus `outputs` extra
// pipes on fd 3, 4, ... that match the pipe:N targets in args.
func StartEncoder(ctx context.Context, ffmpegPath string, args []string, outputs int) (*Encoder, error) {
	if outputs < 0 {
		return nil, fmt.Errorf("invalid output count %d", outputs)
	}

	readEnds := make([]*os.File, 0, outputs)
	writeEnds := make([]*os.File, 0, outputs)
	closeAll := func() {
		for _, f := range readEnds {
			closeLogged(ctx, "ffmpeg output read end", f)
		}
		for _, f := range writeEnds {
			closeLogged(ctx, "ffmpeg output write end", f)
		}
	}

	for i := 0; i < outputs; i++ {
		r, w, err := os.Pipe()
		if err != nil {
			closeAll()
			return nil, fmt.Errorf("create output pipe %d: %w", i, err)
		}
		readEnds = append(readEnds, r)
		writeEnds = append(writeEnds, w)
	}

	cmd := exec.CommandContext(ctx, ffmpegPath, args...)
	cmd.ExtraFiles = writeEnds
	cmd.Cancel = func() error { return cmd.Process.Signal(syscall.SIGTERM) }
	cmd.WaitDelay = stopGracePeriod

	stdin, err := cmd.StdinPipe()
	if err != nil {
		closeAll()
		return nil, fmt.Errorf("open ffmpeg stdin: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		closeLogged(ctx, "ffmpeg stdin", stdin)
		closeAll()
		return nil, fmt.Errorf("open ffmpeg stderr: %w", err)
	}

	if err := cmd.Start(); err != nil {
		closeLogged(ctx, "ffmpeg stdin", stdin)
		closeAll()
		return nil, fmt.Errorf("start ffmpeg: %w", err)
	}

	// The child owns its copies now; keeping ours would stop readers from seeing EOF.
	for _, f := range writeEnds {
		closeLogged(ctx, "ffmpeg output write end", f)
	}

	e := &Encoder{
		ctx:     ctx,
		cmd:     cmd,
		stdin:   stdin,
		outputs: make([]io.ReadCloser, len(readEnds)),
		done:    make(chan struct{}),
	}
	for i, f := range readEnds {
		e.outputs[i] = f
	}

	stderrDone := make(chan struct{})
	go func() {
		defer close(stderrDone)
		logStderr(ctx, stderr)
	}()

	go func() {
		// Wait closes the stderr pipe, so reading must finish first.
		<-stderrDone
		e.waitErr = cmd.Wait()
		close(e.done)
	}()

	return e, nil
}

// closeLogged closes c and logs any failure other than "already closed".
func closeLogged(ctx context.Context, what string, c io.Closer) {
	if err := c.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
		logger.Errorf(ctx, "close %s: %v", what, err)
	}
}

// killLogged force-kills the process and logs any failure other than "already exited".
func killLogged(ctx context.Context, p *os.Process) {
	if err := p.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
		logger.Errorf(ctx, "kill ffmpeg: %v", err)
	}
}

func logStderr(ctx context.Context, r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		logger.Errorf(ctx, "ffmpeg: %s", scanner.Text())
	}
	if err := scanner.Err(); err != nil && !errors.Is(err, os.ErrClosed) {
		logger.Errorf(ctx, "read ffmpeg stderr: %v", err)
	}
}

// Input is the writer ffmpeg reads FLV from. Close it to let ffmpeg flush and exit.
func (e *Encoder) Input() io.WriteCloser { return e.stdin }

// Outputs returns the MPEG-TS readers; index i belongs to the i-th encoded rendition.
func (e *Encoder) Outputs() []io.ReadCloser { return e.outputs }

// Wait blocks until ffmpeg exits and returns its exit error. It is safe to call repeatedly.
func (e *Encoder) Wait() error {
	<-e.done
	return e.waitErr
}

// Stop terminates ffmpeg (SIGTERM, then SIGKILL after 3s), waits for it to exit
// and closes the input and output pipes. It is safe to call more than once and
// after Wait.
func (e *Encoder) Stop() {
	select {
	case <-e.done:
	default:
		if err := e.cmd.Process.Signal(syscall.SIGTERM); err != nil && !errors.Is(err, os.ErrProcessDone) {
			logger.Errorf(e.ctx, "send SIGTERM to ffmpeg: %v", err)
			killLogged(e.ctx, e.cmd.Process)
		}

		timer := time.NewTimer(stopGracePeriod)
		select {
		case <-e.done:
		case <-timer.C:
			killLogged(e.ctx, e.cmd.Process)
			<-e.done
		}
		timer.Stop()
	}

	e.closeOnce.Do(func() {
		closeLogged(e.ctx, "ffmpeg stdin", e.stdin)
		for i, o := range e.outputs {
			closeLogged(e.ctx, fmt.Sprintf("ffmpeg output %d", i), o)
		}
	})
}
