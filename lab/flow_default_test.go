package lab

// Enter takes the capitalised default of every question (the owner's live run of 26 October: the first question refused
// an empty Enter). The answers stand in for the terminal, one per line; an empty line is Enter.

import (
	"strings"
	"testing"
)

func TestFlowEnterTakesTheCapitalisedDefaultOfEveryQuestion(t *testing.T) {
	gw, _ := stubGateway(t, 0, "Authorization", "Bearer "+goodToken)
	detect := []string{"DETECT_ZETARIS_URL=http://127.0.0.1:1/x", "DETECT_NIGHTSHIFT_URL=http://127.0.0.1:1/y"}
	cases := []struct {
		name    string
		answers []string
		env     []string
		wantEnv string
		wantOut string
	}{
		{"mode [SAMPLE/live]", []string{"", ""}, nil, "LAB_MODE=sample\n", "[SAMPLE/live]"},
		{"judge [y/N], then own setup [y/N]", []string{"live", "", "", ""}, detect, "LAB_MODE=live on your own setup\n", "[y/N]"},
		{"backend [OURS/own]", []string{"live", "", "y", ""},
			[]string{"PROVEN_LAB_GATEWAY_AUTH=token", "JUDGES_GATEWAY_URL=" + gw.URL, "JUDGES_GATEWAY_TOKEN=" + goodToken},
			"LAB_MODE=live via our gateway\n", "[OURS/own]"},
	}
	run := func(tree string, c int) upResult {
		return runUp(t, upOpts{tree: tree, tty: ttyAnswers(t, cases[c].answers...), env: cases[c].env})
	}
	for i, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := run("", i)
			if !strings.Contains(r.out, c.wantOut) || !strings.Contains(r.dockerEnv, c.wantEnv) {
				t.Errorf("Enter did not take the default (want %q in the prompt and %q in the lab's settings):\n%s\n%s", c.wantOut, c.wantEnv, r.out, r.dockerEnv)
			}
			if strings.Contains(r.out, "Again, please.") || strings.Contains(r.dockerEnv, "ZETARIS=local") {
				t.Errorf("Enter was refused, or took a non-default:\n%s", r.out)
			}
		})
	}
	// The negative control: Enter is refused again, as the owner measured. Mutated field: the empty-answer branch of
	// flow_ask.
	tree := copyLab(t)
	mutate(t, tree, "lab/scripts/flow.sh", `if [ -z "$fa_ans" ]; then`, `if [ -z "$fa_ans" ] && false; then`)
	m := run(tree, 0)
	if !strings.Contains(m.out, "Again, please.") {
		t.Error("the control did not turn red: with the empty-answer branch off, Enter was still taken as the default")
	}
}
