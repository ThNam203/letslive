package live

import (
	"errors"
	"fmt"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h264"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/mpeg4audio"
	"github.com/nareix/joy5/av"
)

// VideoAU splits an RTMP H.264 packet (AVCC, 4-byte length-prefixed NAL
// units) into NAL units.
func VideoAU(pkt av.Packet) ([][]byte, error) {
	var avcc h264.AVCC
	if err := avcc.Unmarshal(pkt.Data); err != nil {
		return nil, fmt.Errorf("parse AVCC access unit: %w", err)
	}
	return avcc, nil
}

// AudioConfig parses the AudioSpecificConfig of an AAC sequence header packet.
// joy5's FLV demuxer puts it in Data; ASeqHdr is the fallback for packets
// built by other sources.
func AudioConfig(pkt av.Packet) (*mpeg4audio.AudioSpecificConfig, error) {
	if pkt.Type != av.AACDecoderConfig {
		return nil, fmt.Errorf("packet type %v is not an AAC decoder config", pkt.Type)
	}

	buf := pkt.Data
	if len(buf) == 0 {
		buf = pkt.ASeqHdr
	}
	if len(buf) == 0 {
		return nil, errors.New("AAC decoder config is empty")
	}

	var asc mpeg4audio.AudioSpecificConfig
	if err := asc.Unmarshal(buf); err != nil {
		return nil, fmt.Errorf("parse AudioSpecificConfig: %w", err)
	}
	return &asc, nil
}
