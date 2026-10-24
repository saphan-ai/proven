package testrun

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

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

// serviceURI matches a uri value of a podman service destination.
var serviceURI = regexp.MustCompile(`(?i)\buri\s*=\s*["']([^"']*)["']`)

// checkRuntime refuses, in local mode, a runtime whose daemon address Proven
// cannot judge (ADR-0034). The runtime is any command, and each one finds its
// daemon its own way. docker is judged by DOCKER_HOST, CONTAINER_HOST and the
// docker context (checkDaemon). podman is judged by the remote services of its
// containers.conf files and its connections file under HOME. Any other runtime
// is refused by name: Proven does not know where it would connect.
func checkRuntime(policy *network.Policy, env func(string) string, runtime string) error {
	if !policy.Local() {
		return nil
	}
	switch filepath.Base(runtime) {
	case "docker":
		return nil
	case "podman":
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
		for _, uri := range serviceDestinations(string(b)) {
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

// serviceDestinations returns every uri set in a service_destinations table of a
// containers.conf text: under a [engine.service_destinations.<name>] header, or
// on a line that also names service_destinations (an inline table).
func serviceDestinations(text string) []string {
	var uris []string
	inService := false
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[") {
			inService = strings.Contains(line, "service_destinations")
			continue
		}
		if !inService && !strings.Contains(line, "service_destinations") {
			continue
		}
		for _, m := range serviceURI.FindAllStringSubmatch(line, -1) {
			uris = append(uris, m[1])
		}
	}
	return uris
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
