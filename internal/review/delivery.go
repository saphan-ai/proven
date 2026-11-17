package review

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/saphan-ai/proven/internal/delivery"
	"github.com/saphan-ai/proven/internal/envdoc"
)

const EnvDeliveryFile = "PROVEN_PR_CONTEXT"

func DeliveryEnvironment() []envdoc.Setting {
	return []envdoc.Setting{{Name: EnvDeliveryFile, Default: "empty (local commit descriptions in model-free Git scans)", Description: "external JSON with PR title, description, commit messages and separate comments; checks description completeness without a model and assesses each source against code when a model is connected"}}
}

func loadDelivery(env func(string) string) (*delivery.Metadata, error) {
	if env == nil || env(EnvDeliveryFile) == "" {
		return nil, nil
	}
	refused := func() error {
		return &ConfigError{Field: EnvDeliveryFile, Reason: "expected a regular metadata JSON file of at most 64 KiB", Allowed: []string{"empty", "a PR metadata JSON file outside the reviewed checkout"}}
	}
	info, err := os.Lstat(env(EnvDeliveryFile))
	if err != nil || !info.Mode().IsRegular() || info.Size() > 64<<10 {
		return nil, refused()
	}
	f, err := os.Open(env(EnvDeliveryFile))
	if err != nil {
		return nil, refused()
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, (64<<10)+1))
	if err != nil || len(b) > 64<<10 {
		return nil, refused()
	}
	var metadata delivery.Metadata
	decoder := json.NewDecoder(strings.NewReader(string(b)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&metadata); err != nil {
		return nil, refused()
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return nil, refused()
	}
	if metadata.Context != "" && metadata.Context != "git" && metadata.Context != "pull_request" {
		return nil, refused()
	}
	seen := map[string]bool{}
	for _, c := range metadata.Comments {
		key := fmt.Sprintf("%s:%d", c.Kind, c.ID)
		if !delivery.CommentKind(c.Kind) || c.ID <= 0 || seen[key] {
			return nil, refused()
		}
		seen[key] = true
	}
	return &metadata, nil
}
