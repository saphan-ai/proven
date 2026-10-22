package generator

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"

	"github.com/saphan-ai/proven/internal/service"
)

// maxStream is the most of one completion body this generator will read.
// A longer body fails as a bad stream instead of growing without a bound.
const maxStream = 8 << 20

type streamMeta struct {
	text    string
	model   string
	receipt string
	route   string
}

type sseChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
	Meterless *struct {
		ReceiptID   string `json:"receipt_id"`
		RoutedModel string `json:"routed_model"`
		Winning     *struct {
			Provider string `json:"provider"`
		} `json:"winning_route"`
	} `json:"meterless"`
}

// parseStream reads an OpenAI-compatible server-sent stream to the end.
// The routed model and the receipt id are taken only from a top-level
// meterless object. The chunk's model field is not a receipt.
func parseStream(r io.Reader) (streamMeta, error) {
	sc := bufio.NewScanner(io.LimitReader(r, maxStream))
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20)
	var text strings.Builder
	var meta streamMeta
	for sc.Scan() {
		line := strings.TrimSpace(strings.TrimRight(sc.Text(), "\r"))
		if line == "" || strings.HasPrefix(line, ":") || !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			break
		}
		var chunk sseChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			return streamMeta{}, down("bad stream")
		}
		for _, c := range chunk.Choices {
			text.WriteString(c.Delta.Content)
		}
		if chunk.Meterless == nil {
			continue
		}
		if chunk.Meterless.RoutedModel != "" {
			meta.model = chunk.Meterless.RoutedModel
		}
		if chunk.Meterless.ReceiptID != "" {
			meta.receipt = chunk.Meterless.ReceiptID
		}
		if chunk.Meterless.Winning != nil && chunk.Meterless.Winning.Provider != "" {
			meta.route = chunk.Meterless.Winning.Provider
		}
	}
	if err := sc.Err(); err != nil {
		return streamMeta{}, err
	}
	meta.text = text.String()
	return meta, nil
}

// discard keeps a non-success body from pinning the connection.
func discard(r io.Reader) {
	_, _ = io.Copy(io.Discard, io.LimitReader(r, 4096))
}

// down is a named Nightshift failure. The review degrades and continues.
func down(reason string) *service.Down {
	return &service.Down{Service: "nightshift", Reason: reason}
}
