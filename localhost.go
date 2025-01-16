package sup

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"

	"github.com/pkg/errors"
)

// LocalhostClient is a wrapper over the SSH connection/sessions.
type LocalhostClient struct {
	cmd     *exec.Cmd
	user    string
	stdin   io.WriteCloser
	stdout  io.Reader
	stderr  io.Reader
	running bool
	env     string //export FOO="bar"; export BAR="baz";
}

func (c *LocalhostClient) Connect() (err error) {
	var u *user.User
	if u, err = user.Current(); err != nil {
		return
	}

	c.user = u.Username
	return
}

func (c *LocalhostClient) Run(task *Task) (err error) {
	if c.running {
		return fmt.Errorf("Command already running")
	}

	// Create shell command that includes environment setup
	shellCmd := task.Run
	if c.env != "" {
		shellCmd = c.env + shellCmd
	}

	// Create command with shell interpretation
	cmd := exec.Command("sh", "-c", shellCmd)

	// Set up environment variables
	cmd.Env = os.Environ()

	// Set up pipes
	if c.stdin, err = cmd.StdinPipe(); err != nil {
		return errors.Wrap(err, "failed to create stdin pipe")
	}

	if c.stdout, err = cmd.StdoutPipe(); err != nil {
		return errors.Wrap(err, "failed to create stdout pipe")
	}

	if c.stderr, err = cmd.StderrPipe(); err != nil {
		return errors.Wrap(err, "failed to create stderr pipe")
	}

	// Start the command
	if err = cmd.Start(); err != nil {
		return ErrTask{task, err.Error()}
	}

	// Handle input if provided
	if task.Input != nil {
		if _, err = io.Copy(c.stdin, task.Input); err != nil {
			return errors.Wrap(err, "copying input failed")
		}
		if err = c.stdin.Close(); err != nil {
			return errors.Wrap(err, "closing input failed")
		}
	}

	c.cmd = cmd
	c.running = true
	return nil
}

func (c *LocalhostClient) Wait() error {
	if !c.running {
		return fmt.Errorf("Trying to wait on stopped command")
	}
	err := c.cmd.Wait()
	c.running = false
	return err
}

func (c *LocalhostClient) Close() error {
	return nil
}

func (c *LocalhostClient) Stdin() io.WriteCloser {
	if c.cmd != nil && c.cmd.Stdin != nil {
		if writer, ok := c.cmd.Stdin.(io.WriteCloser); ok {
			return writer
		}
	}
	return c.stdin
}

func (c *LocalhostClient) Stderr() io.Reader {
	if c.cmd != nil && c.cmd.Stderr != nil {
		if reader, ok := c.cmd.Stderr.(io.Reader); ok {
			return reader
		}
	}
	return c.stderr
}

func (c *LocalhostClient) Stdout() io.Reader {
	if c.cmd != nil && c.cmd.Stdout != nil {
		if reader, ok := c.cmd.Stdout.(io.Reader); ok {
			return reader
		}
	}
	return c.stdout
}

func (c *LocalhostClient) Prefix() (string, int) {
	host := c.user + "@localhost" + " | "
	return ResetColor + host, len(host)
}

func (c *LocalhostClient) Write(p []byte) (n int, err error) {
	return c.stdin.Write(p)
}

func (c *LocalhostClient) WriteClose() error {
	return c.stdin.Close()
}

func (c *LocalhostClient) Signal(sig os.Signal) error {
	return c.cmd.Process.Signal(sig)
}

func ResolveLocalPath(cwd, path, env string) (string, error) {
	// Check if file exists first. Use bash to resolve $ENV_VARs.
	cmd := exec.Command("bash", "-c", env+"echo -n "+path)
	cmd.Dir = cwd
	resolvedFilename, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, "resolving path failed")
	}

	return string(resolvedFilename), nil
}
