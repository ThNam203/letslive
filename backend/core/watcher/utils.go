package watcher

import (
	"bufio"
	"fmt"
	"os"
	"sen1or/lets-live/models"
)

func generateRemotePlaylist(playlistPath string, variant models.HLSVariant) (string, error) {
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
				line = segment.IPFSRemoteId
			}
		}

		newPlaylist = newPlaylist + line + "\n"
	}

	return newPlaylist, nil
}
