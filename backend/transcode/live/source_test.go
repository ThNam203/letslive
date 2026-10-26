package live

import (
	"math"
	"testing"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h264"
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

// baseSPS is a 4:2:0 progressive 640x360-ish SPS (40x23 macroblocks, 368 rows) before cropping.
func baseSPS() h264.SPS {
	return h264.SPS{
		ChromaFormatIdc:           1,
		FrameMbsOnlyFlag:          true,
		PicWidthInMbsMinus1:       39,
		PicHeightInMapUnitsMinus1: 22,
	}
}

func withTiming(sps h264.SPS, numUnitsInTick, timeScale uint32) h264.SPS {
	sps.VUI = &h264.SPS_VUI{TimingInfo: &h264.SPS_TimingInfo{NumUnitsInTick: numUnitsInTick, TimeScale: timeScale}}
	return sps
}

func TestSourceFromSPSRejectsInvalidDimensions(t *testing.T) {
	crop := func(left, right, top, bottom uint32) h264.SPS {
		sps := baseSPS()
		sps.FrameCropping = &h264.SPS_FrameCropping{LeftOffset: left, RightOffset: right, TopOffset: top, BottomOffset: bottom}
		return sps
	}
	oversized := baseSPS()
	oversized.PicWidthInMbsMinus1 = 1000
	tall := baseSPS()
	tall.PicHeightInMapUnitsMinus1 = 1000

	cases := map[string]h264.SPS{
		"crop equals width":   crop(160, 160, 0, 0),
		"crop exceeds width":  crop(400, 400, 0, 0),
		"crop equals height":  crop(0, 0, 92, 92),
		"crop exceeds height": crop(0, 0, 500, 500),
		"width above cap":     oversized,
		"height above cap":    tall,
	}
	for name, sps := range cases {
		if _, _, _, err := sourceFromSPS(sps); err == nil {
			t.Errorf("%s: expected error, got none", name)
		}
	}

	w, h, _, err := sourceFromSPS(baseSPS())
	if err != nil {
		t.Fatalf("valid SPS: %v", err)
	}
	if w != 640 || h != 368 {
		t.Errorf("size = %dx%d, want 640x368", w, h)
	}
}

func TestSourceFromSPSFPS(t *testing.T) {
	cases := []struct {
		name string
		sps  h264.SPS
		want float64
	}{
		{"no VUI", baseSPS(), 0},
		{"valid 30fps", withTiming(baseSPS(), 1, 60), 30},
		{"zero ticks gives +Inf", withTiming(baseSPS(), 0, 60), 0},
		{"zero ticks and scale gives NaN", withTiming(baseSPS(), 0, 0), 0},
		{"zero time scale", withTiming(baseSPS(), 1, 0), 0},
	}
	for _, tc := range cases {
		_, _, fps, err := sourceFromSPS(tc.sps)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if fps != tc.want {
			t.Errorf("%s: fps = %v, want %v", tc.name, fps, tc.want)
		}
	}
}
