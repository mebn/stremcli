package cli

import "testing"

func TestParseTime(t *testing.T) {
	for in, want := range map[string]float64{"5025": 5025, "83:45": 5025, "1:23:45": 5025, "0:30": 30} {
		if got, err := parseTime(in); err != nil || got != want {
			t.Errorf("parseTime(%q) = %v, %v; want %v", in, got, err, want)
		}
	}
	for _, in := range []string{"", "abc", "1:-2", "1::2"} {
		if _, err := parseTime(in); err == nil {
			t.Errorf("parseTime(%q) succeeded, want error", in)
		}
	}
}
