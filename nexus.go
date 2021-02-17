package endeavour

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Nexus struct {
	Url         string
	Username    string
	Password    string
	Files       []string
	Repo        string
	Path        string
	Verbose     bool
	currentPath string
}

// This function drives the uploading of _any_ number of files.
// It takes no arguments, and instead takes Nexus as a receiver.
func (o *Nexus) Upload() error {
	// Iterate through all files / dirs
	// HTTP POST file with basic auth
	// progress loop.
	if o.Url == "" {
		return fmt.Errorf("a URL must be set")
	}
	if o.Username == "" {
		return fmt.Errorf("nexus username must be set")
	}
	if o.Password == "" {
		return fmt.Errorf("nexus password must be set")
	}
	if len(o.Files) < 1 {
		return fmt.Errorf("files to upload must be provided")
	}
	if o.Repo == "" {
		return fmt.Errorf("a repo must be set")
	}
	if o.Path == "" {
		return fmt.Errorf("a path must be set")
	}

	for _, filePath := range o.Files {
		info, err := os.Stat(filePath)
		if err != nil {
			return fmt.Errorf("could not stat %s - %w", filePath, err)
		}
		o.currentPath = filePath
		if info.IsDir() {
			// if this is a directory, walk it.
			err := filepath.Walk(filePath, o.uploadDirectory)
			if err != nil {
				return fmt.Errorf("failed to upload files in %s  - %w", filePath, err)
			}
		} else {
			// if this is just a file, just write it to the server
			err = o.writeFileToServer(filePath, o.nexusURL(fileUploadInfo{
				FilePath:       info.Name(),
				IsUploadingDir: false,
			}))
			if err != nil {
				return fmt.Errorf("could not upload single file %s - %w", filePath, err)
			}
		}
	}
	return nil
}

// Callback function called by filepath.Walk()
// This function has some defined behavior:
// 1. If it encounters an error from a previous file/directory, skip the current file/directory.
// 2. If it encounters a directory, simply step into the directory and do nothing.
// 3. If it encounters a file, it will write that file to the server.
//
// If this function ever finds an error, it will force all the remaining callbacks to be skipped,
// so no files from the current argument, or any argument after this one, will be skipped.
func (o Nexus) uploadDirectory(aPath string, info os.FileInfo, err error) error {
	if err != nil {
		// Whatever the error was, we just want to skip the rest of the files.
		return fmt.Errorf("error: could not access file/dir - %w", err)
	}

	// If this is a directory, skip it.
	if info.IsDir() {
		if o.Verbose {
			log.Printf("%s was a directory, stepping in", info.Name())
		}
		return nil
	}

	if o.currentPath == "" {
		return fmt.Errorf("current root path cannot be empty")
	}

	url := o.nexusURL(fileUploadInfo{
		FilePath:       strings.TrimPrefix(filepath.ToSlash(aPath), filepath.ToSlash(o.currentPath)),
		IsUploadingDir: true,
	})

	err = o.writeFileToServer(aPath, url)
	if err != nil {
		return fmt.Errorf("could not upload file - %w", err)
	}
	return nil
}

type fileUploadInfo struct {
	FilePath       string
	IsUploadingDir bool
}

// Generates a URL to upload the file to.
// This function is pretty hard-coded for Nexus.
// It normalizes the slashes, then constructs a URL based on the file path and upload mode.
func (o Nexus) nexusURL(info fileUploadInfo) string {
	urlBuilder := strings.Builder{}
	urlBuilder.WriteString(o.Url)
	urlBuilder.WriteString("/")
	urlBuilder.WriteString(o.Repo)
	urlBuilder.WriteString("/")
	urlBuilder.WriteString(o.Path)
	urlBuilder.WriteString("/")
	urlBuilder.WriteString(filepath.ToSlash(info.FilePath))
	return urlBuilder.String()
}

// writeFileToServer simply uploads the file to Nexus using an HTTP PUT with basic auth.
func (o Nexus) writeFileToServer(path, url string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("could not open file %s - %w", path, err)
	}
	defer f.Close()

	req, err := http.NewRequest("PUT", url, f)
	if err != nil {
		return fmt.Errorf("could not form http request for file %s - %w", f.Name(), err)
	}
	req.SetBasicAuth(o.Username, o.Password)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("could not upload %s - %w", f.Name(), err)
	}
	defer resp.Body.Close()

	err = checkError(resp)
	if err != nil {
		return fmt.Errorf("could not upload %s - %w", f.Name(), err)
	}
	if o.Verbose {
		log.Printf("got HTTP %d for %s - written to server", resp.StatusCode, f.Name())
	}
	return nil
}

func checkError(response *http.Response) error {
	if response.StatusCode == http.StatusCreated {
		return nil
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("response was %d, expected 201; could not read response body - %w",
			response.StatusCode, err)
	}
	if len(body) > 0 {
		return fmt.Errorf("response was %d, expected 201; response body: %s", response.StatusCode, body)
	} else {
		return fmt.Errorf("response was %d, expected 201; there was no response body", response.StatusCode)
	}
}
