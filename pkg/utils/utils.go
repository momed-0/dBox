package utils

import (
	"os"
	"io"
	"log"
	"path/filepath"
	"os/exec"
	"fmt"
	"encoding/json"
	"net/http"
	"io/ioutil"
	"archive/tar"
	"compress/gzip"
)

func WriteToFile(filename, message string) {
	err := os.WriteFile(filename, []byte(message), 0644)
	if err != nil {
		log.Fatalf("failed to write to file: %v", err)
	}
}

func ChangeOwnershipImage(imageDir string) error {
	/*	
		change owner ship from root to user
	*/
	log.Printf("Trying to change the ownership of %s to user..",imageDir)
	cmd := exec.Command("chown", "1000:1000", "-R", imageDir)
	// Run the command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Error executing chown command: %v\n", err)
	}

	log.Printf("Ownership changed successfully")
	return nil
}

// parseJSONResponse parses the JSON response body into the provided struct
func ParseJSONResponse(resp *http.Response, target interface{}) error {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("failed to parse JSON response: %w", err)
	}
	return nil
}

// extractTarFile extracts a single tar file based on its type
func ExtractTarFile(target string, header *tar.Header, tarReader *tar.Reader) error {
	switch header.Typeflag {
	case tar.TypeDir:
		return os.MkdirAll(target, 0755)
	case tar.TypeReg:
		file, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
		if err != nil {
			return fmt.Errorf("failed to create file: %v", err)
		}
		defer file.Close()

		if _, err := io.Copy(file, tarReader); err != nil {
			return fmt.Errorf("failed to write file: %v", err)
		}
	case tar.TypeSymlink:
		// Handle symbolic links explicitly
		// Check if the target already exists
		if _, err := os.Lstat(target); err == nil {
			// Target exists, skip creating symlink or handle conflict
			log.Println("Symlink already exists, skipping:", target)
			return nil // Skip this case, or handle it in another way
		}
		if err := os.Symlink(header.Linkname, target); err != nil {
			return fmt.Errorf("failed to create symlink: %v", err)
		}
	}
	return nil
}

func ExtractLayer(destDir string, layerData io.Reader) error {
	gz, err := gzip.NewReader(layerData)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %v", err)
	}
	defer gz.Close()

	tarReader := tar.NewReader(gz)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %v", err)
		}
		//create the directory 
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory:%q : %v",destDir, err)
		}
		target := filepath.Join(destDir, header.Name)
		log.Printf("Creating %s\n", target)
		if err := ExtractTarFile(target, header, tarReader); err != nil {
			return err
		}
	}
	return ChangeOwnershipImage(destDir)
}