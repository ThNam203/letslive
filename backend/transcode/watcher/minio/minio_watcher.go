package miniowatcher

import (
	"fmt"
	"os"
	"path/filepath"
	"sen1or/lets-live/pkg/logger"
	"sen1or/lets-live/transcode/config"
	"sen1or/lets-live/transcode/domains"
	"sen1or/lets-live/transcode/storage"
	mywatcher "sen1or/lets-live/transcode/watcher"
	"strconv"
	"strings"

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

type MinIOFileWatcherStrategy struct {
	storage    storage.Storage
	config     config.Config
	vodHandler mywatcher.VODHandler
}

func NewMinIOFileWatcherStrategy(vodHandler mywatcher.VODHandler, storage storage.Storage, config config.Config) mywatcher.FileWatcherStrategy {
	return &MinIOFileWatcherStrategy{
		storage:    storage,
		config:     config,
		vodHandler: vodHandler,
	}
}

func (w *MinIOFileWatcherStrategy) OnCreate(event watcher.Event) {
	components := strings.Split(event.Path, "/")
	publishName := components[len(components)-1]

	if len(publishName) < 10 {
		return
	}
	if err := os.MkdirAll(filepath.Join(w.config.Transcode.PublicHLSPath, publishName), os.ModePerm); err != nil {
		logger.Errorw("making dir failed", "failed to create publish folder", err, "path", filepath.Join(w.config.Transcode.PublicHLSPath, publishName))
		return
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

}

func (w *MinIOFileWatcherStrategy) OnMaster(event watcher.Event) {
	components := strings.Split(event.Path, "/")
	publishName := components[len(components)-2]

	if err := mywatcher.CopyFile(event.Path, filepath.Join(w.config.Transcode.PublicHLSPath, publishName, w.config.Transcode.FFMpegSetting.MasterFileName)); err != nil {
		logger.Errorw("failed to copy master file", err)
	}
}

func (w *MinIOFileWatcherStrategy) OnVariant(event watcher.Event) {
	info, err := w.getInfoFromPath(event.Path)
	if err != nil {
		logger.Errorw("failed to get variant info", err)
		return
	}

	variant := streams[info.PublishName].Variants[info.VariantIndex]
	newPlaylist, err := mywatcher.GenerateRemotePlaylist(w.vodHandler, event.Path, variant)
	if err != nil {
		logger.Errorw("error generating remote playlist", err)
		return
	}

	variantIndexStr := strconv.Itoa(info.VariantIndex)

	mywatcher.WritePlaylist(newPlaylist, filepath.Join(w.config.Transcode.PublicHLSPath, info.PublishName, variantIndexStr, info.Filename))
}
func (w *MinIOFileWatcherStrategy) OnSegment(event watcher.Event) {
	segment, err := getSegmentFromPath(event.Path)
	if segment == nil {
		logger.Errorw("error getting segment", err)
		return
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
			newObjectPath, err = w.storage.AddFile(event.Path, stream.PublishName)

			if err != nil {
				logger.Errorf("error while saving segments into storage", err)
			}
		}

		newObjectPathChannel <- newObjectPath
	}()
	newObjectPath := <-newObjectPathChannel

	logger.Infof("added a new file with url: %s", newObjectPath)

	segment.RemoteID = newObjectPath
	variant.Segments = append(variant.Segments, *segment)
}

type pathInfo struct {
	VariantIndex int
	Filename     string
	PublishName  string
}

// MUST NOT USE FOR INDEX FILE
func (_ *MinIOFileWatcherStrategy) getInfoFromPath(filePath string) (*pathInfo, error) {
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
