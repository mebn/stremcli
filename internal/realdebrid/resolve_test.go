package realdebrid

import "testing"

func TestPickFile(t *testing.T) {
	files := []File{
		{ID: 1, Path: "/Show/S01E01.mkv", Bytes: 900},
		{ID: 2, Path: "/Show/S01E02.mkv", Bytes: 800},
		{ID: 3, Path: "/Show/sample.txt", Bytes: 5000},
	}

	tests := []struct {
		name   string
		target Target
		want   int
	}{
		{"by filename", Target{Filename: "s01e02.MKV", FileIdx: -1}, 2},
		{"by index", Target{FileIdx: 1}, 2},
		{"index points at non-video", Target{FileIdx: 2}, 1},
		{"largest video fallback", Target{FileIdx: -1}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := pickFile(files, tt.target)
			if !ok || got.ID != tt.want {
				t.Fatalf("pickFile() = %d, %v; want %d", got.ID, ok, tt.want)
			}
		})
	}
}
