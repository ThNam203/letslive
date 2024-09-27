package internal

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sen1or/lets-live/internal/config"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/radovskyb/watcher"
)

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

func getSegmentFromPath(segmentFullPath string) *HLSSegment {
	pathComponents := strings.Split(segmentFullPath, "/")
	index, err := strconv.Atoi(pathComponents[len(pathComponents)-2])
	if err != nil {
		log.Println("invalid segment path")
		return nil
	}

	name := pathComponents[len(pathComponents)-1]
	publishName := pathComponents[len(pathComponents)-3]

	return &HLSSegment{
		VariantIndex:       index,
		FullLocalPath:      segmentFullPath,
		RelativeRemotePath: filepath.Join(string(index), name),
		PublishName:        publishName,
	}
}

var cfg = config.GetConfig()
var streams = make(map[string]HLSStream)

func MonitorHLSStreamContent(monitorPath string, storage Storage) {
	w := watcher.New()

	go func() {
		for {
			select {
			case event := <-w.Event:
				if event.Op == watcher.Remove {
					continue
				}

				// NOT SURE WHAT IS CREATED FIRST
				// handle creating the publish name folder
				if event.IsDir() && event.Op == watcher.Create {
					components := strings.Split(event.Path, "/")
					publishName := components[len(components)-1]

					// TODO: properly handle
					if utf8.RuneCountInString(publishName) < 10 {
						continue
					}

					publishFolderHash, err := storage.AddDirectory(event.Path)
					if err != nil {
						fmt.Println("failed to add publish folder into storage")
						return
					}

					variants := make([]HLSVariant, len(cfg.FFMpegSetting.Qualities))
					for index := range variants {
						variants[index] = HLSVariant{uint8(index), make([]HLSSegment, 0)}
					}

					streams[publishName] = HLSStream{
						PublishName:           publishName,
						Variants:              variants,
						PublishFolderRemoteId: publishFolderHash,
					}

					log.Printf("created hls stream %+v\n", streams[publishName])

					continue
				}

				fileType := getEventFileType(event.Path)
				if fileType == "Master" {
					components := strings.Split(event.Path, "/")
					pushlishName := components[len(components)-2]

					if err := copy(event.Path, filepath.Join(cfg.PublicHLSPath, pushlishName, cfg.FFMpegSetting.MasterFileName)); err != nil {
						log.Panicf("failed to copy file: %s", err)
					}
				} else if fileType == "Variant" {
					info, err := getInfoFromPath(event.Path)
					if err != nil {
						fmt.Println(err)
						continue
					}
					variant := streams[info.PublishName].Variants[info.VariantIndex]
					newPlaylist, err := storage.GenerateRemotePlaylist(event.Path, variant)
					if err != nil {
						fmt.Println("error generating remote playlist")
						continue
					}

					variantIndexStr := strconv.Itoa(info.VariantIndex)

					writePlaylist(newPlaylist, filepath.Join(cfg.PublicHLSPath, info.PublishName, variantIndexStr, info.Filename))
				} else if fileType == "Segment" {
					segment := getSegmentFromPath(event.Path)
					if segment == nil {
						log.Printf("error creating segment")
						continue
					}

					variant := &(streams[segment.PublishName].Variants[segment.VariantIndex])

					newObjectPathChannel := make(chan string, 1)
					go func() {
						newObjectPath, err := storage.SaveIntoHLSDirectory(event.Path)

						if err != nil {
							fmt.Printf("error while saving segments into ipfs: %s\n", err)
						}

						newObjectPathChannel <- newObjectPath
					}()
					newObjectPath := <-newObjectPathChannel

					segment.IPFSRemoteId = newObjectPath
					variant.Segments = append(variant.Segments, *segment)
				}
			case err := <-w.Error:
				log.Panicf("something failed while running watcher: %s", err)
			case <-w.Closed:
				return
			}
		}
	}()

	// Watch the hls segment storage folder recursively for changes.
	if err := w.AddRecursive(monitorPath); err != nil {
		log.Fatalln(err)
	}

	if err := w.Start(time.Millisecond * 100); err != nil {
		log.Fatalln(err)
	}
}

type pathInfo struct {
	VariantIndex int
	Filename     string
	PublishName  string
}

// MUST NOT USE FOR INDEX FILE
func getInfoFromPath(filePath string) (*pathInfo, error) {
	components := strings.Split(filePath, "/")
	filename := components[len(components)-1]
	variantIndex, err := strconv.Atoi(components[len(components)-2])
	publishName := components[len(components)-3]
	if err != nil {
		return nil, fmt.Errorf("error getting variant index: %s", err)
	}

	info := &pathInfo{
		VariantIndex: variantIndex,
		Filename:     filename,
		PublishName:  publishName,
	}

	return info, nil
}

// getFileType return one of the three: Master, Variant, Segment
func getEventFileType(filePath string) string {
	pathComponents := strings.Split(filePath, "/")

	if filepath.Ext(filePath) == ".m3u8" {
		// the parent folder of Variant type is an index (1, 2, 3,...)
		if utf8.RuneCountInString(pathComponents[len(pathComponents)-2]) == 1 {
			return "Variant"
		}

		return "Master"
	} else if filepath.Ext(filePath) == ".ts" {
		return "Segment"
	}

	return ""
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

func writePlaylist(data string, filePath string) {
	parentDir := filepath.Dir(filePath)
	if err := os.MkdirAll(parentDir, os.ModePerm); err != nil {
		fmt.Printf("failed to create parent folder %s: %s", parentDir, err)
		return
	}

	f, err := os.Create(filePath)
	defer f.Close()

	if err != nil {
		fmt.Printf("failed to create file %s: %s", filePath, err)
		return
	}
	_, err = f.WriteString(data)
	if err != nil {
		fmt.Printf("failed to write data into %s: %s", filePath, err)
		return
	}
}
