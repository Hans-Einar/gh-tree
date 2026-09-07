package broker

import (
	"errors"
	"io"
	"strings"
	"testing"
)

func TestFreeBSDCensusBoundCannotBypassWriter(t *testing.T) {
	output := &censusWriter{limit: 16}
	// Hide source WriterTo so io.Copy would select a promoted destination
	// ReaderFrom if embedding ever accidentally bypasses the byte ceiling.
	source := struct{ io.Reader }{strings.NewReader(strings.Repeat("x", 32))}
	if _, err := io.Copy(output, source); !errors.Is(err, ErrCensus) || len(output.String()) > 16 {
		t.Fatal("observer output exceeded bound", err)
	}
}
