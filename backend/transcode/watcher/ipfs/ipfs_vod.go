package ipfswatcher

import (
	"fmt"
	"os"
	"path/filepath"
	"sen1or/lets-live/pkg/logger"
	"sen1or/lets-live/transcode/domains"
	"sen1or/lets-live/transcode/watcher"
	"strings"
	"sync"
	"time"
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

type IPFSVODStrategy struct {
	vodsData map[string]*VODData
	mu       sync.RWMutex
}

func GetIPFSVODHandler() watcher.VODHandler {
	return &IPFSVODStrategy{
		vodsData: make(map[string]*VODData),
	}
}

func (u *IPFSVODStrategy) OnStreamStart(publishName string) {
	u.mu.Lock()
	_, exist := u.vodsData[publishName]
	if !exist {
		u.vodsData[publishName] = CreateNewVODData()
	}
	u.mu.Unlock()
}

func (u *IPFSVODStrategy) OnStreamEnd(publishName string, publicHLSPath string, masterFileName string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	defer delete(u.vodsData, publishName)

	var masterFileDirPath = filepath.Join(publicHLSPath, publishName)
	var nowString = time.Now().Format(time.RFC3339)
	var outputPath = filepath.Join(masterFileDirPath, "vods", nowString)

	vodData, exist := u.vodsData[publishName]
	if !exist {
		return
	}

	// create the base output directory if it doesn't exist
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		logger.Errorf("failed to create base output directory: %w", err)
		return
	}

	// generate and save the variant playlists
	if err := u.generateVariantVODPlaylists(*vodData, outputPath); err != nil {
		logger.Errorf("failed to generate variant playlists: %w", err)
		return
	}

	// copy the master file to vod folder
	newMasterFilePath := filepath.Join(outputPath, masterFileName)
	watcher.CopyFile(filepath.Join(masterFileDirPath, masterFileName), newMasterFilePath)

	// generate master files for other gateways
	for _, otherGatewayURL := range otherGateways {
		generateMasterFileVODSForOtherGateway(newMasterFilePath, otherGatewayURL)
	}
}

func (u *IPFSVODStrategy) generateVariantVODPlaylist(data VODData, index int) string {
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

func (u *IPFSVODStrategy) generateVariantVODPlaylists(vodData VODData, outputPath string) error {
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

		// generate for other gateways, TODO: refactor
		for _, otherGateway := range otherGateways {
			serverName := otherGateway[7:]
			playlistNewData := strings.ReplaceAll(playlist, "localhost:8888", serverName)
			if err := os.WriteFile(filepath.Join(filepath.Dir(playlistPath), serverName+"_stream.m3u8"), []byte(playlistNewData), 0644); err != nil {
				logger.Errorf("failed to write playlist file for gateway %s: %w", otherGateway, err)
			}
		}
	}

	return nil
}

// save data for creating VOD
func (u *IPFSVODStrategy) OnGeneratingNewLineForRemotePlaylist(line string, variant domains.HLSVariant) {
	if len(variant.Segments) == 0 || len(line) == 0 {
		return
	}

	sampleSegment := variant.Segments[0]
	variantIndex := sampleSegment.VariantIndex
	publishName := sampleSegment.PublishName

	vodData, ok := u.vodsData[publishName]
	if !ok {
		return
	}

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

func generateMasterFileVODSForOtherGateway(masterFilePath, otherGatewayURL string) error {
	gatewayServerName := otherGatewayURL[7:]
	masterFile, err := os.ReadFile(masterFilePath)
	if err != nil {
		return fmt.Errorf("failed to open master file (%s): %s", masterFilePath, err)
	}

	newData := strings.ReplaceAll(string(masterFile), "stream.m3u8", gatewayServerName+"_stream.m3u8")

	path := filepath.Join(filepath.Dir(masterFilePath), gatewayServerName+"_index.m3u8")

	err = os.WriteFile(path, []byte(newData), 0644)
	if err != nil {
		return fmt.Errorf("failed to generate master file for gateway (%s): %s", otherGatewayURL, err)
	}

	return nil
}
