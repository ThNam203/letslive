package live

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	tagM3U                 = "#EXTM3U"
	tagVersion             = "#EXT-X-VERSION:"
	tagIndependentSegments = "#EXT-X-INDEPENDENT-SEGMENTS"
	tagStreamInf           = "#EXT-X-STREAM-INF:"
)

var (
	renditionNameRe = regexp.MustCompile(`^[a-z0-9]+$`)
	// Attribute names are matched after ':' or ',' so e.g. TYPE=AUDIO is left alone.
	groupAttrRe = regexp.MustCompile(`([:,])(GROUP-ID|AUDIO)="([^"]*)"`)
	uriAttrRe   = regexp.MustCompile(`([:,])URI="([^"]*)"`)
)

// NamedPlaylist is one rendition muxer's multivariant playlist; Name is the
// URL path segment the muxer is mounted under.
type NamedPlaylist struct {
	Name string
	Body []byte
}

// MergeMultivariant combines per-rendition multivariant playlists into one.
// Every URI is prefixed with its rendition name and every audio group is
// renamed "{group}-{name}", so each variant keeps its own audio rendition.
// The header holds one #EXTM3U, the highest #EXT-X-VERSION and, if any input
// has it, one #EXT-X-INDEPENDENT-SEGMENTS. Inputs keep their order.
func MergeMultivariant(parts []NamedPlaylist) ([]byte, error) {
	if len(parts) == 0 {
		return nil, errors.New("no playlists to merge")
	}

	version := 0
	independent := false
	bodies := make([][]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))

	for _, p := range parts {
		if !renditionNameRe.MatchString(p.Name) {
			return nil, fmt.Errorf("rendition name %q must match %s", p.Name, renditionNameRe)
		}
		if _, dup := seen[p.Name]; dup {
			return nil, fmt.Errorf("duplicate rendition name %q", p.Name)
		}
		seen[p.Name] = struct{}{}

		pl, err := parseMultivariant(p)
		if err != nil {
			return nil, fmt.Errorf("rendition %s: %w", p.Name, err)
		}
		version = max(version, pl.version)
		independent = independent || pl.independent
		bodies = append(bodies, pl.lines)
	}

	var out bytes.Buffer
	out.WriteString(tagM3U + "\n")
	if version > 0 {
		fmt.Fprintf(&out, "%s%d\n", tagVersion, version)
	}
	if independent {
		out.WriteString(tagIndependentSegments + "\n")
	}
	for _, lines := range bodies {
		out.WriteString("\n")
		for _, line := range lines {
			out.WriteString(line + "\n")
		}
	}
	return out.Bytes(), nil
}

type multivariant struct {
	version     int
	independent bool
	lines       []string // rewritten body without header tags and blank lines
}

func parseMultivariant(p NamedPlaylist) (multivariant, error) {
	rawLines := strings.Split(string(p.Body), "\n")
	if strings.TrimRight(rawLines[0], "\r") != tagM3U {
		return multivariant{}, errors.New("playlist does not start with #EXTM3U")
	}

	var pl multivariant
	expectURI := false
	variants := 0

	for _, raw := range rawLines[1:] {
		line := strings.TrimRight(raw, "\r")
		switch {
		case line == "":
			continue

		case strings.HasPrefix(line, tagVersion):
			v, err := strconv.Atoi(strings.TrimPrefix(line, tagVersion))
			if err != nil || v <= 0 {
				return multivariant{}, fmt.Errorf("invalid %s", line)
			}
			pl.version = max(pl.version, v)

		case line == tagIndependentSegments:
			pl.independent = true

		case strings.HasPrefix(line, "#"):
			if expectURI {
				return multivariant{}, fmt.Errorf("%s has no URI line", tagStreamInf)
			}
			pl.lines = append(pl.lines, rewriteTag(line, p.Name))
			expectURI = strings.HasPrefix(line, tagStreamInf)

		default:
			if !expectURI {
				return multivariant{}, fmt.Errorf("URI line %q does not follow %s", line, tagStreamInf)
			}
			pl.lines = append(pl.lines, p.Name+"/"+line)
			expectURI = false
			variants++
		}
	}

	if expectURI {
		return multivariant{}, fmt.Errorf("%s has no URI line", tagStreamInf)
	}
	if variants == 0 {
		return multivariant{}, errors.New("playlist has no variant stream")
	}
	return pl, nil
}

func rewriteTag(line, name string) string {
	line = groupAttrRe.ReplaceAllString(line, `$1$2="$3-`+name+`"`)
	return uriAttrRe.ReplaceAllString(line, `${1}URI="`+name+`/$2"`)
}
