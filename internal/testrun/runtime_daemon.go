package testrun

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/saphan-ai/proven/internal/network"
)

// envRuntimeSetting is the setting a refusal about the runtime names.
const envRuntimeSetting = EnvRuntime

// podmanSystemConfs are the files podman reads before the user's own. A file
// there can name a remote service as well as one under HOME.
var podmanSystemConfs = []string{
	"/usr/share/containers/containers.conf",
	"/etc/containers/containers.conf",
}

// checkRuntime refuses, in local mode, a runtime whose daemon address Proven
// cannot judge (ADR-0034). The runtime is any command, and each one finds its
// daemon its own way. docker is judged by DOCKER_HOST, CONTAINER_HOST and the
// docker context (checkDaemon). podman is judged by the remote services of its
// containers.conf files and its connections file under HOME. A command named
// docker may be podman (the podman-docker package installs one) and Proven
// cannot tell, so it is judged by the podman sources too. Any other runtime
// is refused by name: Proven does not know where it would connect.
func checkRuntime(policy *network.Policy, env func(string) string, runtime string) error {
	if !policy.Local() {
		return nil
	}
	switch filepath.Base(runtime) {
	case "docker", "podman":
		return checkPodman(policy, env)
	}
	return refuse(envRuntimeSetting, clip(runtime),
		"PROVEN_NETWORK=local cannot tell where this runtime's daemon is, so it is not run",
		"docker or podman, whose daemon address Proven reads and checks")
}

// checkPodman reads every file podman takes a remote service from and checks
// each destination, not only the active one: an address that is configured
// is an address a changed active_service or --connection reaches. A file that
// exists and cannot be read is refused, because its content is not judged.
func checkPodman(policy *network.Policy, env func(string) string) error {
	var files []string
	files = append(files, podmanSystemConfs...)
	files = append(files, confDropIns("/etc/containers")...)
	if home := env("HOME"); home != "" {
		dir := filepath.Join(home, ".config", "containers")
		files = append(files, filepath.Join(dir, "containers.conf"))
		files = append(files, confDropIns(dir)...)
	}
	for _, path := range files {
		b, err := readSmall(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return refuse(envRuntimeSetting, clip(path), "a podman configuration file cannot be read: "+err.Error(),
				"a podman configuration that Proven can read, or none")
		}
		uris, err := serviceDestinations(b)
		if err != nil {
			return refuse(envRuntimeSetting, clip(path), "a podman configuration file is not TOML Proven can read: "+err.Error(),
				"a podman configuration that Proven can read, or none")
		}
		for _, uri := range uris {
			if err := policy.CheckDaemon("podman service destination in "+clip(path), uri); err != nil {
				return fromNetwork(err)
			}
		}
	}
	if home := env("HOME"); home != "" {
		path := filepath.Join(home, ".config", "containers", "podman-connections.json")
		b, err := readSmall(path)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return refuse(envRuntimeSetting, clip(path), "podman's connections file cannot be read: "+err.Error(),
				"a connections file that Proven can read, or none")
		}
		uris, err := connectionURIs(b)
		if err != nil {
			return refuse(envRuntimeSetting, clip(path), "podman's connections file is not JSON Proven can read: "+err.Error(),
				"a connections file that Proven can read, or none")
		}
		for _, uri := range uris {
			if err := policy.CheckDaemon("podman connection in "+clip(path), uri); err != nil {
				return fromNetwork(err)
			}
		}
	}
	return nil
}

// confDropIns lists the *.conf files of dir/containers.conf.d, in name order.
func confDropIns(dir string) []string {
	matches, _ := filepath.Glob(filepath.Join(dir, "containers.conf.d", "*.conf"))
	sort.Strings(matches)
	return matches
}

// serviceDestinations parses a containers.conf as TOML and returns the uri of
// every [engine.service_destinations.<name>] it holds, whatever way the file
// spells them: a dotted key, a quoted key or table name, an inline table, a
// literal string, an escape. Podman reads the file with the same parser and
// matches a field name without regard to case, so a key is matched that way.
// A document that does not parse, or whose destinations, remote flag or active
// service are not the types podman reads, is an error: what cannot be read
// with certainty is not judged to be safe.
func serviceDestinations(b []byte) ([]string, error) {
	var doc map[string]any
	if _, err := toml.Decode(string(b), &doc); err != nil {
		return nil, err
	}
	engine, err := tableKey(doc, "engine")
	if err != nil || engine == nil {
		return nil, err
	}
	if v, ok := lookupKey(engine, "remote"); ok {
		if _, isBool := v.(bool); !isBool {
			return nil, fmt.Errorf("engine.remote is not a boolean")
		}
	}
	if v, ok := lookupKey(engine, "active_service"); ok {
		if _, isString := v.(string); !isString {
			return nil, fmt.Errorf("engine.active_service is not a string")
		}
	}
	dests, err := tableKey(engine, "service_destinations")
	if err != nil || dests == nil {
		return nil, err
	}
	var uris []string
	for name, raw := range dests {
		dest, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("engine.service_destinations.%s is not a table", name)
		}
		for key, v := range dest {
			if !strings.EqualFold(key, "uri") {
				continue
			}
			uri, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("engine.service_destinations.%s.uri is not a string", name)
			}
			uris = append(uris, uri)
		}
	}
	sort.Strings(uris)
	return uris, nil
}

// lookupKey returns the value of the key named want in m, matched without regard
// to case as podman's TOML decoder matches a field. Keys that differ only by case
// are read as the first in sorted order; tableKey refuses a table set twice.
func lookupKey(m map[string]any, want string) (any, bool) {
	var found []string
	for k := range m {
		if strings.EqualFold(k, want) {
			found = append(found, k)
		}
	}
	if len(found) == 0 {
		return nil, false
	}
	sort.Strings(found)
	return m[found[0]], true
}

// tableKey returns the table named want in m, nil when m has none, and an error
// when the key is present and is not a table or is present in more than one case.
func tableKey(m map[string]any, want string) (map[string]any, error) {
	n := 0
	for k := range m {
		if strings.EqualFold(k, want) {
			n++
		}
	}
	if n > 1 {
		return nil, fmt.Errorf("%s is set more than once, in different case", want)
	}
	v, ok := lookupKey(m, want)
	if !ok {
		return nil, nil
	}
	t, ok := v.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s is not a table", want)
	}
	return t, nil
}

// connectionURIs returns every URI of podman-connections.json.
func connectionURIs(b []byte) ([]string, error) {
	var f struct {
		Connection struct {
			Connections map[string]struct {
				URI string `json:"URI"`
			} `json:"Connections"`
		} `json:"Connection"`
	}
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, err
	}
	var uris []string
	for _, c := range f.Connection.Connections {
		if c.URI != "" {
			uris = append(uris, c.URI)
		}
	}
	return uris, nil
}
