package webserver

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

type WebServer struct {
	ListenAddr      string
	AllowedSuffixes []string
	BaseDirectory   string
}

func sanitizeRequestPath(requestPath string) string {
	trimmedPath := strings.TrimSpace(requestPath)
	cleanedRequestPath := filepath.Clean(trimmedPath)
	return cleanedRequestPath
}

func (ws *WebServer) serveFile(rw http.ResponseWriter, rq *http.Request) {
	requestPath := sanitizeRequestPath(rq.URL.Path)
	fileDestination := filepath.Join(ws.BaseDirectory, requestPath)

	if !strings.HasPrefix(fileDestination, ws.BaseDirectory) {
		http.Error(rw, "The destination is not allowed!", http.StatusForbidden)
		return
	}
	fileStat, err := os.Stat(fileDestination)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(rw, "File not found!", http.StatusNotFound)
		} else {
			http.Error(rw, "Can't get the file information!", http.StatusInternalServerError)
		}
		return
	}

	if fileStat.IsDir() {
		http.Error(rw, "Requested destination is not a file!", http.StatusForbidden)
		return
	}

	file, err := os.Open(fileDestination)
	defer file.Close()

	if err != nil {
		http.Error(rw, "Can't open file!", http.StatusInternalServerError)
		return
	}

	fileExtension := filepath.Ext(fileDestination)
	if !slices.Contains(ws.AllowedSuffixes, fileExtension) {
		http.Error(rw, "File not allowed!", http.StatusForbidden)
		return
	}

	switch fileExtension {
	case ".ts":
		rw.Header().Set("Content-Type", "video/mp2t")
	case ".m3u8":
		rw.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	default:
		log.Fatal("File extension not supported!")
	}

	rw.Header().Set("Content-Length", strconv.FormatInt(fileStat.Size(), 10))
	io.Copy(rw, file)
}

func (ws *WebServer) ListenAndServe() {
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/", ws.serveFile)
	go (func() {
		log.Fatal(http.ListenAndServe(ws.ListenAddr, serveMux))
	})()
	log.Println("web server started")
}

func NewWebServer(listenAddr string, allowedSuffixes []string, baseDirectory string) *WebServer {
	return &WebServer{
		ListenAddr:      listenAddr,
		AllowedSuffixes: allowedSuffixes,
		BaseDirectory:   baseDirectory,
	}
}
