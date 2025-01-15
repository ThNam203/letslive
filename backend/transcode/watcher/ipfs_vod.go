package watcher

import (
	"fmt"
	"os"
	"path/filepath"
	"sen1or/lets-live/pkg/logger"
	"sen1or/lets-live/transcode/domains"
	"strings"
	"sync"
)

type VODData struct {
	HLSVersion        string
	HLSTargetDuration string
	HLSINF            string
	Segments          [3][]string
}

func CreateNewVODData() *VODData {
	var newVOD = VODData{}

	for i := range newVOD.Segments {
		newVOD.Segments[i] = []string{} // Initialize as an empty slice
	}

	return &newVOD
}

type IPFSVOD struct {
	vodsData map[string]*VODData
	mu       sync.RWMutex
}

func GetIPFSVOD() *IPFSVOD {
	return &IPFSVOD{
		vodsData: make(map[string]*VODData),
	}
}

func (u *IPFSVOD) OnStreamStart(publishName string) {
	u.mu.Lock()
	_, exist := u.vodsData[publishName]
	if !exist {
		u.vodsData[publishName] = CreateNewVODData()
	}
	u.mu.Unlock()
}

func (u *IPFSVOD) OnStreamEnd(publishName string, outputPath string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	defer delete(u.vodsData, publishName)

	vodData, exist := u.vodsData[publishName]
	if !exist {
		return
	}

	// Create the base output directory if it doesn't exist
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		logger.Errorf("failed to create base output directory: %w", err)
		return
	}

	// Generate and save the variant playlists
	if err := u.generateVariantVODPlaylists(*vodData, outputPath); err != nil {
		logger.Errorf("failed to generate variant playlists: %w", err)
		return
	}
}

func (u *IPFSVOD) generateVariantVODPlaylist(data VODData, index int) string {
	var playlist strings.Builder

	// Write header
	playlist.WriteString("#EXTM3U\n")
	playlist.WriteString(fmt.Sprintf("%s\n", data.HLSVersion))
	playlist.WriteString(fmt.Sprintf("%s\n", data.HLSTargetDuration))
	playlist.WriteString("#EXT-X-PLAYLIST-TYPE:VOD\n")

	// Write segments
	for _, segment := range data.Segments[index] {
		playlist.WriteString(fmt.Sprintf("%s\n", data.HLSINF))
		playlist.WriteString(segment + "\n")
	}

	// Write end marker
	playlist.WriteString("#EXT-X-ENDLIST\n")

	return playlist.String()
}

func (u *IPFSVOD) generateVariantVODPlaylists(vodData VODData, outputPath string) error {
	// Generate and save each variant playlist
	for i := 0; i < 3; i++ {
		// Create directory for this quality level
		qualityDir := filepath.Join(outputPath, fmt.Sprintf("%d", i))
		if err := os.MkdirAll(qualityDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", qualityDir, err)
		}

		// Generate playlist content
		playlist := u.generateVariantVODPlaylist(vodData, i)

		// Create playlist file path
		playlistPath := filepath.Join(qualityDir, "stream.m3u8")

		// Write playlist to file
		if err := os.WriteFile(playlistPath, []byte(playlist), 0644); err != nil {
			return fmt.Errorf("failed to write playlist file %s: %w", playlistPath, err)
		}
	}

	return nil
}

// save data for creating VOD
func (u *IPFSVOD) OnGeneratingNewLineForRemotePlaylist(line string, variant domains.HLSVariant) {
	if len(variant.Segments) == 0 || len(line) == 0 {
		return
	}

	sampleSegment := variant.Segments[0]
	variantIndex := sampleSegment.VariantIndex
	publishName := sampleSegment.PublishName

	vodData, _ := u.vodsData[publishName] // should always be there, created when watcher receives a 'create folder' event

	isDone := false
	if strings.HasPrefix(line, "#EXT-X-VERSION") {
		vodData.HLSVersion = line
		isDone = true
	}
	if strings.HasPrefix(line, "#EXT-X-TARGETDURATION") {
		vodData.HLSTargetDuration = line
		isDone = true
	}
	if strings.HasPrefix(line, "#EXTINF") {
		vodData.HLSINF = line
		isDone = true
	}
	if strings.HasPrefix(line, "#") {
		isDone = true
	}

	if !isDone {
		isNew := true
		for _, dataLine := range vodData.Segments[variantIndex] {
			if dataLine == line {
				isNew = false
				break
			}
		}

		if isNew {
			vodData.Segments[variantIndex] = append(vodData.Segments[variantIndex], line)
		}
	}
}
