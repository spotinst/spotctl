package dep

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// download downloads a file from the given URL.
func download(url, path string) (err error) {
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
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("deps: download of %s failed with return code %d", url, resp.StatusCode)
		return err
	}

	// Writer the body to file.
	if _, err = io.Copy(out, resp.Body); err != nil {
		return err
	}

	// Make it executable.
	return os.Chmod(path, 0755)
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

// ungzip uncompresses a gzip archive.
func ungzip(source, target string) error {
	reader, err := os.Open(source)
	if err != nil {
		return err
	}
	defer reader.Close()

	archive, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer archive.Close()

	target = filepath.Join(target, archive.Name)
	writer, err := os.Create(target)
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = io.Copy(writer, archive)
	return err
}

// untar unpacks a tarball archive.
func untar(tarball, target string) error {
	reader, err := os.Open(tarball)
	if err != nil {
		return err
	}
	defer reader.Close()
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		path := filepath.Join(target, header.Name)
		info := header.FileInfo()

		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}

		if _, err = io.Copy(file, tarReader); err != nil {
			_ = file.Close()
			return err
		}

		_ = file.Close()
	}

	return nil
}

func userHomeDir() string {
	if runtime.GOOS == "windows" { // Windows
		return os.Getenv("USERPROFILE")
	}

	// *nix
	return os.Getenv("HOME")
}
