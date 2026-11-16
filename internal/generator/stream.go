package generator

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/saphan-ai/proven/internal/modelcall"
	"github.com/saphan-ai/proven/internal/service"
)

// maxStream is the most of one completion body this generator will read.
// A longer body fails as a bad stream instead of growing without a bound.
const maxStream = 8 << 20

// errBadStream is a stream with an event that is not JSON. Complete names it
// for the provider it was reading.
var errBadStream = errors.New("bad stream")
var errProviderStream = errors.New("provider stream error")

// streamMeta is what one completion stream carried: the answer text and,
// from Nightshift's receipt, the model and route that answered. chunkModel is
// the model field of the chunks themselves, which only nvidia mode reads.
type streamMeta struct {
	text       string
	model      string
	receipt    string
	route      string
	chunkModel string
	// responseID is the id of the chat-completion response. Nightshift's
	// receipt names the same id as its request id, so it is the key that ties
	// an answer to its receipt.
	responseID       string
	finishReason     string
	reasoningBytes   int
	reasoning        string
	reasoningClipped bool
	usage            *modelcall.Usage
}

// sseChunk is one data event of the stream. Only the fields the review reads
// are decoded; the rest of the event is ignored.
type sseChunk struct {
	// Model is the chunk's own model field. Nightshift's is not a receipt
	// ; the hosted API's is the only statement of who answered.
	Model string `json:"model"`
	// ID is the response id; every chunk of one response carries the same one.
	ID      string `json:"id"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Content          string          `json:"content"`
			ReasoningContent json.RawMessage `json:"reasoning_content"`
			Reasoning        json.RawMessage `json:"reasoning"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage     json.RawMessage `json:"usage"`
	Error     json.RawMessage `json:"error"`
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
	return parseStreamOptions(r, 0)
}

func parseStreamOptions(r io.Reader, reasoningLimit int) (meta streamMeta, resultErr error) {
	lr := &io.LimitedReader{R: r, N: maxStream + 1}
	sc := bufio.NewScanner(lr)
	// One event line may be up to 1 MiB; a longer line ends the scan with an error.
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20)
	var text strings.Builder
	defer func() { meta.text = text.String() }()
	for sc.Scan() {
		line := strings.TrimSpace(strings.TrimRight(sc.Text(), "\r"))
		// Blank lines, comment lines and non-data fields carry no part of the answer.
		if line == "" || strings.HasPrefix(line, ":") || !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			break
		}
		var chunk sseChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			return meta, errBadStream
		}
		if len(chunk.Error) > 0 && string(chunk.Error) != "null" {
			return meta, errProviderStream
		}
		for _, c := range chunk.Choices {
			if c.Index != 0 {
				continue
			}
			text.WriteString(c.Delta.Content)
			if c.FinishReason != "" {
				meta.finishReason = c.FinishReason
			}
			// Some endpoints expose both aliases. Count and retain one, never add it to the answer.
			reasoning := reasoningText(c.Delta.ReasoningContent)
			if reasoning == "" {
				reasoning = reasoningText(c.Delta.Reasoning)
			}
			meta.reasoningBytes += len(reasoning)
			if left := reasoningLimit - len(meta.reasoning); left > 0 && !meta.reasoningClipped {
				part := reasoning
				if len(part) > left {
					part = part[:left]
					for !utf8.ValidString(part) {
						part = part[:len(part)-1]
					}
				}
				meta.reasoning += part
			}
			if reasoningLimit > 0 && meta.reasoningBytes > len(meta.reasoning) {
				meta.reasoningClipped = true
			}
		}
		if u := readUsage(chunk.Usage); u != nil {
			meta.usage = u
		}
		// The first chunk that names a model wins: a later one cannot change who answered.
		if meta.chunkModel == "" {
			meta.chunkModel = chunk.Model
		}
		// The first id wins for the same reason: a later chunk cannot change which response this is.
		if meta.responseID == "" {
			meta.responseID = chunk.ID
		}
		// in nightshift mode the answering model is read from the receipt, never from the
		// chunk's own model field, so a chunk without the meterless object adds text only.
		if chunk.Meterless == nil {
			continue
		}
		// The receipt may arrive across several chunks; a later empty field does not
		// erase an earlier value.
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
	if lr.N == 0 {
		return meta, errBadStream
	}
	if err := sc.Err(); err != nil {
		return meta, err
	}
	return meta, nil
}

func reasoningText(raw json.RawMessage) string {
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return text
	}
	var parts []struct {
		Text string `json:"text"`
	}
	if json.Unmarshal(raw, &parts) == nil {
		var b strings.Builder
		for _, p := range parts {
			b.WriteString(p.Text)
		}
		return b.String()
	}
	return ""
}

func readUsage(raw json.RawMessage) *modelcall.Usage {
	var u struct {
		Prompt     *int64 `json:"prompt_tokens"`
		Completion *int64 `json:"completion_tokens"`
		Total      *int64 `json:"total_tokens"`
		Details    struct {
			Reasoning *int64 `json:"reasoning_tokens"`
		} `json:"completion_tokens_details"`
	}
	if len(raw) == 0 || json.Unmarshal(raw, &u) != nil {
		return nil
	}
	if u.Prompt == nil && u.Completion == nil && u.Total == nil && u.Details.Reasoning == nil {
		return nil
	}
	for _, v := range []*int64{u.Prompt, u.Completion, u.Total, u.Details.Reasoning} {
		if v != nil && *v < 0 {
			return nil
		}
	}
	return &modelcall.Usage{PromptTokens: u.Prompt, CompletionTokens: u.Completion, TotalTokens: u.Total, ReasoningTokens: u.Details.Reasoning}
}

// discard keeps a non-success body from pinning the connection.
func discard(r io.Reader) {
	_, _ = io.Copy(io.Discard, io.LimitReader(r, 4096))
}

// down is a named failure of the provider these settings choose. The review
// degrades and continues.
func (l *Live) down(reason string) *service.Down {
	if l == nil {
		return &service.Down{Service: "generator", Reason: reason}
	}
	return &service.Down{Service: l.cfg.serviceName(), Reason: reason}
}
