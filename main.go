package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/webdav"
)

func main() {
	// Command-line flags for port, directories, username, and password
	port := flag.Int("port", 8080, "Port to serve WebDAV on")
	dirs := flag.String("dirs", "./webdav1,./webdav2,./webdav3", "Comma-separated list of directories to mirror (first one is primary)")
	username := flag.String("username", "admin", "Username for Basic Auth")
	password := flag.String("password", "password", "Password for Basic Auth")

	flag.Parse()

	directories := strings.Split(*dirs, ",")
	if len(directories) < 1 {
		log.Fatal("At least one directory must be specified")
	}

	var filesystems []webdav.FileSystem
	for _, dir := range directories {
		dir = strings.TrimSpace(dir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		filesystems = append(filesystems, webdav.Dir(dir))
		log.Printf("Added directory: %s\n", dir)
	}

	multiFS := &MultiFS{
		filesystems: filesystems,
	}

	handler := &webdav.Handler{
		FileSystem: multiFS,
		LockSystem: webdav.NewMemLS(),
		Logger: func(r *http.Request, err error) {
			if err != nil {
				log.Printf("WebDAV %s: %s, ERROR: %s\n", r.Method, r.URL, err)
			} else {
				log.Printf("WebDAV %s: %s\n", r.Method, r.URL)
			}
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Authenticate the user using Basic Auth
		user, pass, ok := r.BasicAuth()
		if !ok || user != *username || pass != *password {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		handler.ServeHTTP(w, r)
	})

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting WebDAV server on %s\n", addr)
	log.Printf("Primary directory: %s\n", directories[0])
	log.Printf("Mirror directories: %s\n", strings.Join(directories[1:], ", "))
	log.Printf("Using Basic Auth credentials: username=%s\n", *username)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
