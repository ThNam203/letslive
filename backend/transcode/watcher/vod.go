package watcher

import "sen1or/letslive/transcode/domains"

type VODHandler interface {
	OnStreamStart(publishName string)
	OnStreamEnd(publishName string, publicHLSPath string, masterFileName string)
	OnGeneratingNewLineForRemotePlaylist(line string, variant domains.HLSVariant)
}
