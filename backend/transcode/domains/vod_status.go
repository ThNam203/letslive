package domains

type VODStatus string

const (
	VODStatusUploading  VODStatus = "uploading"
	VODStatusProcessing VODStatus = "processing"
	VODStatusReady      VODStatus = "ready"
	VODStatusFailed     VODStatus = "failed"
)
