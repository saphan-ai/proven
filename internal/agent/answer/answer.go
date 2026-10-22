// Package answer reads a model reply into the JSON an agent asked for.
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
	if err := Decode(raw, dst); err == nil {
		return nil
	}
	again, err := gen.Complete(ctx, agent, Repair(spec, raw, prompt))
	if err != nil {
		return err
	}
	if err := Decode(again, dst); err != nil {
		return &service.Down{Service: agent, Reason: FailReason}
	}
	return nil
}

// Decode unmarshals JSON taken from a fenced code block or from surrounding
// prose. A trailing comma is accepted.
func Decode(raw string, dst any) error {
	cands := candidates(raw)
	if len(cands) == 0 {
		cands = []string{strings.TrimSpace(raw)}
	}
	var last error
	for _, c := range cands {
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

func candidates(raw string) []string {
	text := raw
	if body, ok := firstFence(raw); ok {
		text = body
	}
	return jsonValues(text)
}

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

func reset(dst any) {
	v := reflect.ValueOf(dst)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return
	}
	v.Elem().Set(reflect.Zero(v.Elem().Type()))
}
