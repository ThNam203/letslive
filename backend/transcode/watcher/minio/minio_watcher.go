package miniowatcher

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sen1or/letslive/transcode/config"
	"sen1or/letslive/transcode/domains"
	"sen1or/letslive/transcode/pkg/logger"
	"sen1or/letslive/transcode/storage"
	mywatcher "sen1or/letslive/transcode/watcher"
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

	publishName := pathComponents[len(pathComponents)-3]

	return &domains.HLSSegment{
		VariantIndex:  index,
		FullLocalPath: segmentFullPath,
		PublishName:   publishName,
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

func (w *MinIOFileWatcherStrategy) OnCreate(event watcher.Event) error {
	components := strings.Split(event.Path, "/")
	publishName := components[len(components)-1]

	if len(publishName) < 10 {
		return errors.New("publish name must be longer than 9 (uuid)")
	}
	if err := os.MkdirAll(filepath.Join(w.config.Transcode.PublicHLSPath, publishName), os.ModePerm); err != nil {
		logger.Errorw("making dir failed", "failed to create publish folder", err, "path", filepath.Join(w.config.Transcode.PublicHLSPath, publishName))
		return errors.New("failed to create publish folder")
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

	return nil
}

func (w *MinIOFileWatcherStrategy) OnMaster(event watcher.Event) error {
	components := strings.Split(event.Path, "/")
	publishName := components[len(components)-2]

	if err := mywatcher.CopyFile(event.Path, filepath.Join(w.config.Transcode.PublicHLSPath, publishName, w.config.Transcode.FFMpegSetting.MasterFileName)); err != nil {
		logger.Errorf("failed to copy master file: %s", err)
		return errors.New("failed to copy master file")
	}

	return nil
}

func (w *MinIOFileWatcherStrategy) OnVariant(event watcher.Event) error {
	info, err := w.getInfoFromPath(event.Path)
	if err != nil {
		logger.Errorf("failed to get variant info: %s", err)
		return errors.New("failed to get variant info")
	}

	variant := streams[info.PublishName].Variants[info.VariantIndex]
	newPlaylist, err := mywatcher.GenerateRemotePlaylist(w.vodHandler, event.Path, variant)
	if err != nil {
		logger.Errorf("error generating remote playlist: %s", err)
		return errors.New("error generating remote playlist")
	}

	variantIndexStr := strconv.Itoa(info.VariantIndex)

	mywatcher.WritePlaylist(newPlaylist, filepath.Join(w.config.Transcode.PublicHLSPath, info.PublishName, variantIndexStr, info.Filename))
	return nil
}

func (w *MinIOFileWatcherStrategy) OnSegment(event watcher.Event) error {
	segment, err := getSegmentFromPath(event.Path)
	if segment == nil {
		logger.Errorf("error getting segment on segment: %s", err)
		return errors.New("error getting segment")
	}

	stream, ok := streams[segment.PublishName]
	if !ok {
		logger.Errorf("missing entry for publish name on segment: %s", segment.PublishName)
		return errors.New("missing entry for publish name")
	}

	variant := &(stream.Variants[segment.VariantIndex])
	newObjectPathChannel := make(chan string, 1)

	// if there is no remote storage method available, we dont do anything
	go func() {
		var newObjectPath string = event.Path
		var err error

		if w.storage != nil {
			newObjectPath, err = w.storage.AddSegment(event.Path, stream.PublishName, int(variant.VariantIndex))

			if err != nil {
				logger.Errorf("error while saving segments into storage", err)
			}
		}

		newObjectPathChannel <- newObjectPath
	}()
	newObjectPath := <-newObjectPathChannel

	segment.RemoteID = newObjectPath
	variant.Segments = append(variant.Segments, *segment)
	return nil
}

func (w *MinIOFileWatcherStrategy) OnThumbnail(event watcher.Event) error {
	publishName := filepath.Base(filepath.Dir(event.Path))
	stream, ok := streams[publishName]
	if !ok {
		logger.Errorf("missing entry for publish name on thumbnail: %s", publishName)
		return errors.New("missing entry for publish name")
	}

	if w.storage != nil {
		savedPath, err := w.storage.AddThumbnail(event.Path, stream.PublishName, "image/jpeg")

		if err != nil {
			logger.Errorf("error while saving thumbnail into storage", err)
		} else {
			logger.Debugf("saved thumbnail into %s", savedPath)
		}
	}

	return nil
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
