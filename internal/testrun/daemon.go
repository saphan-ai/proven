package testrun

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/saphan-ai/proven/internal/network"
)

// Names the container runtime reads to find its daemon. The runtime process keeps
// the environment (commandOutput), so each of these decides where the test runs.
const (
	envDockerHost    = "DOCKER_HOST"
	envContainerHost = "CONTAINER_HOST"
	envDockerContext = "DOCKER_CONTEXT"
	envDockerConfig  = "DOCKER_CONFIG"
)

// maxDaemonFile bounds a docker config or context file, which is read to judge an address.
const maxDaemonFile = 1 << 20

// checkDaemon refuses, in local mode, a daemon address that is not a local socket
// or a host on the organisation's network (ADR-0034). The address comes from
// DOCKER_HOST, CONTAINER_HOST, or the docker context in force. The dial check of
// the HTTP clients never sees this connection, so the load check is the guard.
func checkDaemon(policy *network.Policy, env func(string) string) error {
	if !policy.Local() {
		return nil
	}
	for _, name := range []string{envDockerHost, envContainerHost} {
		if raw := env(name); raw != "" {
			if err := policy.CheckDaemon(name, raw); err != nil {
				return fromNetwork(err)
			}
		}
	}
	ctxName, err := dockerContext(env)
	if err != nil {
		return refuse(envDockerContext, "", err.Error(), "a docker context whose daemon address can be read, or none")
	}
	if ctxName == "" || ctxName == "default" {
		return nil
	}
	host, otelEndpoint, err := contextHost(env, ctxName)
	if err != nil {
		return refuse(envDockerContext, clip(ctxName), "the docker context's daemon address cannot be read: "+err.Error(), "a context whose daemon is a local socket or a host on the organisation's network, or the default context")
	}
	if err := checkContextTelemetry(policy, ctxName, otelEndpoint); err != nil {
		return err
	}
	if host == "" {
		return nil
	}
	if err := policy.CheckDaemon(envDockerContext, host); err != nil {
		if ce, ok := network.As(err); ok {
			ce.Value = clip(ctxName + " -> " + host)
			return fromNetwork(ce)
		}
		return err
	}
	return nil
}

// checkContextTelemetry judges the OTLP endpoint a docker context names in its
// metadata (the field "otel", key OTEL_EXPORTER_OTLP_ENDPOINT): the docker CLI
// exports its own telemetry there (docker/cli, cli/command/telemetry_docker.go).
// A unix socket is local. Any other endpoint must be an IP address on the
// organisation's network, because the CLI resolves a name with its own resolver.
func checkContextTelemetry(policy *network.Policy, ctxName, endpoint string) error {
	if endpoint == "" || strings.HasPrefix(endpoint, "unix://") {
		return nil
	}
	value := clip(ctxName + " -> otel " + endpoint)
	if err := policy.CheckURL(envDockerContext, endpoint); err != nil {
		if ce, ok := network.As(err); ok {
			ce.Value = value
			return fromNetwork(ce)
		}
		return err
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return refuse(envDockerContext, value, "the docker context's telemetry endpoint cannot be read: "+err.Error(), "a context whose telemetry endpoint is a local socket or an IP address on the organisation's network, or none")
	}
	bare, _, _ := strings.Cut(u.Hostname(), "%")
	if _, err := netip.ParseAddr(bare); err != nil {
		return refuse(envDockerContext, value, "host "+u.Hostname()+" is a name, and the docker CLI would resolve it with its own resolver to send its telemetry (PROVEN_NETWORK=local)", "a context whose telemetry endpoint is a local socket or an IP address on the organisation's network, or none")
	}
	return nil
}

// fromNetwork maps a refusal of the network package onto this package's ConfigError.
func fromNetwork(err error) error {
	if ne, ok := network.As(err); ok {
		return &ConfigError{Field: ne.Field, Value: ne.Value, Reason: ne.Reason, Allowed: ne.Allowed}
	}
	return err
}

// dockerContext returns the context in force: DOCKER_CONTEXT, else the
// currentContext of the docker config file. A missing config file is no context.
func dockerContext(env func(string) string) (string, error) {
	if name := env(envDockerContext); name != "" {
		return name, nil
	}
	dir := dockerConfigDir(env)
	if dir == "" {
		return "", nil
	}
	b, err := readSmall(filepath.Join(dir, "config.json"))
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	var cfg struct {
		CurrentContext string `json:"currentContext"`
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		return "", err
	}
	return cfg.CurrentContext, nil
}

// contextHost returns the docker endpoint of a named context and the OTLP
// endpoint its metadata names for the docker CLI's telemetry, if any. Docker keeps it in
// contexts/meta/<sha256 of the name>/meta.json.
func contextHost(env func(string) string, name string) (string, string, error) {
	dir := dockerConfigDir(env)
	if dir == "" {
		return "", "", os.ErrNotExist
	}
	sum := sha256.Sum256([]byte(name))
	b, err := readSmall(filepath.Join(dir, "contexts", "meta", hex.EncodeToString(sum[:]), "meta.json"))
	if err != nil {
		return "", "", err
	}
	var meta struct {
		Metadata  map[string]any `json:"Metadata"`
		Endpoints map[string]struct {
			Host string `json:"Host"`
		} `json:"Endpoints"`
	}
	if err := json.Unmarshal(b, &meta); err != nil {
		return "", "", err
	}
	otelEndpoint := ""
	if otel, ok := meta.Metadata["otel"].(map[string]any); ok {
		otelEndpoint, _ = otel["OTEL_EXPORTER_OTLP_ENDPOINT"].(string)
	}
	return meta.Endpoints["docker"].Host, otelEndpoint, nil
}

// dockerConfigDir is DOCKER_CONFIG, else .docker in the home directory.
func dockerConfigDir(env func(string) string) string {
	if dir := env(envDockerConfig); dir != "" {
		return dir
	}
	if home := env("HOME"); home != "" {
		return filepath.Join(home, ".docker")
	}
	return ""
}

// readSmall reads a file of at most maxDaemonFile bytes. A larger file is an
// error, never a prefix: the runtime reads all of it, so all of it is judged.
func readSmall(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, maxDaemonFile+1))
	if err == nil && len(b) > maxDaemonFile {
		return nil, fmt.Errorf("the file is larger than %d bytes", maxDaemonFile)
	}
	return b, err
}
