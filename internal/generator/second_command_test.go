package generator

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/service"
)

func TestSecondCommandHasIndependentArgvModelAndPolicy(t *testing.T) {
	env := secondEnv(map[string]string{
		EnvMode: ModeCommand, EnvCommand: `["/bin/false"]`, EnvModel: "first-model", EnvSkepticModel: "first-skeptic",
		EnvSecondCommand: `["/usr/bin/printf","%s:%s","{agent}","{model}"]`, EnvSecondModel: "second-model",
		EnvTimeout: "9s", EnvSecondTimeout: "2s", EnvRecoveryRetries: "3", "PROVEN_SECOND_REVIEWER_RECOVERY_RETRIES": "0",
		// Neither API address/key is used by command mode.
		EnvSecondBaseURL: "not-an-api", EnvSecondKey: "unused",
	})
	cfg, connected, err := LoadSecond(env)
	if err != nil || !connected || cfg.Mode != ModeCommand {
		t.Fatalf("cfg=%+v connected=%v err=%v", cfg, connected, err)
	}
	g := NewSecond(cfg, nil)
	answer, err := g.Complete(context.Background(), "skeptic", "prompt; $(false)")
	if err != nil || answer != "skeptic:second-model" {
		t.Fatalf("%q %v", answer, err)
	}
	u, ok := g.LastUse()
	if !ok || u.Model != "unknown" || u.Requested != "second-model" || u.Route != "command" || u.FromReceipt {
		t.Fatalf("CLI identity was invented: %+v %v", u, ok)
	}
	if p := PolicyOf(g); p.AttemptTimeout != 2*time.Second || p.RecoveryRetries != 0 {
		t.Fatalf("second policy lost: %+v", p)
	}
	if attempts := g.LastAttempts(); len(attempts) != 0 {
		t.Fatalf("CLI reported HTTP requests: %+v", attempts)
	}
}

func TestSecondCommandRefusesInvalidOrMissingInputs(t *testing.T) {
	for _, tc := range []struct{ field, value, want string }{
		{EnvSecondCommand, "not-json", EnvSecondCommand},
		{EnvSecondCommand, "", EnvSecondCommand},
		{EnvSecondModel, "", EnvSecondModel},
		{EnvSecondModel, "bad model", EnvSecondModel},
		{EnvSecondTimeout, "0s", EnvSecondTimeout},
		{"PROVEN_SECOND_REVIEWER_RECOVERY_RETRIES", "bad", "PROVEN_SECOND_REVIEWER_RECOVERY_RETRIES"},
		{EnvSecondMode, "vendor", EnvSecondMode},
		{"PROVEN_NETWORK", "local", EnvSecondMode},
	} {
		t.Run(tc.field+"="+tc.value, func(t *testing.T) {
			values := map[string]string{EnvSecondMode: ModeCommand, EnvSecondCommand: `["/bin/cat"]`, EnvSecondModel: "second-model", EnvCommand: `["/bin/cat"]`, EnvModel: "first-model"}
			values[tc.field] = tc.value
			_, connected, err := LoadSecond(secondEnv(values))
			var ce *ConfigError
			if connected || !errors.As(err, &ce) || ce.Field != tc.want {
				t.Fatalf("connected=%v err=%v want field %s", connected, err, tc.want)
			}
		})
	}
}

func TestSecondCommandFailureIsAttributedAndClearsIdentity(t *testing.T) {
	g := NewSecond(SecondSettings{Mode: ModeCommand, Command: []string{"/bin/sleep", "5"}, Model: "second", Timeout: 20 * time.Millisecond}, nil)
	_, err := g.Complete(context.Background(), "planner", "prompt")
	var down *service.Down
	if !errors.As(err, &down) || down.Service != SecondService || !strings.Contains(down.Reason, "timeout") {
		t.Fatalf("%v", err)
	}
	if _, ok := g.LastUse(); ok {
		t.Fatal("failed command reported identity")
	}
}

func TestSecondModeOffIgnoresStaleConnection(t *testing.T) {
	_, connected, err := LoadSecond(secondEnv(map[string]string{EnvSecondMode: ModeOff, EnvSecondCommand: "bad", EnvSecondBaseURL: "bad", EnvSecondKey: "bad"}))
	if err != nil || connected {
		t.Fatalf("connected=%v err=%v", connected, err)
	}
}
