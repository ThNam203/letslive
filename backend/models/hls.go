package models

import "path/filepath"

type HLSSegment struct {
	PublishName        string
	VariantIndex       int
	FullLocalPath      string // the full path to the file on disk
	RelativeRemotePath string // for example "1/stream0.ts", without the first part "http://...."
	IPFSRemoteId       string // hash id used with ipfs
}

// Multiple bitrates
type HLSVariant struct {
	VariantIndex uint8
	Segments     []HLSSegment
}

type HLSStream struct {
	Variants              []HLSVariant
	PublishName           string
	PublishFolderRemoteId string
}

func (v *HLSVariant) GetSegmentByFilename(fileName string) *HLSSegment {
	for _, segment := range v.Segments {
		if filepath.Base(segment.FullLocalPath) == fileName {
			return &segment
		}
	}

	return nil
}
