package core

type Storage interface {
	// // Save the file and return its hash string
	//Save(filePath string) (string, error)

	// The function save the file and add into a hls directory
	// Return final hash after adding file hash into directory
	SaveIntoHLSDirectory(filePath string) (string, error)

	// Generate (or get the new remote playlist from the original local playlist)
	GenerateRemotePlaylist(playlistPath string, variant HLSVariant) (string, error)

	// Return the hash
	AddDirectory(directoryName string) (string, error)
}
