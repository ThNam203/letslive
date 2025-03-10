package domains

import "path/filepath"

type HLSSegment struct {
	PublishName   string
	VariantIndex  int
	FullLocalPath string // the full path to the file on disk
	RemoteID      string // the full remove id
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
