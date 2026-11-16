package review

import (
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/saphan-ai/proven/internal/delivery"
	"github.com/saphan-ai/proven/internal/envdoc"
)

const EnvDeliveryFile = "PROVEN_PR_CONTEXT"

func DeliveryEnvironment() []envdoc.Setting {
	return []envdoc.Setting{{Name: EnvDeliveryFile, Default: "empty (disabled)", Description: "external JSON file with PR title, description and commit messages; enables advisory delivery assessment with validated citations, separate from proved findings"}}
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
	return &metadata, nil
}
