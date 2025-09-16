// waeb.go
package traefik_plugin_adminapi_webui

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var htmlRoot = "/admin_panel/html/"

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

var conf string

// New created a new AdminAPI & WebUI plugin.
func New(_ context.Context, _ http.Handler, config *Config, _ string) (http.Handler, error) {
	if _, err := os.Stat(htmlRoot); os.IsNotExist(err) {
		err := os.MkdirAll(htmlRoot, 0o777)
		log.Print(err)
		dlFile()
	}
	conf = config.Root
	mux := http.NewServeMux()
	mux.HandleFunc("/", root)

	fs := http.FileServer(http.Dir(htmlRoot + "/static"))
	mux.Handle("/static/", http.StripPrefix("/static", neuter(fs)))
	return mux, nil
}

func neuter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	// Maximum upload of 10 MB files
	log.Print("Start upload file...")
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		log.Print("Error parsing multipart form: ", err)
		http.Error(w, "Error parsing multipart form", http.StatusInternalServerError)
		return
	}

	// Get handler for filename, size and headers
	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Print("Error Retrieving the File: ", err)
		http.Error(w, "Error Retrieving the File", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	log.Printf("Uploaded File: %+v\n", handler.Filename)
	log.Printf("File Size: %+v\n", handler.Size)
	log.Printf("MIME Header: %+v\n", handler.Header)

	// Create file
	dst, err := os.Create(fmt.Sprintf(conf) + handler.Filename)
	if err != nil {
		log.Print("Error creating file: ", err)
		http.Error(w, "Error creating file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy the uploaded file to the created file on the filesystem
	if _, err := io.Copy(dst, file); err != nil {
		log.Print("Error copying file: ", err)
		http.Error(w, "Error copying file", http.StatusInternalServerError)
		return
	}

	log.Print("Successfully Uploaded File\n")
	fmt.Fprintf(w, "Successfully Uploaded File\n")
}

func dlFile() {
	fullURLFile := "https://github.com/Miromani4/traefik-plugin-AdminAPI_WebUI/releases/download/v1.1.0/web_panel.zip"
	log.Print("start dl file...")
	// Build fileName from fullPath
	fileURL, err := url.Parse(fullURLFile)
	if err != nil {
		log.Print(err)
		return
	}
	path := fileURL.Path
	segments := strings.Split(path, "/")
	fileName := segments[len(segments)-1]

	// Create blank file
	file, err := os.Create(htmlRoot + fileName)
	if err != nil {
		log.Print("File exist!")
		return
	}
	defer file.Close()

	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	// Put content on file
	resp, err := client.Get(fullURLFile)
	if err != nil {
		log.Print(err)
		return
	}
	defer resp.Body.Close()

	size, err := io.Copy(file, resp.Body)
	if err != nil {
		log.Print(err)
		return
	}

	log.Print("Downloaded a file ", fileName, " with size: ", size)
	unzip()
}

func unzip() {
	dst := htmlRoot
	archive, err := zip.OpenReader(htmlRoot + "web_panel.zip")
	if err != nil {
		log.Print("Error opening zip file: ", err)
		return
	}
	defer archive.Close()

	for _, fileInArchive := range archive.File {
		filePath := filepath.Join(dst, fileInArchive.Name)
		log.Print("unzipping file ", filePath)

		if !strings.HasPrefix(filePath, filepath.Clean(dst)+string(os.PathSeparator)) {
			log.Print("invalid file path")
			return
		}
		if fileInArchive.FileInfo().IsDir() {
			log.Print("creating directory...")
			err := os.MkdirAll(filePath, os.ModePerm)
			if err != nil {
				log.Print("Error creating directory: ", err)
				return
			}
			continue
		}

		err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
		if err != nil {
			log.Print("Error creating directory: ", err)
			return
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fileInArchive.Mode())
		if err != nil {
			log.Print("Error opening file: ", err)
			return
		}

		fileInArchiveFile, err := fileInArchive.Open()
		if err != nil {
			log.Print("Error opening file in archive: ", err)
			return
		}

		if _, err := io.Copy(dstFile, fileInArchiveFile); err != nil {
			log.Print("Error copying file: ", err)
			dstFile.Close()
			fileInArchiveFile.Close()
			return
		}

		dstFile.Close()
		fileInArchiveFile.Close()
	}
}

func root(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		{
			http.ServeFile(w, r, htmlRoot)
		}
	case "/api", "/api/":
		{
			apis(w, r)
		}
	default:
		{
			errorHandler(w, r, http.StatusNotFound)
		}
	}
}

func apis(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handlePostRequest(w, r)
	default:
		errorHandler(w, r, http.StatusNotFound)
	}
}

func handlePostRequest(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Rewrite") != "" {
		handleRewriteRequest(w, r)
	} else {
		handleFormRequest(w, r)
	}
}

func handleRewriteRequest(w http.ResponseWriter, r *http.Request) {
	body2, err4 := io.ReadAll(r.Body)
	if err4 != nil {
		errorHandler(w, r, http.StatusBadRequest)
		return
	}

	f, err := os.OpenFile(conf+"/"+r.Header.Get("Rewrite"), os.O_APPEND|os.O_WRONLY, 0o777)
	if err != nil {
		errorHandler(w, r, http.StatusBadRequest)
		return
	}
	defer f.Close()

	err = f.Truncate(0)
	if err != nil {
		errorHandler(w, r, http.StatusBadRequest)
		return
	}

	_, err = f.Seek(0, 0)
	if err != nil {
		errorHandler(w, r, http.StatusBadRequest)
		return
	}

	_, err2 := f.WriteString(string(body2))
	if err2 != nil {
		errorHandler(w, r, http.StatusBadRequest)
		return
	}

	errorHandler(w, r, http.StatusAccepted)
}

func handleFormRequest(w http.ResponseWriter, r *http.Request) {
	switch r.FormValue("atr") {
	case "list":
		fmt.Fprint(w, listFile())
	case "open":
		handleOpenRequest(w, r)
	case "create":
		handleCreateRequest(w, r)
	case "upload":
		uploadFile(w, r)
	case "delete":
		handleDeleteRequest(w, r)
	default:
		errorHandler(w, r, http.StatusBadRequest)
	}
}

func handleOpenRequest(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("file") != "" {
		log.Print(string(openFile(r.FormValue("file"))))
		fmt.Fprint(w, string(openFile(r.FormValue("file"))))
	} else {
		errorHandler(w, r, http.StatusBadRequest)
	}
}

func handleCreateRequest(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("file") != "" {
		myfile, err := os.Create(conf + "/" + r.FormValue("file"))
		if err != nil {
			log.Fatal(err)
			errorHandler(w, r, http.StatusBadRequest)
		}
		errorHandler(w, r, http.StatusAccepted)
		log.Println(myfile)
		myfile.Close()
	}
}

func handleDeleteRequest(w http.ResponseWriter, r *http.Request) {
	err := os.Remove(conf + "/" + r.FormValue("file"))
	if err != nil {
		errorHandler(w, r, http.StatusBadRequest)
		return
	}
	errorHandler(w, r, http.StatusAccepted)
}

func openFile(file string) []byte {
	content, err := os.ReadFile(conf + "/" + file)
	if err != nil {
		log.Fatal(err)
	}
	return content
}

func listFile() []string {
	var s []string
	var length int
	entries, err := os.ReadDir(conf)
	if err != nil {
		log.Fatal(err)
	}
	length = len(entries) - 1
	s = append(s, "{")
	for z, e := range entries {
		if z != length {
			s = append(s, "\""+strconv.Itoa(z)+"\":\""+e.Name()+"\",")
		} else {
			s = append(s, "\""+strconv.Itoa(z)+"\":\""+e.Name()+"\"")
		}
	}
	s = append(s, "}")
	return s
}

func errorHandler(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)
	switch status {
	case http.StatusNotFound:
		{
			fmt.Fprint(w, "<h1>Page not found 404</h1>")
		}
	case http.StatusBadRequest:
		{
			fmt.Fprint(w, "<h1>Bad request 400</h1>")
		}
	case http.StatusAccepted:
		{
			fmt.Fprint(w, "<h1>Accepted 202</h1>")
		}
	}
}
