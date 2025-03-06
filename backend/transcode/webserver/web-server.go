package webserver

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sen1or/lets-live/pkg/logger"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type WebServer struct {
	ListenPort      int
	AllowedSuffixes []string
	BaseDirectory   string
}

func NewWebServer(listenPort int, allowedSuffixes []string, baseDirectory string) *WebServer {
	return &WebServer{
		ListenPort:      listenPort,
		AllowedSuffixes: allowedSuffixes,
		BaseDirectory:   baseDirectory,
	}
}

func sanitizeRequestPath(requestPath string) string {
	trimmedPath := strings.TrimSpace(requestPath)
	cleanedRequestPath := filepath.Clean(trimmedPath)
	return cleanedRequestPath
}

func (ws *WebServer) ListenAndServe() {
	router := mux.NewRouter()
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(ws.BaseDirectory))))
	router.HandleFunc("/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router.Use(corsMiddleware)

	server := &http.Server{
		Addr:         ":" + strconv.Itoa(ws.ListenPort),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go (func() {
		if err := server.ListenAndServe(); err != nil {
			logger.Errorf("failed to start web server: %s", err.Error())
		}
	})()

	logger.Infow("web server started", "port", ws.ListenPort)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Cache,Cache-Control,Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (ws *WebServer) serveFile(rw http.ResponseWriter, rq *http.Request) {
	requestPath := sanitizeRequestPath(rq.URL.Path)
	fileDestination := filepath.Join(ws.BaseDirectory, requestPath)

	if !strings.HasPrefix(fileDestination, ws.BaseDirectory) {
		http.Error(rw, "the destination is not allowed!", http.StatusForbidden)
		return
	}
	fileStat, err := os.Stat(fileDestination)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Errorf("webserver: File not found (%s)", err.Error())
			http.Error(rw, "file not found!", http.StatusNotFound)
		} else {
			http.Error(rw, "can't get the file information!", http.StatusInternalServerError)
		}
		return
	}

	if fileStat.IsDir() {
		http.Error(rw, "requested destination is not a file!", http.StatusForbidden)
		return
	}

	file, err := os.Open(fileDestination)
	defer file.Close()

	if err != nil {
		http.Error(rw, "can't open file!", http.StatusInternalServerError)
		return
	}

	fileExtension := filepath.Ext(fileDestination)
	if !slices.Contains(ws.AllowedSuffixes, fileExtension) {
		http.Error(rw, "file not allowed!", http.StatusForbidden)
		return
	}

	switch fileExtension {
	case ".ts":
		rw.Header().Set("Content-Type", "video/mp2t")
		rw.Header().Set("Cache-Control", "public, max-age=1800")
	case ".m3u8":
		rw.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		rw.Header().Set("Cache-Control", "public, no-store, max-age=0")
	default:
		log.Fatal("File extension not supported!")
		http.Error(rw, "file extension not supported!", http.StatusForbidden)
		return
	}

	rw.Header().Set("Content-Length", strconv.FormatInt(fileStat.Size(), 10))
	io.Copy(rw, file)
}
