package watcher

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sen1or/lets-live/transcode/domains"
	"strings"
)

// rewrite the local playlist to point to remote resources
func GenerateRemotePlaylist(vodHandler VODHandler, playlistPath string, variant domains.HLSVariant) (string, error) {
	file, err := os.Open(playlistPath)
	if err != nil {
		return "", fmt.Errorf("can't open playlist %s: %s", playlistPath, err)
	}
	defer file.Close()

	var newPlaylist string = ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line[0] != '#' {
			segment := variant.GetSegmentByFilename(line)
			if segment == nil || segment.FullLocalPath == "" {
				line = ""
			} else {
				// adding fileName allow players to know the file is .ts instead of just file cid
				line = fmt.Sprintf("%s?fileName=%s", segment.RemoteID, filepath.Base(segment.FullLocalPath))
			}
		}

		newPlaylist = newPlaylist + line + "\n"
		vodHandler.OnGeneratingNewLineForRemotePlaylist(line, variant)
	}

	return newPlaylist, nil
}

func CopyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("error reading file: %s", err)
	}

	err = os.WriteFile(dst, input, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error copying file: %s", err)
	}

	return nil
}

// write the playlist (memory) into file destination
func WritePlaylist(data string, filePath string) error {
	parentDir := filepath.Dir(filePath)
	if err := os.MkdirAll(parentDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create parent folder %s: %s", parentDir, err)
	}

	f, err := os.Create(filePath)
	defer f.Close()

	if err != nil {
		return fmt.Errorf("failed to create file %s: %s", filePath, err)

	}
	_, err = f.WriteString(data)
	if err != nil {
		return fmt.Errorf("failed to write data into %s: %s", filePath, err)
	}

	return nil
}

func WritePlaylistForOtherGateway(data, originalGateway, alternativeGateway, alternativeGatewayFilename string) error {
	newData := strings.ReplaceAll(data, originalGateway, alternativeGateway)
	return WritePlaylist(newData, alternativeGatewayFilename)
}

func CopyMasterFileForOtherGateway(masterFilePath, otherGatewayURL, publicPath string) error {
	gatewayServerName := otherGatewayURL[7:]
	masterFile, err := os.ReadFile(masterFilePath)
	if err != nil {
		return fmt.Errorf("failed to open master file (%s): %s", masterFilePath, err)
	}

	newData := strings.ReplaceAll(string(masterFile), "stream.m3u8", gatewayServerName+"_stream.m3u8")

	f, err := os.Create(filepath.Join(publicPath, filepath.Base(filepath.Dir(masterFilePath)), gatewayServerName+"_index.m3u8"))
	defer f.Close()

	_, err = f.WriteString(newData)
	if err != nil {
		return fmt.Errorf("failed to copy master file for gateway (%s): %s", otherGatewayURL, err)
	}

	return nil
}
