// Package childenv builds the environment of every process Proven starts: the
// container runtime and git. The environment is explicit and minimal, never
// Proven's own, so no key Proven holds can reach a child process or the
// output a child prints (ADR-0025, ADR-0005).
package childenv

import "os"

// base is the closed set of variables a child may inherit by default. PATH
// finds the tool and HOME is where the tool looks for its own settings. Neither
// holds a secret. A variable not on this list is never passed.
var base = []string{"PATH", "HOME"}

// Minimal returns the child environment: the base variables, then each name in
// extra, copied from Proven's own environment when it is set there. A caller
// names an extra variable only for something the tool needs and that is not a
// credential, such as the address of the container daemon.
func Minimal(extra ...string) []string {
	return MinimalFrom(os.LookupEnv, extra...)
}

// MinimalFrom is Minimal with the lookup given, so a test can supply its own.
func MinimalFrom(lookup func(string) (string, bool), extra ...string) []string {
	names := append(append([]string(nil), base...), extra...)
	env := make([]string, 0, len(names))
	for _, name := range names {
		if value, ok := lookup(name); ok {
			env = append(env, name+"="+value)
		}
	}
	return env
}
