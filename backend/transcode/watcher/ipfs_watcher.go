package watcher

import (
	"fmt"
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

func getSegmentFromPath(segmentFullPath string) (*domains.HLSSegment, error) {
	pathComponents := strings.Split(segmentFullPath, "/")
	index, err := strconv.Atoi(pathComponents[len(pathComponents)-2])
	if err != nil {
		return nil, fmt.Errorf("invalid segment path %s", segmentFullPath)
	}

	name := pathComponents[len(pathComponents)-1]
	publishName := pathComponents[len(pathComponents)-3]

	return &domains.HLSSegment{
		VariantIndex:       index,
		FullLocalPath:      segmentFullPath,
		RelativeRemotePath: filepath.Join(string(index), name),
		PublishName:        publishName,
	}, nil
}

var streams = make(map[string]domains.HLSStream)

// there is another type of Storage (KuboStorage which implements my Storage inteface)
// but it has so many features and my custom storage can not implement the Storage interface right now
// so for now I will use the CustomStorage directly (not using the Storage Interface)
// TODO: hope to be able to implement the Storage interface to the CustomStorage
type IPFSStreamWatcher struct {
	monitorPath string
	storage     storage.Storage
	config      config.Config
	ipfsVOD     *IPFSVOD
}

func NewIPFSWatcher(monitorPath string, ipfsVOD *IPFSVOD, ipfsStorage storage.Storage, config config.Config) Watcher {
	return &IPFSStreamWatcher{
		monitorPath: monitorPath,
		storage:     ipfsStorage,
		config:      config,
		ipfsVOD:     ipfsVOD,
	}
}

func (w *IPFSStreamWatcher) Watch() {
	myWatcher := watcher.New()

	go func() {
		for {
			select {
			case event := <-myWatcher.Event:
				if event.Op == watcher.Remove {
					continue
				}

				if event.IsDir() && event.Op == watcher.Create {
					components := strings.Split(event.Path, "/")
					publishName := components[len(components)-1]

					if len(publishName) < 10 {
						continue
					}

					if err := os.MkdirAll(filepath.Join(w.config.Transcode.PublicHLSPath, publishName), os.ModePerm); err != nil {
						logger.Errorw("failed to create publish folder", err, "path", filepath.Join(w.config.Transcode.PublicHLSPath, publishName))
						continue
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

					logger.Infof("created hls stream with path: %s", streams[publishName])

					continue
				}

				fileType := w.getEventFileType(event.Path)
				if fileType == "Master" {
					components := strings.Split(event.Path, "/")
					pushlishName := components[len(components)-2]

					if err := copy(event.Path, filepath.Join(w.config.Transcode.PublicHLSPath, pushlishName, w.config.Transcode.FFMpegSetting.MasterFileName)); err != nil {
						logger.Errorw("failed to copy master file", err)
					}
				} else if fileType == "Variant" {
					info, err := w.getInfoFromPath(event.Path)
					if err != nil {
						logger.Errorw("failed to get variant info", err)
						continue
					}

					variant := streams[info.PublishName].Variants[info.VariantIndex]
					newPlaylist, err := generateRemotePlaylist(w.ipfsVOD, event.Path, variant)
					if err != nil {
						logger.Errorw("error generating remote playlist", err)
						continue
					}

					variantIndexStr := strconv.Itoa(info.VariantIndex)

					writePlaylist(newPlaylist, filepath.Join(w.config.Transcode.PublicHLSPath, info.PublishName, variantIndexStr, info.Filename))
				} else if fileType == "Segment" {
					segment, err := getSegmentFromPath(event.Path)
					if segment == nil {
						logger.Errorw("error getting segment", err)
						continue
					}

					stream, ok := streams[segment.PublishName]
					if !ok {
						logger.Errorw("missing entry for publish name", segment.PublishName)
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
								logger.Errorf("error while saving segments into storage", err)
							}
						}

						newObjectPathChannel <- newObjectPath
					}()
					newObjectPath := <-newObjectPathChannel

					segment.IPFSRemoteId = newObjectPath
					variant.Segments = append(variant.Segments, *segment)
				}
			case err := <-myWatcher.Error:
				logger.Errorf("something failed while running watcher", err)
			case <-myWatcher.Closed:
				return
			}
		}
	}()

	// Watch the hls segment storage folder recursively for changes.
	if err := myWatcher.AddRecursive(w.monitorPath); err != nil {
		logger.Panicw("error while setting up", err)
	}

	if err := myWatcher.Start(time.Millisecond * 100); err != nil {
		logger.Panicw("error starting watcher", err)
	}
}

type pathInfo struct {
	VariantIndex int
	Filename     string
	PublishName  string
}

// MUST NOT USE FOR INDEX FILE
func (_ *IPFSStreamWatcher) getInfoFromPath(filePath string) (*pathInfo, error) {
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

// getFileType should return one of the three: Master, Variant, Segment
func (_ *IPFSStreamWatcher) getEventFileType(filePath string) string {
	pathComponents := strings.Split(filePath, "/")
	fileExtension := filepath.Ext(filePath)

	if fileExtension == ".m3u8" {
		// the parent folder of Variant type is an index (1, 2, 3,...)
		if utf8.RuneCountInString(pathComponents[len(pathComponents)-2]) == 1 {
			return "Variant"
		}

		return "Master"
	} else if filepath.Ext(filePath) == ".ts" {
		return "Segment"
	}

	return filepath.Ext(filePath)
}
