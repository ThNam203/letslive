package vod

import (
	"strings"
	"testing"
)

func TestUploadTitle(t *testing.T) {
	tests := []struct {
		name      string
		title     string
		filename  string
		wantTitle string
		wantOK    bool
	}{
		{"typed title is trimmed", "  My stream  ", "a.mp4", "My stream", true},
		{"one character is enough", "x", "a.mp4", "x", true},
		{"255 characters pass", strings.Repeat("a", 255), "a.mp4", strings.Repeat("a", 255), true},
		{"255 multi-byte characters pass", strings.Repeat("ữ", 255), "a.mp4", strings.Repeat("ữ", 255), true},
		{"256 typed characters are rejected", strings.Repeat("a", 256), "a.mp4", strings.Repeat("a", 256), false},
		{"blank title falls back to the file name", "   ", "clip.mp4", "clip.mp4", true},
		{"long file name is cut to 255", "", strings.Repeat("b", 300) + ".mp4", strings.Repeat("b", 255), true},
		{"blank title and blank file name are rejected", "", "  ", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTitle, gotOK := uploadTitle(tt.title, tt.filename)
			if gotOK != tt.wantOK {
				t.Fatalf("ok = %v, want %v", gotOK, tt.wantOK)
			}
			if gotOK && gotTitle != tt.wantTitle {
				t.Fatalf("title = %q, want %q", gotTitle, tt.wantTitle)
			}
		})
	}
}

func TestValidDescription(t *testing.T) {
	tests := []struct {
		name        string
		description string
		want        bool
	}{
		{"empty is allowed", "", true},
		{"1000 characters pass", strings.Repeat("a", 1000), true},
		{"1000 multi-byte characters pass", strings.Repeat("ữ", 1000), true},
		{"1001 characters are rejected", strings.Repeat("a", 1001), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validDescription(tt.description); got != tt.want {
				t.Fatalf("validDescription = %v, want %v", got, tt.want)
			}
		})
	}
}
