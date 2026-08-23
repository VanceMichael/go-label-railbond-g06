package query_test

import (
	"github.com/VanceMichael/go-base-railbond-g06/internal/query"
	"testing"
)

func TestCursorRoundTrip(t *testing.T) {
	in := query.Cursor{LastID: "a-1", Snapshot: "snapshot"}
	out, err := query.DecodeCursor(query.EncodeCursor(in))
	if err != nil || out != in {
		t.Fatalf("%v %#v", err, out)
	}
}
func TestCursorRejectsMalformed(t *testing.T) {
	if _, err := query.DecodeCursor("bad"); err == nil {
		t.Fatal("malformed cursor accepted")
	}
	if query.ParseLimit("999", 20, 100) != 100 {
		t.Fatal("limit not capped")
	}
}
