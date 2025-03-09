package watcher

import (
	"path/filepath"
	"sen1or/lets-live/pkg/logger"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/radovskyb/watcher"
)

type FileWatcherStrategy interface {
	OnCreate(event watcher.Event)
	OnMaster(event watcher.Event)
	OnVariant(event watcher.Event)
	OnSegment(event watcher.Event)
}

type FFMpegFileWatcher struct {
	watcherStrategy FileWatcherStrategy
	monitorPath     string
}

func NewFFMpegFileWatcher(monitorPath string, watcherStrategy FileWatcherStrategy) *FFMpegFileWatcher {
	return &FFMpegFileWatcher{
		watcherStrategy: watcherStrategy,
		monitorPath:     monitorPath,
	}
}

func (w *FFMpegFileWatcher) SetStrategy(watcherStrategy FileWatcherStrategy) {
	w.watcherStrategy = watcherStrategy
}

func (w *FFMpegFileWatcher) Watch() {
	myWatcher := watcher.New()

	go func() {
		for {
			select {
			case event := <-myWatcher.Event:
				if event.Op == watcher.Remove {
					continue
				}

				if event.IsDir() && event.Op == watcher.Create {
					w.watcherStrategy.OnCreate(event)
					continue
				}

				fileType := w.getEventFileType(event.Path)
				if fileType == "Master" {
					w.watcherStrategy.OnMaster(event)
				} else if fileType == "Variant" {
					w.watcherStrategy.OnVariant(event)
				} else if fileType == "Segment" {
					w.watcherStrategy.OnSegment(event)
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
		logger.Panicf("error while setting up watcher path: %s", err)
	}

	if err := myWatcher.Start(time.Millisecond * 100); err != nil {
		logger.Panicf("error starting watcher: %s", err)
	}
}

// getFileType should return one of the three: Master, Variant, Segment
func (_ *FFMpegFileWatcher) getEventFileType(filePath string) string {
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
