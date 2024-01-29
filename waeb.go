// Upload config files
package traefik_plugin_waeb2

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// Config the plugin configuration.
type Config struct {
	Root string `json:"root,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		Root: ".",
	}
}

// New created a new Waeb plugin.
//
//	func New(_ context.Context, _ http.Handler, config *Config, _ string) (http.Handler, error) {
//		return http.FileServer(http.Dir(config.Root)), nil
//	}

func uploadFile(w http.ResponseWriter, r *http.Request, config *Config) {
	// Maximum upload of 10 MB files
	log.Print("Start upload file...")
	r.ParseMultipartForm(10 << 20)

	// Get handler for filename, size and headers
	file, handler, err := r.FormFile("file")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		log.Print("Error Retrieving the File")
		fmt.Println(err)
		log.Print(err)
		return
	}

	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	log.Print("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	log.Print("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)
	log.Print("MIME Header: %+v\n", handler.Header)

	// Create file
	dst, err := os.Create(fmt.Sprintf(config.Root) + handler.Filename)
	defer dst.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Copy the uploaded file to the created file on the filesystem
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Print("Successfully Uploaded File\n")
	fmt.Fprintf(w, "Successfully Uploaded File\n")
}

func New(_ context.Context, _ http.Handler, config *Config, _ string) (http.Handler, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case "POST":
			{
				uploadFile(w, r, config)

			}
		case "GET":
			{

				dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
				if err != nil {
					log.Fatal(err)
				}
				log.Print("current dir: " + dir)
				// log.Print("current dir: " + fmt.Sprintf(os.Executable()))
				// log.Print(os.Getwd())
				// log.Print("test")
				http.ServeFile(w, r, "plugins-local/src/github.com/Miromani4/traefik-plugin-waeb2/html/")
				// log.Print("test2")
			}

		}
	})

	log.Print(os.Executable())
	// index := http.FileServer(http.Dir(config.Root))
	// mux.Handle("/static/", http.StripPrefix("/static", index))
	return mux, nil
}
