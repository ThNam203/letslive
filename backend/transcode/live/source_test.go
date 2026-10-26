package live

import (
	"math"
	"testing"

	"github.com/nareix/joy5/av"
)

func firstDecoderConfig(t *testing.T, name string) []byte {
	t.Helper()
	for _, pkt := range readFLV(t, name) {
		if pkt.Type == av.H264DecoderConfig {
			return pkt.Data
		}
	}
	t.Fatalf("%s has no H264DecoderConfig packet", name)
	return nil
}

func TestProbeAVCConfig(t *testing.T) {
	t.Run("landscape 640x360", func(t *testing.T) {
		info, err := ProbeAVCConfig(firstDecoderConfig(t, "src_640x360.flv"))
		if err != nil {
			t.Fatalf("ProbeAVCConfig: %v", err)
		}
		if info.Width != 640 || info.Height != 360 {
			t.Errorf("size = %dx%d, want 640x360", info.Width, info.Height)
		}
		if math.Abs(info.FPS-30) > 0.01 {
			t.Errorf("FPS = %v, want 30", info.FPS)
		}
		if len(info.SPS) == 0 || len(info.PPS) == 0 {
			t.Errorf("SPS len %d, PPS len %d, want both non-empty", len(info.SPS), len(info.PPS))
		}
		if info.Portrait() {
			t.Error("Portrait() = true for 640x360")
		}
		if info.ShortSide() != 360 {
			t.Errorf("ShortSide() = %d, want 360", info.ShortSide())
		}
	})

	t.Run("portrait 360x640", func(t *testing.T) {
		info, err := ProbeAVCConfig(firstDecoderConfig(t, "src_360x640.flv"))
		if err != nil {
			t.Fatalf("ProbeAVCConfig: %v", err)
		}
		if !info.Portrait() {
			t.Errorf("Portrait() = false for %dx%d", info.Width, info.Height)
		}
		if info.ShortSide() != 360 {
			t.Errorf("ShortSide() = %d, want 360", info.ShortSide())
		}
	})

	t.Run("malformed", func(t *testing.T) {
		valid := firstDecoderConfig(t, "src_640x360.flv")
		cases := map[string][]byte{
			"nil":              nil,
			"too short":        {1, 2},
			"wrong version":    append([]byte{2}, valid[1:]...),
			"truncated in SPS": valid[:9],
			"no PPS":           valid[:8+(int(valid[6])<<8|int(valid[7]))],
		}
		for name, in := range cases {
			if _, err := ProbeAVCConfig(in); err == nil {
				t.Errorf("%s: expected error", name)
			}
		}
	})
}
