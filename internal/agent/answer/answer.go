// Package answer reads a model reply into the JSON an agent asked for.
//
// It is shared by the model-calling agents (Planner, Investigator, Prover and
// Skeptic): each states a Spec, sends it through Instruct, and reads the reply
// with Read. A reply that is not the asked-for JSON gets one repair, then the
// step degrades by the agent's name rather than failing the review.
package answer

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/service"
)

// FailReason is the degraded outcome when a model answer still does not
// follow the schema after one repair.
const FailReason = "the model's answer did not follow the schema"

// Spec is the JSON one agent asks for.
type Spec struct {
	Task    string
	Shape   string
	Example string
}

// Instruct is the live prompt: the task, the JSON, one example, and the
// change-specific data. It asks for that JSON only.
func Instruct(spec Spec, data string) string {
	var b strings.Builder
	b.WriteString("Task: ")
	b.WriteString(spec.Task)
	b.WriteString("\n\nReply with JSON only.\n\nThe JSON is ")
	b.WriteString(spec.Shape)
	b.WriteString("\n\nExample:\n")
	b.WriteString(spec.Example)
	if data != "" {
		b.WriteString("\n\n")
		b.WriteString(data)
	}
	return b.String()
}

// Repair asks the model once to make its previous answer match the schema.
func Repair(spec Spec, previous, original string) string {
	var b strings.Builder
	b.WriteString("The previous answer did not follow the schema. Reply with JSON only.\n\n")
	b.WriteString("Task: ")
	b.WriteString(spec.Task)
	b.WriteString("\n\nThe JSON is ")
	b.WriteString(spec.Shape)
	b.WriteString("\n\nExample:\n")
	b.WriteString(spec.Example)
	b.WriteString("\n\nPrevious answer:\n")
	b.WriteString(previous)
	b.WriteString("\n\nOriginal request:\n")
	b.WriteString(original)
	return b.String()
}

// Read asks the generator and decodes the answer into dst.
// A decode failure is sent back once as a repair. A second failure is a
// named degraded outcome for agent. A generator error is returned as it is.
func Read(ctx context.Context, gen generator.Generator, agent, prompt string, spec Spec, dst any) error {
	raw, err := gen.Complete(ctx, agent, prompt)
	if err != nil {
		return err
	}
	// A reply that decodes is used as it is, with no second call.
	if err := Decode(raw, dst); err == nil {
		return nil
	}
	again, err := gen.Complete(ctx, agent, Repair(spec, raw, prompt))
	if err != nil {
		return err
	}
	if err := Decode(again, dst); err != nil {
		// The failure is named for the agent, not for the model service: the model
		// answered, but not in the form asked for.
		return &service.Down{Service: agent, Reason: FailReason}
	}
	return nil
}

// Decode unmarshals JSON taken from a fenced code block or from surrounding
// prose. A trailing comma is accepted.
func Decode(raw string, dst any) error {
	// With no bracketed value found, the whole reply is tried, so the decode
	// error names what was wrong with it.
	cands := candidates(raw)
	if len(cands) == 0 {
		cands = []string{strings.TrimSpace(raw)}
	}
	var last error
	for _, c := range cands {
		// Each try starts from a zero value, so a partial decode of one
		// candidate cannot leave fields behind for the next.
		reset(dst)
		err := json.Unmarshal([]byte(stripTrailingCommas(c)), dst)
		if err == nil {
			return nil
		}
		last = err
	}
	reset(dst)
	if last == nil {
		last = fmt.Errorf("answer: no json")
	}
	return last
}

// candidates lists the JSON values in raw, in order. When raw has a fenced
// code block, only the first block is searched: prose around it may hold
// example JSON that is not the answer.
func candidates(raw string) []string {
	text := raw
	if body, ok := firstFence(raw); ok {
		text = body
	}
	return jsonValues(text)
}

// firstFence returns the body of the first ``` block, after its language line.
func firstFence(s string) (string, bool) {
	i := strings.Index(s, "```")
	if i < 0 {
		return "", false
	}
	rest := s[i+3:]
	nl := strings.IndexByte(rest, '\n')
	if nl < 0 {
		return "", false
	}
	rest = rest[nl+1:]
	j := strings.Index(rest, "```")
	if j < 0 {
		return "", false
	}
	return rest[:j], true
}

// jsonValues finds each balanced {...} or [...] value in s. After a match the
// scan resumes past it, so a nested value is not listed twice.
func jsonValues(s string) []string {
	var out []string
	for i := 0; i < len(s); i++ {
		if s[i] != '{' && s[i] != '[' {
			continue
		}
		v, n, ok := matchValue(s, i)
		if !ok {
			continue
		}
		out = append(out, v)
		if n <= i {
			continue
		}
		i = n - 1
	}
	return out
}

// matchValue finds the bracket that closes the one at start. Brackets inside
// strings, escaped quotes included, do not count. It returns the value and the
// index just past it.
func matchValue(s string, start int) (string, int, bool) {
	open := s[start]
	close := byte('}')
	if open == '[' {
		close = ']'
	}
	depth := 0
	inStr := false
	esc := false
	for i := start; i < len(s); i++ {
		c := s[i]
		if inStr {
			if esc {
				esc = false
				continue
			}
			if c == '\\' {
				esc = true
				continue
			}
			if c == '"' {
				inStr = false
			}
			continue
		}
		switch c {
		case '"':
			inStr = true
		case open:
			depth++
		case close:
			depth--
			if depth == 0 {
				return s[start : i+1], i + 1, true
			}
		}
	}
	return "", start, false
}

// stripTrailingCommas removes a comma that comes right before a closing bracket,
// a common model slip. A comma inside a string is kept.
func stripTrailingCommas(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	inStr := false
	esc := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if inStr {
			b.WriteByte(c)
			if esc {
				esc = false
				continue
			}
			if c == '\\' {
				esc = true
				continue
			}
			if c == '"' {
				inStr = false
			}
			continue
		}
		if c == '"' {
			inStr = true
			b.WriteByte(c)
			continue
		}
		if c == ',' && commaBeforeClose(s, i+1) {
			continue
		}
		b.WriteByte(c)
	}
	return b.String()
}

// commaBeforeClose reports whether only white space lies between i and a closing bracket.
func commaBeforeClose(s string, i int) bool {
	for i < len(s) {
		switch s[i] {
		case ' ', '\n', '\r', '\t':
			i++
		case '}', ']':
			return true
		default:
			return false
		}
	}
	return false
}

// reset sets what dst points to back to its zero value.
func reset(dst any) {
	v := reflect.ValueOf(dst)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return
	}
	v.Elem().Set(reflect.Zero(v.Elem().Type()))
}
