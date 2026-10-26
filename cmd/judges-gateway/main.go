// Command judges-gateway lets a named judge's lab reach our model service and
// our evidence store with the judge's own expiring token (ADR-0051).
//
//	judges-gateway [serve]
//	judges-gateway token new --id <id> --expires <RFC 3339 time> [--per-minute N] [--per-day N]
//	judges-gateway token revoke --id <id>
//	judges-gateway token list
//
// It listens on 127.0.0.1 only; the Cloudflare tunnel connects to it there.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/saphan-ai/proven/internal/gateway"
)

const (
	exitOK    = 0
	exitError = 1
	exitUsage = 2
)

// The defaults of a new token's limits. The settings file holds the values
// that apply; these only fill a flag left out.
const (
	defaultPerMinute = 20
	defaultPerDay    = 500
)

const helpText = `judges-gateway serves a judge's lab through our tunnel.

Usage:
  judges-gateway [serve]
  judges-gateway token new --id <id> --expires <RFC 3339 time> [--per-minute N] [--per-day N]
  judges-gateway token revoke --id <id>
  judges-gateway token list

Settings come from the environment; docs/JUDGES_GATEWAY.md lists them.
JUDGES_GATEWAY_TOKENS_FILE names the token settings file, which the token
commands read and write. 'token new' prints the token once and stores only its hash.

Exit codes: 0 done, 1 refused or failed, 2 usage.
`

func main() {
	os.Exit(run(os.Args[1:], os.Getenv, os.Stdout, os.Stderr, nil))
}

// run is main with its inputs named, so a test can drive it. ready, when it
// is not nil, receives the address the server listens on.
func run(args []string, getenv func(string) string, stdout, stderr io.Writer, ready func(net.Addr)) int {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help" || args[0] == "help") {
		fmt.Fprint(stdout, helpText)
		return exitOK
	}
	if len(args) == 0 || args[0] == "serve" {
		return serve(getenv, stderr, ready)
	}
	if args[0] == "token" {
		return token(args[1:], getenv, stdout, stderr)
	}
	fmt.Fprint(stderr, helpText)
	return exitUsage
}

func serve(getenv func(string) string, stderr io.Writer, ready func(net.Addr)) int {
	s, err := gateway.Load(getenv)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitError
	}
	s.Config.Alarm = stderr
	s.Config.Fault = stderr
	s.Config.Log = os.Stdout
	if s.LogFile != "" {
		f, ferr := os.OpenFile(s.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
		if ferr != nil {
			fmt.Fprintf(stderr, "configuration refused: %s: the log file cannot be opened\n", gateway.EnvLogFile)
			return exitError
		}
		defer f.Close()
		s.Config.Log = f
	}
	// 127.0.0.1 only: the tunnel connects here, and nothing else may.
	ln, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(s.Port)))
	if err != nil {
		fmt.Fprintf(stderr, "cannot listen on 127.0.0.1:%d: %v\n", s.Port, err)
		return exitError
	}
	srv := &http.Server{
		Handler:           gateway.New(s.Config),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       time.Minute,
		IdleTimeout:       time.Minute,
	}
	fmt.Fprintf(stderr, "judges-gateway listening on %s\n", ln.Addr())
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if ready != nil {
		ready(ln.Addr())
	}
	done := make(chan struct{})
	go func() {
		<-ctx.Done()
		shut, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shut)
		close(done)
	}()
	if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Fprintln(stderr, "serve:", err)
		return exitError
	}
	<-done
	return exitOK
}

func token(args []string, getenv func(string) string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, helpText)
		return exitUsage
	}
	path := getenv(gateway.EnvTokensFile)
	if path == "" {
		fmt.Fprintf(stderr, "configuration refused: %s: the token settings file is not named\n", gateway.EnvTokensFile)
		return exitError
	}
	fs := flag.NewFlagSet("token "+args[0], flag.ContinueOnError)
	fs.SetOutput(stderr)
	id := fs.String("id", "", "the token's id: a judge's name, lower case")
	expires := fs.String("expires", "", "when the token stops working, RFC 3339")
	perMinute := fs.Int("per-minute", defaultPerMinute, "calls a minute")
	perDay := fs.Int("per-day", defaultPerDay, "calls a day")
	if err := fs.Parse(args[1:]); err != nil {
		return exitUsage
	}
	now := time.Now()
	switch args[0] {
	case "new":
		exp, err := time.Parse(time.RFC3339, *expires)
		if err != nil {
			fmt.Fprintln(stderr, "--expires is required and is an RFC 3339 time, for example 2026-10-28T00:00:00Z")
			return exitUsage
		}
		tok, err := gateway.Add(path, *id, exp, *perMinute, *perDay, now)
		if err != nil {
			fmt.Fprintln(stderr, "token new refused:", err)
			return exitError
		}
		// The one place the token is shown. Send it to the judge privately.
		fmt.Fprintf(stdout, "id: %s\nexpires: %s\ntoken: %s\n", *id, exp.UTC().Format(time.RFC3339), tok)
		fmt.Fprintln(stderr, "This is the only time the token is shown. Only its hash is stored.")
		return exitOK
	case "revoke":
		if err := gateway.Revoke(path, *id, now); err != nil {
			fmt.Fprintln(stderr, "token revoke refused:", err)
			return exitError
		}
		fmt.Fprintf(stdout, "revoked: %s\n", *id)
		return exitOK
	case "list":
		recs, err := gateway.List(path)
		if err != nil {
			fmt.Fprintln(stderr, "token list refused:", err)
			return exitError
		}
		tw := tabwriter.NewWriter(stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(tw, "id\texpires\tper_minute\tper_day\tstate")
		for _, r := range recs {
			state := "active"
			switch {
			case r.RevokedAt != "":
				state = "revoked " + r.RevokedAt
			case !now.Before(r.Expires):
				state = "expired"
			}
			fmt.Fprintf(tw, "%s\t%s\t%d\t%d\t%s\n", r.ID, r.Expires.UTC().Format(time.RFC3339), r.PerMinute, r.PerDay, state)
		}
		_ = tw.Flush()
		return exitOK
	}
	fmt.Fprint(stderr, helpText)
	return exitUsage
}
