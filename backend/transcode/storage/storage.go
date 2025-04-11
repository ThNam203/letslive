package storage

import "context"

type Storage interface {
	// Save the file and return its final remote path
	AddSegment(ctx context.Context, filePath string, streamId string, qualityIndex int) (string, error)
	AddThumbnail(ctx context.Context, filePath string, streamId string, contentType string) (string, error)
}
