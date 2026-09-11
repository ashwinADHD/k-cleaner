package schedule

import "testing"

func TestParseInterval(t *testing.T) {
	if ParseInterval("daily") != Daily {
		t.Error("expected daily")
	}
	if ParseInterval("weekly") != Weekly {
		t.Error("expected weekly")
	}
	if ParseInterval("") != Weekly {
		t.Error("expected default weekly")
	}
}
