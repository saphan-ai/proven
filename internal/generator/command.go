package generator

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/saphan-ai/proven/internal/service"
)

// Command calls an operator-selected CLI. It is a model transport, not a tool runner.
// The CLI owns its login/subscription; Proven never reads its credential files.
type Command struct {
	cfg  Settings
	last Use
	has  bool
}

func NewCommand(cfg Settings) *Command { return &Command{cfg: cfg} }

func (c *Command) LastUse() (Use, bool) { return c.last, c.has }

func (c *Command) Complete(parent context.Context, agent, prompt string) (string, error) {
	c.last, c.has = Use{}, false
	if err := parent.Err(); err != nil {
		return "", err
	}
	if len(c.cfg.Command) == 0 || c.cfg.Timeout <= 0 {
		return "", &service.Down{Service: "command", Reason: "not configured"}
	}
	ctx, cancel := context.WithTimeout(parent, c.cfg.Timeout)
	defer cancel()
	model := c.cfg.ModelFor(agent)
	args := make([]string, len(c.cfg.Command))
	replace := strings.NewReplacer("{model}", model, "{agent}", agent)
	for i, arg := range c.cfg.Command {
		args[i] = replace.Replace(arg)
	}
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Stdin = strings.NewReader(prompt)
	cmd.Env = commandEnvironment()
	cmd.WaitDelay = 2 * time.Second
	confineCommand(cmd)
	var out limitedAnswer
	cmd.Stdout, cmd.Stderr = &out, io.Discard
	err := cmd.Run()
	if parent.Err() != nil {
		return "", parent.Err()
	}
	if ctx.Err() != nil {
		return "", &service.Down{Service: "command", Reason: "timeout"}
	}
	if err != nil {
		return "", &service.Down{Service: "command", Reason: "CLI failed (check login and command configuration)"}
	}
	answer := strings.TrimSpace(out.String())
	if answer == "" {
		return "", &service.Down{Service: "command", Reason: "empty answer"}
	}
	c.last, c.has = Use{Model: "unknown", Requested: model, Route: "command", ReceiptID: NoReceipt}, true
	return answer, nil
}

type limitedAnswer struct{ bytes.Buffer }

func (b *limitedAnswer) Write(p []byte) (int, error) {
	if b.Len()+len(p) > maxStream {
		return 0, errors.New("model answer exceeds limit")
	}
	return b.Buffer.Write(p)
}

// Only login/configuration locations and ordinary process settings reach the CLI.
// Database keys and Proven's API keys are not inherited by it.
func commandEnvironment() []string {
	var env []string
	for _, key := range []string{"PATH", "HOME", "USER", "LOGNAME", "TMPDIR", "TEMP", "TMP", "LANG", "LC_ALL", "XDG_CONFIG_HOME", "XDG_CACHE_HOME", "CODEX_HOME", "CLAUDE_CONFIG_DIR", "SYSTEMROOT"} {
		if value, ok := os.LookupEnv(key); ok {
			env = append(env, key+"="+value)
		}
	}
	return env
}
