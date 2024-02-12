// Upload config files
package traefik_plugin_AdminAPI_WebUI

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

var html_root = "/admin_panel/html/"

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
//
//	func New(_ context.Context, _ http.Handler, config *Config, _ string) (http.Handler, error) {
//		return http.FileServer(http.Dir(config.Root)), nil
//	}
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
	dst, err := os.Create(fmt.Sprintf(conf) + handler.Filename)
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

var (
	fileName    string
	fullURLFile string
)

func dl_file() {

	fullURLFile = "https://github.com/Miromani4/traefik-plugin-AdminAPI_WebUI/releases/download/v.1.1.0/web_panel.zip"
	log.Print("start dl file...")
	// Build fileName from fullPath
	fileURL, err := url.Parse(fullURLFile)
	if err != nil {
		log.Print(err)
	}
	path := fileURL.Path
	segments := strings.Split(path, "/")
	fileName = segments[len(segments)-1]

	// Create blank file
	file, err := os.Create(html_root + fileName)
	if err != nil {
		log.Print("File exist!")
	}
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
	}
	defer resp.Body.Close()

	size, err := io.Copy(file, resp.Body)
	if err != nil {
		log.Print(err)
	}
	defer file.Close()

	log.Print("Downloaded a file ", fileName, " with size: ", size)
	unzip()
}
func unzip() {
	dst := html_root
	archive, err := zip.OpenReader(html_root + "web_panel.zip")
	if err != nil {
		panic(err)
	}
	defer archive.Close()

	for _, f := range archive.File {
		filePath := filepath.Join(dst, f.Name)
		fmt.Println("unzipping file ", filePath)

		if !strings.HasPrefix(filePath, filepath.Clean(dst)+string(os.PathSeparator)) {
			fmt.Println("invalid file path")
			return
		}
		if f.FileInfo().IsDir() {
			fmt.Println("creating directory...")
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			panic(err)
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			panic(err)
		}

		fileInArchive, err := f.Open()
		if err != nil {
			panic(err)
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			panic(err)
		}

		dstFile.Close()
		fileInArchive.Close()

	}
}

func New(_ context.Context, _ http.Handler, config *Config, _ string) (http.Handler, error) {
	if _, err := os.Stat(html_root); os.IsNotExist(err) {
		err := os.MkdirAll(html_root, 0777)
		log.Print(err)
		dl_file()
	}
	conf = config.Root
	mux := http.NewServeMux()
	mux.HandleFunc("/", root)

	// mux.HandleFunc("/api", apis)
	fs := http.FileServer(http.Dir(html_root + "/static"))
	mux.Handle("/static/", http.StripPrefix("/static", neuter(fs)))
	// mux.HandleFunc("/static", func(w http.ResponseWriter, r *http.Request) {
	// 	http.ServeFile(w, r, html_root+"/static/css/style.css")
	// })
	return mux, nil
}
func root(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		{
			http.ServeFile(w, r, html_root)
		}
	case "/api", "/api/":
		{
			apis(w, r)
		}
	// case "/static/":
	// 	{
	// 		fs := http.FileServer(http.Dir(html_root + "/static"))
	// 		http.Handle("/static/", http.StripPrefix("/static/", fs))
	// 	}
	default:
		{
			errorHandler(w, r, http.StatusNotFound)
		}
	}
	// switch r.Method {
	// case http.MethodPost:
	// 	{
	// 		uploadFile(w, r)
	// 	}
	// case http.MethodGet:
	// 	{
	// 		http.ServeFile(w, r, html_root)
	// 	}

	// }
}

func apis(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		{
			if r.Header.Get("Rewrite") != "" {
				body2, err4 := io.ReadAll(r.Body)
				if err4 != nil {
					//log.Fatalf("ERROR: %s", err4)
					errorHandler(w, r, http.StatusBadRequest)
					return
				}
				//fmt.Fprint(w, string(body2))
				f, err := os.OpenFile(conf+"/"+r.Header.Get("Rewrite"), os.O_APPEND|os.O_WRONLY, 0777)
				if err != nil {
					//log.Fatal(err)
					errorHandler(w, r, http.StatusBadRequest)
					return
				}

				defer f.Close()
				err = f.Truncate(0)
				_, err = f.Seek(0, 0)
				// _, err = fmt.Fprintf(f, "%d", len(b))
				_, err2 := f.WriteString(string(body2))

				if err2 != nil {
					//log.Fatal(err2)
					errorHandler(w, r, http.StatusBadRequest)
					return
				}

				errorHandler(w, r, http.StatusAccepted)
			} else {
				switch r.FormValue("atr") {
				case "list":
					{
						fmt.Fprint(w, list_file())
					}
				case "open":
					{
						if r.FormValue("file") != "" {
							log.Print(string(openfile(r.FormValue("file"))))
							fmt.Fprint(w, string(openfile(r.FormValue("file"))))
						} else {
							errorHandler(w, r, http.StatusBadRequest)
							return
						}

					}
				case "create":
					{
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
				case "upload":
					{
						uploadFile(w, r)

					}
				case "delete":
					{
						err := os.Remove(conf + "/" + r.FormValue("file"))
						if err != nil {
							errorHandler(w, r, http.StatusBadRequest)
							return
						}
						errorHandler(w, r, http.StatusAccepted)
					}
				default:
					{
						errorHandler(w, r, http.StatusBadRequest)
						return
					}
				}
			}
		}

	default:
		{
			errorHandler(w, r, http.StatusNotFound)
			return
		}

	}

}

func openfile(file string) (content []byte) {
	content, err := os.ReadFile(conf + "/" + file)

	if err != nil {
		log.Fatal(err)
	}

	//fmt.Println(string(content))
	return content
}

func list_file() []string {
	var s []string
	var linght int
	entries, err := os.ReadDir(conf)
	if err != nil {
		log.Fatal(err)
	}
	linght = len(entries) - 1
	s = append(s, "{")
	for z, e := range entries {
		if z != linght {
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
