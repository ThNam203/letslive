package live

import (
	"errors"
	"fmt"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h264"
)

const (
	avcConfigVersion = 1
	avcConfigHeader  = 6 // version, profile, compat, level, lengthSize, numSPS
)

// SourceInfo describes the incoming H.264 stream. FPS is 0 when the SPS carries no timing info.
type SourceInfo struct {
	Width, Height int
	FPS           float64
	SPS, PPS      []byte
}

func (s SourceInfo) ShortSide() int {
	return min(s.Width, s.Height)
}

func (s SourceInfo) Portrait() bool {
	return s.Height > s.Width
}

// ProbeAVCConfig parses an AVCDecoderConfigurationRecord (ISO 14496-15) and the
// first SPS it carries. The first SPS and PPS are kept in SourceInfo.
func ProbeAVCConfig(record []byte) (SourceInfo, error) {
	if len(record) < avcConfigHeader {
		return SourceInfo{}, fmt.Errorf("AVC config too short: %d bytes", len(record))
	}
	if record[0] != avcConfigVersion {
		return SourceInfo{}, fmt.Errorf("unsupported AVC config version %d", record[0])
	}

	numSPS := int(record[5] & 0x1f)
	if numSPS == 0 {
		return SourceInfo{}, errors.New("AVC config has no SPS")
	}

	rest := record[avcConfigHeader:]
	var sps []byte
	for i := 0; i < numSPS; i++ {
		nalu, next, err := readParamSet(rest)
		if err != nil {
			return SourceInfo{}, fmt.Errorf("read SPS %d: %w", i, err)
		}
		if i == 0 {
			sps = nalu
		}
		rest = next
	}

	if len(rest) < 1 {
		return SourceInfo{}, errors.New("AVC config missing PPS count")
	}
	numPPS := int(rest[0])
	if numPPS == 0 {
		return SourceInfo{}, errors.New("AVC config has no PPS")
	}
	pps, _, err := readParamSet(rest[1:])
	if err != nil {
		return SourceInfo{}, fmt.Errorf("read PPS: %w", err)
	}

	var parsed h264.SPS
	if err := parsed.Unmarshal(sps); err != nil {
		return SourceInfo{}, fmt.Errorf("parse SPS: %w", err)
	}

	return SourceInfo{
		Width:  parsed.Width(),
		Height: parsed.Height(),
		FPS:    parsed.FPS(),
		SPS:    sps,
		PPS:    pps,
	}, nil
}

// readParamSet reads one 16-bit length-prefixed NAL unit and returns it with the remaining bytes.
func readParamSet(buf []byte) (nalu, rest []byte, err error) {
	if len(buf) < 2 {
		return nil, nil, errors.New("truncated length")
	}
	n := int(buf[0])<<8 | int(buf[1])
	if n == 0 || len(buf) < 2+n {
		return nil, nil, fmt.Errorf("invalid length %d with %d bytes left", n, len(buf)-2)
	}
	return buf[2 : 2+n], buf[2+n:], nil
}
