package live

import (
	"strings"
	"testing"
)

// gohlslibLLVideoAudio is gohlslib v2.4.7 muxer_test.go's expected
// multivariant playlist for content "video+audio", variant "lowLatency".
const gohlslibLLVideoAudio = "#EXTM3U\n" +
	"#EXT-X-VERSION:10\n" +
	"#EXT-X-INDEPENDENT-SEGMENTS\n" +
	"\n" +
	"#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID=\"audio\"," +
	"NAME=\"audio2\",AUTOSELECT=YES,DEFAULT=YES,URI=\"audio2_stream.m3u8?key=value\"\n" +
	"\n" +
	"#EXT-X-STREAM-INF:BANDWIDTH=872,AVERAGE-BANDWIDTH=584,CODECS=\"avc1.42c028,mp4a.40.2\"," +
	"RESOLUTION=1920x1080,FRAME-RATE=30.000,AUDIO=\"audio\"\n" +
	"video1_stream.m3u8?key=value\n"

// gohlslibLLVideo is the same file's expectation for content "video",
// variant "lowLatency".
const gohlslibLLVideo = "#EXTM3U\n" +
	"#EXT-X-VERSION:10\n" +
	"#EXT-X-INDEPENDENT-SEGMENTS\n" +
	"\n" +
	"#EXT-X-STREAM-INF:BANDWIDTH=872,AVERAGE-BANDWIDTH=403,CODECS=\"avc1.42c028\"," +
	"RESOLUTION=1920x1080,FRAME-RATE=30.000\n" +
	"video1_stream.m3u8?key=value\n"

func TestMergeMultivariant(t *testing.T) {
	got, err := MergeMultivariant([]NamedPlaylist{
		{Name: "source", Body: []byte(gohlslibLLVideoAudio)},
		{Name: "480p", Body: []byte(gohlslibLLVideoAudio)},
	})
	if err != nil {
		t.Fatalf("MergeMultivariant: %v", err)
	}

	want := "#EXTM3U\n" +
		"#EXT-X-VERSION:10\n" +
		"#EXT-X-INDEPENDENT-SEGMENTS\n" +
		"\n" +
		"#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID=\"audio-source\"," +
		"NAME=\"audio2\",AUTOSELECT=YES,DEFAULT=YES,URI=\"source/audio2_stream.m3u8?key=value\"\n" +
		"#EXT-X-STREAM-INF:BANDWIDTH=872,AVERAGE-BANDWIDTH=584,CODECS=\"avc1.42c028,mp4a.40.2\"," +
		"RESOLUTION=1920x1080,FRAME-RATE=30.000,AUDIO=\"audio-source\"\n" +
		"source/video1_stream.m3u8?key=value\n" +
		"\n" +
		"#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID=\"audio-480p\"," +
		"NAME=\"audio2\",AUTOSELECT=YES,DEFAULT=YES,URI=\"480p/audio2_stream.m3u8?key=value\"\n" +
		"#EXT-X-STREAM-INF:BANDWIDTH=872,AVERAGE-BANDWIDTH=584,CODECS=\"avc1.42c028,mp4a.40.2\"," +
		"RESOLUTION=1920x1080,FRAME-RATE=30.000,AUDIO=\"audio-480p\"\n" +
		"480p/video1_stream.m3u8?key=value\n"

	if string(got) != want {
		t.Errorf("merged playlist mismatch\n got:\n%s\nwant:\n%s", got, want)
	}
}

func TestMergeMultivariantVideoOnly(t *testing.T) {
	got, err := MergeMultivariant([]NamedPlaylist{
		{Name: "source", Body: []byte(gohlslibLLVideo)},
		{Name: "480p", Body: []byte(gohlslibLLVideo)},
	})
	if err != nil {
		t.Fatalf("MergeMultivariant: %v", err)
	}

	s := string(got)
	if strings.Contains(s, "EXT-X-MEDIA") {
		t.Errorf("video-only merge has EXT-X-MEDIA:\n%s", s)
	}
	if strings.Contains(s, "AUDIO=") {
		t.Errorf("video-only merge references an audio group:\n%s", s)
	}
	for _, line := range []string{"source/video1_stream.m3u8?key=value", "480p/video1_stream.m3u8?key=value"} {
		if !strings.Contains(s, "\n"+line+"\n") {
			t.Errorf("merged playlist lacks %q:\n%s", line, s)
		}
	}
	if n := strings.Count(s, "#EXTM3U"); n != 1 {
		t.Errorf("#EXTM3U count = %d, want 1", n)
	}
}

func TestMergeMultivariantHeader(t *testing.T) {
	older := strings.Replace(gohlslibLLVideo, "#EXT-X-VERSION:10", "#EXT-X-VERSION:7", 1)
	got, err := MergeMultivariant([]NamedPlaylist{
		{Name: "source", Body: []byte(older)},
		{Name: "720p", Body: []byte(gohlslibLLVideo)},
		{Name: "480p", Body: []byte(older)},
	})
	if err != nil {
		t.Fatalf("MergeMultivariant: %v", err)
	}

	s := string(got)
	if !strings.HasPrefix(s, "#EXTM3U\n#EXT-X-VERSION:10\n#EXT-X-INDEPENDENT-SEGMENTS\n") {
		t.Errorf("header is not EXTM3U, highest version, one INDEPENDENT-SEGMENTS:\n%s", s)
	}
	for _, tag := range []string{"#EXT-X-VERSION", "#EXT-X-INDEPENDENT-SEGMENTS"} {
		if n := strings.Count(s, tag); n != 1 {
			t.Errorf("%s count = %d, want 1", tag, n)
		}
	}
	src := strings.Index(s, "source/")
	mid := strings.Index(s, "720p/")
	low := strings.Index(s, "480p/")
	if src < 0 || mid < src || low < mid {
		t.Errorf("input order not kept (source %d, 720p %d, 480p %d):\n%s", src, mid, low, s)
	}
}

func TestMergeMultivariantErrors(t *testing.T) {
	cases := map[string][]NamedPlaylist{
		"no parts":        nil,
		"bad name":        {{Name: "../x", Body: []byte(gohlslibLLVideo)}},
		"empty name":      {{Name: "", Body: []byte(gohlslibLLVideo)}},
		"duplicate name":  {{Name: "source", Body: []byte(gohlslibLLVideo)}, {Name: "source", Body: []byte(gohlslibLLVideo)}},
		"missing EXTM3U":  {{Name: "source", Body: []byte(strings.TrimPrefix(gohlslibLLVideo, "#EXTM3U\n"))}},
		"bad version":     {{Name: "source", Body: []byte(strings.Replace(gohlslibLLVideo, ":10", ":x", 1))}},
		"stray URI line":  {{Name: "source", Body: []byte("#EXTM3U\nvideo1_stream.m3u8\n")}},
		"no variant":      {{Name: "source", Body: []byte("#EXTM3U\n#EXT-X-VERSION:10\n")}},
		"dangling STREAM": {{Name: "source", Body: []byte("#EXTM3U\n#EXT-X-STREAM-INF:BANDWIDTH=1\n")}},
	}

	for name, parts := range cases {
		t.Run(name, func(t *testing.T) {
			if got, err := MergeMultivariant(parts); err == nil {
				t.Errorf("MergeMultivariant succeeded:\n%s", got)
			}
		})
	}
}
