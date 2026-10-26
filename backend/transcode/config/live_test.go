package config

import (
	"strings"
	"testing"
)

func TestLiveWithDefaults(t *testing.T) {
	got := Live{}.WithDefaults()
	want := Live{
		PartMinDurationMs:       200,
		SegmentMinDurationMs:    1000,
		SegmentCount:            7,
		RecordSegmentDurationMs: 4000,
		EncoderQueuePackets:     512,
		ThumbnailIntervalMs:     60000,
	}
	if got != want {
		t.Fatalf("WithDefaults() = %+v, want %+v", got, want)
	}

	if got := (Live{PartMinDurationMs: 300}).WithDefaults().PartMinDurationMs; got != 300 {
		t.Fatalf("PartMinDurationMs = %d, want 300", got)
	}
}

func TestQualitySize(t *testing.T) {
	w, h, err := Quality{Resolution: "1280x720"}.Size()
	if err != nil || w != 1280 || h != 720 {
		t.Fatalf("Size() = %d, %d, %v; want 1280, 720, nil", w, h, err)
	}

	for _, res := range []string{"hd720", "1280", "0x720"} {
		if _, _, err := (Quality{Resolution: res}).Size(); err == nil {
			t.Errorf("Size(%q) returned no error", res)
		}
	}
}

func TestValidate(t *testing.T) {
	valid := func() Transcode {
		var tc Transcode
		tc.FFMpegSetting.Qualities = []Quality{{Resolution: "1280x720"}}
		tc.Live = Live{}.WithDefaults()
		return tc
	}

	if err := Validate(valid()); err != nil {
		t.Fatalf("Validate(valid) = %v", err)
	}

	lowCount := valid()
	lowCount.Live.SegmentCount = 6
	err := Validate(lowCount)
	if err == nil || !strings.Contains(err.Error(), "segmentCount must be >= 7") {
		t.Fatalf("Validate(segmentCount=6) = %v, want segmentCount error", err)
	}

	badRes := valid()
	badRes.FFMpegSetting.Qualities[0].Resolution = "hd720"
	err = Validate(badRes)
	if err == nil || !strings.Contains(err.Error(), "qualities[0].resolution") {
		t.Fatalf("Validate(resolution=hd720) = %v, want qualities[0].resolution error", err)
	}
}

func TestPostProcessAppliesLiveDefaults(t *testing.T) {
	var cfg Config
	cfg.Transcode.FFMpegSetting.Qualities = []Quality{{Resolution: "1280x720"}}
	if err := PostProcess(&cfg); err != nil {
		t.Fatalf("PostProcess() = %v", err)
	}
	if cfg.Transcode.Live != (Live{}).WithDefaults() {
		t.Fatalf("Live = %+v, want defaults", cfg.Transcode.Live)
	}

	bad := Config{}
	bad.Transcode.Live.SegmentCount = 6
	if err := PostProcess(&bad); err == nil {
		t.Fatal("PostProcess() accepted segmentCount 6")
	}
}
