package live

import (
	"testing"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h264"
	"github.com/nareix/joy5/av"
)

func TestVideoAU(t *testing.T) {
	var key *av.Packet
	for _, pkt := range readFLV(t, "src_640x360.flv") {
		if pkt.Type == av.H264 && pkt.IsKeyFrame {
			key = &pkt
			break
		}
	}
	if key == nil {
		t.Fatal("no H264 keyframe in fixture")
	}

	au, err := VideoAU(*key)
	if err != nil {
		t.Fatalf("VideoAU: %v", err)
	}
	if len(au) < 1 {
		t.Fatal("VideoAU returned no NAL units")
	}
	if !h264.IsRandomAccess(au) {
		t.Error("keyframe AU is not a random access point")
	}
}

func TestVideoAUInvalid(t *testing.T) {
	cases := map[string][]byte{
		"empty":        nil,
		"truncated":    {0, 0, 0, 9, 1, 2},
		"short header": {0, 0},
		"only empty":   {0, 0, 0, 0},
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			if au, err := VideoAU(av.Packet{Type: av.H264, Data: data}); err == nil {
				t.Errorf("VideoAU = %v, want error", au)
			}
		})
	}
}

func TestAudioConfig(t *testing.T) {
	var cfg *av.Packet
	for _, pkt := range readFLV(t, "src_640x360.flv") {
		if pkt.Type == av.AACDecoderConfig {
			cfg = &pkt
			break
		}
	}
	if cfg == nil {
		t.Fatal("no AACDecoderConfig packet in fixture")
	}

	asc, err := AudioConfig(*cfg)
	if err != nil {
		t.Fatalf("AudioConfig: %v", err)
	}
	if asc.SampleRate != 44100 {
		t.Errorf("SampleRate = %d, want 44100", asc.SampleRate)
	}
}

func TestAudioConfigFallsBackToASeqHdr(t *testing.T) {
	for _, pkt := range readFLV(t, "src_640x360.flv") {
		if pkt.Type != av.AACDecoderConfig {
			continue
		}
		asc, err := AudioConfig(av.Packet{Type: av.AACDecoderConfig, ASeqHdr: pkt.Data})
		if err != nil {
			t.Fatalf("AudioConfig: %v", err)
		}
		if asc.SampleRate != 44100 {
			t.Errorf("SampleRate = %d, want 44100", asc.SampleRate)
		}
		return
	}
	t.Fatal("no AACDecoderConfig packet in fixture")
}

func TestAudioConfigErrors(t *testing.T) {
	if _, err := AudioConfig(av.Packet{Type: av.AAC, Data: []byte{0x12, 0x10}}); err == nil {
		t.Error("raw AAC packet: want error")
	}
	if _, err := AudioConfig(av.Packet{Type: av.AACDecoderConfig}); err == nil {
		t.Error("empty config: want error")
	}
	if _, err := AudioConfig(av.Packet{Type: av.AACDecoderConfig, Data: []byte{0xFF}}); err == nil {
		t.Error("garbage config: want error")
	}
}
