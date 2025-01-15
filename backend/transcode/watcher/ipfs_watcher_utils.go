package watcher

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sen1or/lets-live/transcode/domains"
)

// rewrite the local playlist to point to remote resources
func generateRemotePlaylist(ipfsVOD *IPFSVOD, playlistPath string, variant domains.HLSVariant) (string, error) {
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
		ipfsVOD.OnGeneratingNewLineForRemotePlaylist(line, variant)
	}

	return newPlaylist, nil
}

func copy(src, dst string) error {
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
func writePlaylist(data string, filePath string) error {
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
