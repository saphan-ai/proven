package testrun

import (
	"context"
	"encoding/json"
	"testing"
)

// Native go test events, rather than text printed by a test process.
func nativeFailureJSON(name, output string) string {
	events := []testEvent{{Action: "output", Test: name, Output: output + "\n"}, {Action: "fail", Test: name}, {Action: "fail"}}
	var text string
	for _, ev := range events {
		b, _ := json.Marshal(ev)
		text += string(b) + "\n"
	}
	return text
}

func TestGoFailureRequiresTheSelectedTestToHaveRun(t *testing.T) {
	for _, tc := range []struct {
		output string
		want   runKind
	}{
		{"compile error", kindNotRun},
		{"{\"Action\":\"fail\",\"Package\":\"p\"}\n", kindNotRun},
		{nativeFailureJSON("TestOther", "other failed"), kindNotRun},
		{nativeFailureJSON("TestChosen", "assertion failed"), kindFail},
	} {
		c := &Container{runtime: "docker", cfg: DefaultContainerConfig()}
		failure := exitCode(t, 1)
		c.run = func(context.Context, []string) (string, error) { return tc.output, failure }
		argv, _ := TestArgv("chosen_test.go", "TestChosen")
		_, kind, _ := c.sideExit(context.Background(), t.TempDir(), argv)
		if kind != tc.want {
			t.Fatalf("%q classified %v, wanted %v", tc.output, kind, tc.want)
		}
	}
}
