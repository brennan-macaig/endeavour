// +build integration

package endeavour

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

const (
	usernameVar = "NEXUS_USER"
	passwordVar = "NEXUS_PASSWORD"
	urlVar      = "NEXUS_URL"
	repoVar     = "NEXUS_REPOSITORY"
)

func assertString(t *testing.T, name, got, expected string) {
	t.Helper()
	if got != expected {
		t.Fatalf("did not get correct %s, got %s, expected %s", name, got, expected)
	}
}

func calculateHashFromReader(t *testing.T, reader io.Reader) string {
	t.Helper()
	h := sha256.New()
	if _, err := io.Copy(h, reader); err != nil {
		t.Fatalf("unable to copy file contents into SHA256 reader - %s", err.Error())
	}
	return hex.EncodeToString(h.Sum(nil))
}

func calculateFileHashAsStringOrFail(t *testing.T, path string) string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("unable to open local file - %s", err.Error())
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		t.Fatalf("unable to copy file contents into SHA256 reader - %s", err.Error())
	}
	return hex.EncodeToString(h.Sum(nil))
}

func buildNexusFromEnvVarsAndRandomDataOrFail(t *testing.T) (Nexus, string, string) {
	var nex Nexus
	if val, ok := os.LookupEnv(usernameVar); ok {
		nex.Username = val
	} else {
		t.Fatalf("The username variable (%s) must be set.", usernameVar)
	}

	if val, ok := os.LookupEnv(passwordVar); ok {
		nex.Password = val
	} else {
		t.Fatalf("The password variable (%s) must be set.", passwordVar)
	}

	if val, ok := os.LookupEnv(urlVar); ok {
		nex.Url = val
	} else {
		t.Fatalf("The URL variable (%s) must be set.", urlVar)
	}

	if val, ok := os.LookupEnv(repoVar); ok {
		nex.Repo = val
	} else {
		t.Fatalf("The repository variable (%s) must be set.", repoVar)
	}

	dname, err := ioutil.TempDir("", "endeavourtmp-*")
	if err != nil {
		t.Fatalf("unable to create temporary directory - %s", err.Error())
	}

	tempFileName := make([]byte, 16)
	_, err = rand.Read(tempFileName)
	fname := filepath.Join(dname, hex.EncodeToString(tempFileName))

	someBuff := make([]byte, 7500000) // test file is 7.5mb of data
	_, err = rand.Read(someBuff)
	if err != nil {
		t.Fatalf("Could not generate random file contents - %s", err.Error())
	}

	err = ioutil.WriteFile(fname, someBuff, 0600)
	if err != nil {
		t.Fatalf("Could not write test file %s - %s", fname, err.Error())
	}
	info, err := os.Stat(fname)
	if err != nil {
		t.Fatalf("could not stat temp file %s - %s", fname, err.Error())
	}

	nex.Files = []string{filepath.Join(dname, info.Name())}
	nex.Path = "endeavour/upload-test"
	return nex, filepath.Join(dname, info.Name()), dname
}

func TestUpload(t *testing.T) {
	var hash string

	t.Run("UploadSingleFile", func(t *testing.T) {
		nex, filename, _ := buildNexusFromEnvVarsAndRandomDataOrFail(t)
		info, err := os.Stat(filename)
		if err != nil {
			t.Fatalf("could not stat temp file %s - %s", filename, err.Error())
		}

		url := nex.nexusURL(fileUploadInfo{
			FilePath:       info.Name(),
			IsUploadingDir: false,
		})

		t.Run("UploadFile", func(t *testing.T) {
			err := nex.Upload()
			if err != nil {
				t.Fatalf("unable to upload file %s - %s", filename, err.Error())
			}

			hash = calculateFileHashAsStringOrFail(t, filename)

			t.Run("DeleteLocalTestFile", func(t *testing.T) {
				err := os.Remove(filename)
				if err != nil {
					t.Fatalf("could not delete local file %s - %s", filename, err.Error())
				}
			})
		})
		t.Run("GetFileAndVerify", func(t *testing.T) {
			resp, err := http.Get(url)
			if err != nil {
				t.Fatalf("could not GET file from nexus - %s", err.Error())
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("status was not 200 - %s", resp.Status)
			}
			newHash := calculateHashFromReader(t, resp.Body)
			assertString(t, "file hash", newHash, hash)
		})

		t.Run("Cleanup", func(t *testing.T) {
			t.Run("DeleteFileFromNexus", func(t *testing.T) {
				client := &http.Client{}
				req, err := http.NewRequest("DELETE", url, nil)
				if err != nil {
					t.Fatalf("unable to spawn request - %s", err.Error())
				}
				req.SetBasicAuth(nex.Username, nex.Password)
				resp, err := client.Do(req)
				if err != nil {
					t.Fatalf("unable to perform request - %s", err.Error())
				}
				if resp.StatusCode != http.StatusNoContent {
					t.Fatalf("status was not 200 - status was %s", resp.Status)
				}
			})
		})
	})

	t.Run("UploadDirectoryAndContents", func(t *testing.T) {
		nex, filename, directory := buildNexusFromEnvVarsAndRandomDataOrFail(t)
		info, err := os.Stat(filename)
		if err != nil {
			t.Fatalf("could not stat temp file - %s", err.Error())
		}

		url := nex.nexusURL(fileUploadInfo{
			FilePath:       info.Name(),
			IsUploadingDir: true,
		})

		t.Run("UploadDirectory", func(t *testing.T) {
			nex.Files = []string{directory}
			err = nex.Upload()
			if err != nil {
				t.Fatalf("could not upload directory containing temp files - %s", err.Error())
			}
			hash = calculateFileHashAsStringOrFail(t, filename)

			t.Run("DeleteLocalTestFile", func(t *testing.T) {
				err := os.Remove(filename)
				if err != nil {
					t.Fatalf("could not delete local file %s - %s", filename, err.Error())
				}
			})
		})

		t.Run("GetFileAndVerify", func(t *testing.T) {
			resp, err := http.Get(url)
			if err != nil {
				t.Fatalf("could not GET file from nexus - %s", err.Error())
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("status was not 200 - %s", resp.Status)
			}
			newHash := calculateHashFromReader(t, resp.Body)
			assertString(t, "file hash", newHash, hash)
		})

		t.Run("Cleanup", func(t *testing.T) {
			t.Run("DeleteFileFromNexus", func(t *testing.T) {
				client := &http.Client{}
				req, err := http.NewRequest("DELETE", url, nil)
				if err != nil {
					t.Fatalf("unable to spawn request - %s", err.Error())
				}
				req.SetBasicAuth(nex.Username, nex.Password)
				resp, err := client.Do(req)
				if err != nil {
					t.Fatalf("unable to perform request - %s", err.Error())
				}
				if resp.StatusCode != http.StatusNoContent {
					t.Fatalf("status was not 200 - status was %s", resp.Status)
				}
			})
			t.Run("DeleteTmpDirectory", func(t *testing.T) {
				err = os.RemoveAll(directory)
				if err != nil {
					t.Fatalf("could not delete directory - %s", err.Error())
				}
			})
		})
	})
}
