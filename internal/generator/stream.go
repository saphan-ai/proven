package generator

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/saphan-ai/proven/internal/service"
)

// maxStream is the most of one completion body this generator will read.
// A longer body fails as a bad stream instead of growing without a bound.
const maxStream = 8 << 20

// errBadStream is a stream with an event that is not JSON. Complete names it
// for the provider it was reading.
var errBadStream = errors.New("bad stream")

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
	responseID string
}

// sseChunk is one data event of the stream. Only the fields the review reads
// are decoded; the rest of the event is ignored.
type sseChunk struct {
	// Model is the chunk's own model field. Nightshift's is not a receipt
	// (ADR-0019); the hosted API's is the only statement of who answered.
	Model string `json:"model"`
	// ID is the response id; every chunk of one response carries the same one.
	ID      string `json:"id"`
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
	// One event line may be up to 1 MiB; a longer line ends the scan with an error.
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20)
	var text strings.Builder
	var meta streamMeta
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
			return streamMeta{}, errBadStream
		}
		for _, c := range chunk.Choices {
			text.WriteString(c.Delta.Content)
		}
		// The first chunk that names a model wins: a later one cannot change who answered.
		if meta.chunkModel == "" {
			meta.chunkModel = chunk.Model
		}
		// The first id wins for the same reason: a later chunk cannot change which response this is.
		if meta.responseID == "" {
			meta.responseID = chunk.ID
		}
		// ADR-0019: in nightshift mode the answering model is read from the receipt, never from the
		// chunk's own model field, so a chunk without the meterless object adds text only.
		if chunk.Meterless == nil {
			continue
		}
		// The receipt may arrive across several chunks; a later empty field does not
		// erase an earlier value. The routed model is the one that answered, which a
		// router may choose instead of the model the request named.
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

// down is a named failure of the provider these settings choose. The review
// degrades and continues.
func (l *Live) down(reason string) *service.Down {
	return &service.Down{Service: l.cfg.serviceName(), Reason: reason}
}
