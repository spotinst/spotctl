package dep

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// download downloads a file from the given URL.
func download(url *url.URL, path string) (err error) {
	// Create the target directory, if needed.
	parent := filepath.Dir(path)
	if err := mkdirAll(parent); err != nil {
		return err
	}

	// Create the file.
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	client := &http.Client{
		Timeout: time.Hour,
	}

	// Get the data.
	resp, err := client.Get(url.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("deps: download of %s failed with return code %d", url, resp.StatusCode)
		return err
	}

	// Writer the body to file.
	_, err = io.Copy(out, resp.Body)
	return err
}

// mkdirAll creates a directory named path. If path is already a directory,
// mkdirAll does nothing and returns nil.
func mkdirAll(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

// copyFile copies file source src to destination dst.
func copyFile(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	ss, err := s.Stat()
	if err != nil {
		return err
	}

	if err = mkdirAll(filepath.Dir(dst)); err != nil {
		return err
	}

	d, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, ss.Mode())
	if err != nil {
		return err
	}

	if _, err := io.Copy(d, s); err != nil {
		d.Close()
		return err
	}

	if err := d.Close(); err != nil {
		return err
	}

	// io.Copy can restrict file permissions based on umask.
	return os.Chmod(dst, ss.Mode())
}

// See: https://en.wikipedia.org/wiki/List_of_archive_formats
func checkArchive(path string) (string, bool) {
	for _, suffix := range []string{
		".tar.gz", ".tgz", // application/x-gtar
		".zip", // application/zip
	} {
		if strings.HasSuffix(path, suffix) {
			return suffix, true
		}
	}
	return "", false
}

func userHomeDir() string {
	if runtime.GOOS == "windows" { // Windows
		return os.Getenv("USERPROFILE")
	}

	// *nix
	return os.Getenv("HOME")
}
