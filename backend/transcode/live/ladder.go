package live

import (
	"fmt"
	"math"
	"sort"

	"sen1or/letslive/transcode/config"
)

const sourceRenditionName = "source"

// Rendition is one variant stream. Index 0 of a ladder is always the source,
// which is repackaged without re-encoding.
type Rendition struct {
	Name       string
	ShortSide  int
	Encoded    bool
	MaxBitrate string
	BufSize    string
	FPS        int
}

// BuildLadder returns the source followed by every configured quality whose
// short side is strictly smaller than the source's, in descending order.
// Comparing short sides keeps portrait sources from being upscaled.
func BuildLadder(src SourceInfo, qs []config.Quality) ([]Rendition, error) {
	srcShort := src.ShortSide()
	encoded := make([]Rendition, 0, len(qs))
	seen := make(map[int]struct{}, len(qs))

	for _, q := range qs {
		w, h, err := q.Size()
		if err != nil {
			return nil, fmt.Errorf("quality %q: %w", q.Resolution, err)
		}
		short := min(w, h)

		if _, dup := seen[short]; dup {
			return nil, fmt.Errorf("duplicate rendition %dp", short)
		}
		seen[short] = struct{}{}

		if short >= srcShort {
			continue
		}
		encoded = append(encoded, Rendition{
			Name:       fmt.Sprintf("%dp", short),
			ShortSide:  short,
			Encoded:    true,
			MaxBitrate: q.MaxBitrate,
			BufSize:    q.BufSize,
			FPS:        renditionFPS(q.FPS, src.FPS),
		})
	}

	sort.Slice(encoded, func(i, j int) bool { return encoded[i].ShortSide > encoded[j].ShortSide })

	return append([]Rendition{{Name: sourceRenditionName, ShortSide: srcShort}}, encoded...), nil
}

func renditionFPS(configured int, srcFPS float64) int {
	if math.IsNaN(srcFPS) || math.IsInf(srcFPS, 0) {
		return configured
	}
	rounded := int(math.Round(srcFPS))
	if rounded < 1 {
		return configured
	}
	return min(configured, rounded)
}
