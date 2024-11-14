package watcher

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sen1or/lets-live/transcode/domains"
)

// rewrite the playlist to point to remote resources
func generateRemotePlaylist(playlistPath string, variant domains.HLSVariant) (string, error) {
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
				line = fmt.Sprintf("%s?fileName=%s", segment.IPFSRemoteId, filepath.Base(segment.FullLocalPath))
			}
		}

		newPlaylist = newPlaylist + line + "\n"
	}

	return newPlaylist, nil
}
