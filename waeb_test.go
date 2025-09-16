// waeb_test.go
package traefik_plugin_AdminAPI_WebUI

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestCreateConfig(t *testing.T) {
	config := CreateConfig()
	if config.Root != "." {
		t.Errorf("Expected Root to be '.', got %s", config.Root)
	}
}

func TestNeuter(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	neutered := neuter(handler)

	req := httptest.NewRequest("GET", "/test/", nil)
	rr := httptest.NewRecorder()
	neutered.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, status)
	}
}

func TestUploadFile(t *testing.T) {
	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Create a request to pass to our handler
	req := httptest.NewRequest("POST", "/upload", nil)
	req.Header.Set("Content-Type", "multipart/form-data")

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(uploadFile)

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}
}
