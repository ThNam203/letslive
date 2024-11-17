package watcher

// TODO: use this function again after implement the Storage interface for the CustomStorage
// the main difference is that the KuboStorage allows to add a folder while the Custome one does not
// see more on lite_watcher.go

//func MonitorHLSStreamContent(monitorPath string, storage Storage) {
//	w := watcher.New()
//
//	go func() {
//		for {
//			select {
//			case event := <-w.Event:
//				if event.Op == watcher.Remove {
//					continue
//				}
//
//				// NOT SURE WHAT IS CREATED FIRST
//				// handle creating the publish name folder
//				if event.IsDir() && event.Op == watcher.Create {
//					components := strings.Split(event.Path, "/")
//					publishName := components[len(components)-1]
//
//					// TODO: properly handle
//					if utf8.RuneCountInString(publishName) < 10 {
//						continue
//					}
//
//					if err := os.MkdirAll(filepath.Join(cfg.PublicHLSPath, publishName), os.ModePerm); err != nil {
//						logger.Errorw("failed to create publish folder", "path", filepath.Join(cfg.PublicHLSPath, publishName))
//					} else {
//						logger.Infof("created publish folder: %s", filepath.Join(cfg.PublicHLSPath, publishName))
//					}
//
//					publishFolderHash, err := storage.AddDirectory(event.Path)
//					if err != nil {
//						fmt.Println("failed to add publish folder into storage")
//						return
//					}
//
//					variants := make([]HLSVariant, len(cfg.FFMpegSetting.Qualities))
//					for index := range variants {
//						variants[index] = HLSVariant{uint8(index), make([]HLSSegment, 0)}
//					}
//
//					streams[publishName] = HLSStream{
//						PublishName:           publishName,
//						Variants:              variants,
//						PublishFolderRemoteId: publishFolderHash,
//					}
//
//					log.Printf("created hls stream %+v\n", streams[publishName])
//
//					continue
//				}
//
//				fileType := getEventFileType(event.Path)
//				if fileType == "Master" {
//					components := strings.Split(event.Path, "/")
//					pushlishName := components[len(components)-2]
//
//					if err := copy(event.Path, filepath.Join(cfg.PublicHLSPath, pushlishName, cfg.FFMpegSetting.MasterFileName)); err != nil {
//						log.Panicf("failed to copy file: %s", err)
//					}
//				} else if fileType == "Variant" {
//					info, err := getInfoFromPath(event.Path)
//					if err != nil {
//						fmt.Println(err)
//						continue
//					}
//					variant := streams[info.PublishName].Variants[info.VariantIndex]
//					newPlaylist, err := storage.GenerateRemotePlaylist(event.Path, variant)
//					if err != nil {
//						fmt.Println("error generating remote playlist")
//						continue
//					}
//
//					variantIndexStr := strconv.Itoa(info.VariantIndex)
//
//					writePlaylist(newPlaylist, filepath.Join(cfg.PublicHLSPath, info.PublishName, variantIndexStr, info.Filename))
//				} else if fileType == "Segment" {
//					segment := getSegmentFromPath(event.Path)
//					if segment == nil {
//						log.Printf("error creating segment")
//						continue
//					}
//
//					variant := &(streams[segment.PublishName].Variants[segment.VariantIndex])
//
//					newObjectPathChannel := make(chan string, 1)
//					go func() {
//						newObjectPath, err := storage.SaveIntoHLSDirectory(event.Path)
//
//						if err != nil {
//							fmt.Printf("error while saving segments into ipfs: %s\n", err)
//						}
//
//						newObjectPathChannel <- newObjectPath
//					}()
//					newObjectPath := <-newObjectPathChannel
//
//					segment.IPFSRemoteId = newObjectPath
//					variant.Segments = append(variant.Segments, *segment)
//				}
//			case err := <-w.Error:
//				log.Panicf("something failed while running watcher: %s", err)
//			case <-w.Closed:
//				return
//			}
//		}
//	}()
//
//	// Watch the hls segment storage folder recursively for changes.
//	if err := w.AddRecursive(monitorPath); err != nil {
//		log.Fatalln(err)
//	}
//
//	if err := w.Start(time.Millisecond * 100); err != nil {
//		log.Fatalln(err)
//	}
//}
