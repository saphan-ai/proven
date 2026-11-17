package change

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
)

type CommitMessage struct{ SHA, Message string }

// CommitMessages keeps a stable identity and empty messages. Ordinary Messages
// retains its existing trailer-reading contract. Both use the same Git policy
// and bounded content reads, without running repository hooks or code.
func (p Policy) CommitMessages(target, baseline, head string) ([]CommitMessage, error) {
	if err := p.ensureCommit(target, baseline, "baseline"); err != nil {
		return nil, err
	}
	if head == "" {
		head = "HEAD"
	} else if err := p.ensureCommit(target, head, "head"); err != nil {
		return nil, err
	}
	cmd := p.GitCommand("-C", target, "log", "--reverse", "--format=%H%x00%B%x00", "--end-of-options", baseline+".."+head)
	b, err := p.outputBounded(cmd, p.Bounds().File, "the commit log", EnvMaxFileBytes)
	if err != nil {
		return nil, err
	}
	parts := bytes.Split(b, []byte{0})
	var out []CommitMessage
	for i := 0; i+1 < len(parts); i += 2 {
		sha := strings.TrimSpace(string(parts[i]))
		if _, err := hex.DecodeString(sha); err != nil || len(sha) != 40 && len(sha) != 64 {
			return nil, fmt.Errorf("git: invalid commit identity in bounded log")
		}
		out = append(out, CommitMessage{SHA: sha, Message: strings.TrimSpace(string(parts[i+1]))})
	}
	return out, nil
}
