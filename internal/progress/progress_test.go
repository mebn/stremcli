package progress

import "testing"

func TestFinished(t *testing.T) {
	tests := []struct {
		pos, dur float64
		want     bool
	}{
		{0, 0, false},
		{2600, 2700, true},  // 45 min episode, 100s left
		{2500, 2700, false}, // 200s left
		{6900, 7200, true},  // 2h movie, 5 min left
		{6000, 7200, false}, // 20 min left
		{1500, 1500, true},  // played to the end
	}
	for _, tt := range tests {
		if got := Finished(tt.pos, tt.dur); got != tt.want {
			t.Errorf("Finished(%v, %v) = %v, want %v", tt.pos, tt.dur, got, tt.want)
		}
	}
}

func TestFormatTime(t *testing.T) {
	for sec, want := range map[float64]string{5: "0:05", 754.9: "12:34", 3725: "1:02:05"} {
		if got := FormatTime(sec); got != want {
			t.Errorf("FormatTime(%v) = %q, want %q", sec, got, want)
		}
	}
}
