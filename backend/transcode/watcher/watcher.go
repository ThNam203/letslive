package watcher

import (
	"context"
	"path/filepath"
	"sen1or/letslive/transcode/pkg/logger"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/radovskyb/watcher"
)

type FileWatcherStrategy interface {
	OnCreate(ctx context.Context, event watcher.Event) error
	OnMaster(ctx context.Context, event watcher.Event) error
	OnVariant(ctx context.Context, event watcher.Event) error
	OnSegment(ctx context.Context, event watcher.Event) error
	OnThumbnail(ctx context.Context, event watcher.Event) error
}

type FFMpegFileWatcher struct {
	watcherStrategy FileWatcherStrategy
	monitorPath     string
	watcher         *watcher.Watcher
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

func (w *FFMpegFileWatcher) Watch(ctx context.Context) {
	w.watcher = watcher.New()

	go func() {
		for {
			select {
			case event := <-w.watcher.Event:
				if event.Op == watcher.Remove {
					continue
				}

				if event.IsDir() && event.Op == watcher.Create {
					w.watcherStrategy.OnCreate(ctx, event)
					continue
				}

				fileType := w.getEventFileType(event.Path)
				if fileType == "Master" {
					w.watcherStrategy.OnMaster(ctx, event)
				} else if fileType == "Variant" {
					w.watcherStrategy.OnVariant(ctx, event)
				} else if fileType == "Segment" {
					w.watcherStrategy.OnSegment(ctx, event)
				} else if fileType == "Thumbnail" {
					w.watcherStrategy.OnThumbnail(ctx, event)
				} else {
					if fileType != "" {
						logger.Errorf(ctx, "unknown file appeared when watching: %s", fileType)
					}
				}
			case err := <-w.watcher.Error:
				logger.Errorf(ctx, "something failed while running watcher", err)
			case <-w.watcher.Closed:
				return
			}
		}
	}()

	// Watch the hls segment storage folder recursively for changes.
	if err := w.watcher.AddRecursive(w.monitorPath); err != nil {
		logger.Panicf(ctx, "error while setting up watcher path: %s", err)
	}

	if err := w.watcher.Start(time.Millisecond * 100); err != nil {
		logger.Panicf(ctx, "error starting watcher: %s", err)
	}
}

func (w FFMpegFileWatcher) Shutdown() {
	w.watcher.Close()
	logger.Infof(context.TODO(), "watcher has closed")
}

// getFileType should return one of the three: Master, Variant, Segment
func (_ *FFMpegFileWatcher) getEventFileType(filePath string) string {
	pathComponents := strings.Split(filePath, "/")
	fileExtension := filepath.Ext(filePath)

	logger.Debugf(context.TODO(), "get event file type: %s", filePath)

	switch fileExtension {
	case ".m3u8":
		{
			// the parent folder of Variant type is an index (1, 2, 3,...)
			if utf8.RuneCountInString(pathComponents[len(pathComponents)-2]) == 1 {
				return "Variant"
			}

			return "Master"
		}
	case ".ts":
		{
			return "Segment"
		}
	case ".jpeg", ".jpg":
		{
			return "Thumbnail"
		}
	default:
		{
			return fileExtension
		}
	}
}
