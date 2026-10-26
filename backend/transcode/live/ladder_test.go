package live

import (
	"reflect"
	"testing"

	"sen1or/letslive/transcode/config"
)

func testQualities() []config.Quality {
	return []config.Quality{
		{Resolution: "1920x1080", MaxBitrate: "5000k", BufSize: "10000k", FPS: 30},
		{Resolution: "1280x720", MaxBitrate: "2800k", BufSize: "5600k", FPS: 30},
		{Resolution: "854x480", MaxBitrate: "1400k", BufSize: "2800k", FPS: 30},
	}
}

func names(rs []Rendition) []string {
	out := make([]string, len(rs))
	for i, r := range rs {
		out[i] = r.Name
	}
	return out
}

func TestBuildLadder(t *testing.T) {
	cases := []struct {
		name string
		src  SourceInfo
		want []string
	}{
		{"1080p landscape", SourceInfo{Width: 1920, Height: 1080, FPS: 30}, []string{"source", "720p", "480p"}},
		{"720p landscape", SourceInfo{Width: 1280, Height: 720, FPS: 30}, []string{"source", "480p"}},
		{"480p landscape equals smallest quality", SourceInfo{Width: 854, Height: 480, FPS: 30}, []string{"source"}},
		{"360x640 portrait", SourceInfo{Width: 360, Height: 640, FPS: 30}, []string{"source"}},
		{"1080x1920 portrait", SourceInfo{Width: 1080, Height: 1920, FPS: 30}, []string{"source", "720p", "480p"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := BuildLadder(tc.src, testQualities())
			if err != nil {
				t.Fatalf("BuildLadder: %v", err)
			}
			if !reflect.DeepEqual(names(got), tc.want) {
				t.Fatalf("names = %v, want %v", names(got), tc.want)
			}
			for i, r := range got {
				if r.Encoded != (i != 0) {
					t.Errorf("rendition %d (%s): Encoded = %v", i, r.Name, r.Encoded)
				}
			}
			if got[0].ShortSide != tc.src.ShortSide() {
				t.Errorf("source ShortSide = %d, want %d", got[0].ShortSide, tc.src.ShortSide())
			}
		})
	}

	t.Run("unsorted input yields descending output", func(t *testing.T) {
		qs := testQualities()
		qs[0], qs[2] = qs[2], qs[0]
		got, err := BuildLadder(SourceInfo{Width: 1920, Height: 1080, FPS: 30}, qs)
		if err != nil {
			t.Fatalf("BuildLadder: %v", err)
		}
		if want := []string{"source", "720p", "480p"}; !reflect.DeepEqual(names(got), want) {
			t.Errorf("names = %v, want %v", names(got), want)
		}
	})

	t.Run("duplicate short side", func(t *testing.T) {
		qs := append(testQualities(), config.Quality{Resolution: "1280x720", FPS: 30})
		_, err := BuildLadder(SourceInfo{Width: 1920, Height: 1080, FPS: 30}, qs)
		if err == nil || err.Error() != "duplicate rendition 720p" {
			t.Errorf("err = %v, want %q", err, "duplicate rendition 720p")
		}
	})

	t.Run("invalid resolution", func(t *testing.T) {
		_, err := BuildLadder(SourceInfo{Width: 1920, Height: 1080}, []config.Quality{{Resolution: "bad"}})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("encoded rendition fields", func(t *testing.T) {
		got, err := BuildLadder(SourceInfo{Width: 1920, Height: 1080, FPS: 30}, testQualities())
		if err != nil {
			t.Fatal(err)
		}
		want := Rendition{Name: "720p", ShortSide: 720, Encoded: true, MaxBitrate: "2800k", BufSize: "5600k", FPS: 30}
		if got[1] != want {
			t.Errorf("720p = %+v, want %+v", got[1], want)
		}
	})

	t.Run("fps", func(t *testing.T) {
		qs := []config.Quality{{Resolution: "1280x720", FPS: 30}}
		cases := []struct {
			name   string
			srcFPS float64
			want   int
		}{
			{"source slower than config", 24, 24},
			{"source faster than config", 60, 30},
			{"fractional source rounds", 29.97, 30},
			{"unknown source fps", 0, 30},
		}
		for _, tc := range cases {
			got, err := BuildLadder(SourceInfo{Width: 1920, Height: 1080, FPS: tc.srcFPS}, qs)
			if err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			if got[1].FPS != tc.want {
				t.Errorf("%s: FPS = %d, want %d", tc.name, got[1].FPS, tc.want)
			}
		}
	})
}
