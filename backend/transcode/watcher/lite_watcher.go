package watcher

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sen1or/lets-live/pkg/logger"
	"sen1or/lets-live/transcode/config"
	"sen1or/lets-live/transcode/domains"
	"sen1or/lets-live/transcode/storage"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/radovskyb/watcher"
)

func getSegmentFromPath(segmentFullPath string) *domains.HLSSegment {
	pathComponents := strings.Split(segmentFullPath, "/")
	index, err := strconv.Atoi(pathComponents[len(pathComponents)-2])
	if err != nil {
		log.Println("invalid segment path")
		return nil
	}

	name := pathComponents[len(pathComponents)-1]
	publishName := pathComponents[len(pathComponents)-3]

	return &domains.HLSSegment{
		VariantIndex:       index,
		FullLocalPath:      segmentFullPath,
		RelativeRemotePath: filepath.Join(string(index), name),
		PublishName:        publishName,
	}
}

var streams = make(map[string]domains.HLSStream)

// there is another type of Storage (KuboStorage which implements my Storage inteface)
// but it has so many features and my custom storage can not implement the Storage interface right now
// so for now I will use the CustomStorage directly (not using the Storage Interface)
// TODO: hope to be able to implement the Storage interface to the CustomStorage
type StreamWatcher struct {
	monitorPath string
	storage     storage.Storage
	config      config.Config
}

func NewStreamWatcher(monitorPath string, storage storage.Storage, config config.Config) *StreamWatcher {
	return &StreamWatcher{
		monitorPath: monitorPath,
		storage:     storage,
		config:      config,
	}
}

func (w *StreamWatcher) MonitorHLSStreamContent() {
	myWatcher := watcher.New()

	go func() {
		for {
			select {
			case event := <-myWatcher.Event:
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

					if err := os.MkdirAll(filepath.Join(w.config.Transcode.PublicHLSPath, publishName), os.ModePerm); err != nil {
						logger.Errorw("failed to create publish folder", "path", filepath.Join(w.config.Transcode.PublicHLSPath, publishName))
					} else {
						logger.Infof("created publish folder: %s", filepath.Join(w.config.Transcode.PublicHLSPath, publishName))
					}

					variants := make([]domains.HLSVariant, len(w.config.Transcode.FFMpegSetting.Qualities))
					for index := range variants {
						variants[index] = domains.HLSVariant{
							VariantIndex: uint8(index),
							Segments:     make([]domains.HLSSegment, 0),
						}
					}

					streams[publishName] = domains.HLSStream{
						PublishName: publishName,
						Variants:    variants,
					}

					log.Printf("created hls stream %+v\n", streams[publishName])

					continue
				}

				fileType := getEventFileType(event.Path)
				if fileType == "Master" {
					components := strings.Split(event.Path, "/")
					pushlishName := components[len(components)-2]

					if err := copy(event.Path, filepath.Join(w.config.Transcode.PublicHLSPath, pushlishName, w.config.Transcode.FFMpegSetting.MasterFileName)); err != nil {
						log.Panicf("failed to copy file: %s", err)
					}
				} else if fileType == "Variant" {
					info, err := getInfoFromPath(event.Path)
					if err != nil {
						fmt.Println(err)
						continue
					}
					logger.Infow("WATCHER", "streams", streams)
					logger.Infow("WATCHER", "stream info", info)
					variant := streams[info.PublishName].Variants[info.VariantIndex]
					newPlaylist, err := generateRemotePlaylist(event.Path, variant)
					if err != nil {
						fmt.Println("error generating remote playlist")
						continue
					}

					variantIndexStr := strconv.Itoa(info.VariantIndex)

					writePlaylist(newPlaylist, filepath.Join(w.config.Transcode.PublicHLSPath, info.PublishName, variantIndexStr, info.Filename))
				} else if fileType == "Segment" {
					segment := getSegmentFromPath(event.Path)
					if segment == nil {
						log.Printf("error creating segment")
						continue
					}

					stream, ok := streams[segment.PublishName]
					if !ok {
						logger.Errorw("failed to get stream from publish name", "streams", streams, "publish name", segment.PublishName)
						return
					}

					variant := &(stream.Variants[segment.VariantIndex])
					newObjectPathChannel := make(chan string, 1)

					// if there is no remote storage method available, we dont do anything
					go func() {
						var newObjectPath string = event.Path
						var err error

						if w.storage != nil {
							newObjectPath, err = w.storage.AddFile(event.Path)

							if err != nil {
								fmt.Printf("error while saving segments into storage: %s\n", err)
							}
						}

						newObjectPathChannel <- newObjectPath
					}()
					newObjectPath := <-newObjectPathChannel

					segment.IPFSRemoteId = newObjectPath
					variant.Segments = append(variant.Segments, *segment)
				}
			case err := <-myWatcher.Error:
				log.Panicf("something failed while running watcher: %s", err)
			case <-myWatcher.Closed:
				return
			}
		}
	}()

	// Watch the hls segment storage folder recursively for changes.
	if err := myWatcher.AddRecursive(w.monitorPath); err != nil {
		log.Fatalln(err)
	}

	if err := myWatcher.Start(time.Millisecond * 100); err != nil {
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
