package storage

type Storage interface {
	// Save the file and return its hash string
	AddFile(filePath string) (string, error)

	// Return the hash
	// AddDirectory(directoryName string) (string, error)
}
