package storage

type Storage interface {
	// Save the file and return its final remote path
	AddSegment(filePath string, streamId string, qualityIndex int) (string, error)
	AddThumbnail(filePath string, streamId string, contentType string) (string, error)
}
