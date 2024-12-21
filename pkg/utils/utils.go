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
	"errors"
	"dBox/pkg/model"
	"strconv"
)

func WriteToFile(filename, message string) {
	err := os.WriteFile(filename, []byte(message), 0644)
	if err != nil {
		log.Fatalf("failed to write to file: %v", err)
	}
}

func metaDataTemplate(destination,imageName string) error {
	//create metadata.json
	_, err := os.Create(destination)
	if err != nil {
		return err
	}
	template := model.Image{
		Image_Name: imageName,
	}
	file, _ := json.MarshalIndent(template, "", " ")

	_ = os.WriteFile(destination, file, 0644)
	return nil
}


func WriteMetadata(destination,imageName,tag string) error{
	metaPath :=  filepath.Join(destination,"metadata.json")
	//check if the file exists, if not create metadata template
	if _, err := os.Stat(metaPath); errors.Is(err, os.ErrNotExist) {
		// metadata.json doesn't exists, first time pulling this image
		//create the template
		if err := metaDataTemplate(metaPath,imageName); err != nil {
			return fmt.Errorf("Error trying to write metadata.json template for first time:  %v", err)
		}
	 }
	// now metadata.json exists , parse the json and append it and save it back
	data, err := ioutil.ReadFile(metaPath )
	if err != nil {
		return fmt.Errorf("Error trying to open metadata.json file:  %v", err)
	}
	var image model.Image
	err = json.Unmarshal(data, &image)
	if err != nil {
		return fmt.Errorf("Error unmarshaling metadata.json: %v", err)
	}
	//if the imagetag is nil initialize the tag array else append
	if image.Image_Tag == nil {
		image.Image_Tag = []string{tag}
	} else {
		image.Image_Tag = append(image.Image_Tag, tag)
	}
	image.Latest_Tag = findLatestTagNumOrder(image.Latest_Tag,tag)
	updatedData, err := json.MarshalIndent(image, "", "  ")
	if err != nil {
		return fmt.Errorf("Error marshaling metadata.json: %v", err)
	}

	err = os.WriteFile(metaPath , updatedData, 0644)
	if err != nil {
		return fmt.Errorf("Error trying to write metadata file:  %v", err)
	}
	return nil
}

func findLatestTagNumOrder(currLatest,tag string) string {
	if currLatest == "" {
		//currLatest is empty
		return tag
	} else if tag == "latest" {
		return tag
	} else {
		if tagFloat, err := strconv.ParseFloat(tag, 64); err == nil {
			if latestFloat, err := strconv.ParseFloat(currLatest, 64); err == nil {
				if tagFloat > latestFloat {
					return tag
				} else {
					return currLatest
				}
			} else {
				log.Printf("Error Trying to parse the tag %s , Skipping it while considering latest tag!",currLatest)
			}
		} else {
			log.Printf("Error Trying to parse the tag %s , Skipping it while considering latest tag!",tag)
		}
	}
	return currLatest
}
func ReadMetaData(destination string) ([]string, error){
	metaPath := filepath.Join(destination,"metadata.json")

	// now metadata.json exists , parse the json and append it and save it back
	data, err := ioutil.ReadFile(metaPath )
	if err != nil {
		return nil,err
	}
	var image model.Image
	err = json.Unmarshal(data, &image)
	if err != nil {
		return nil,err
	}
	return image.Image_Tag,nil
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

func ExtractLayer(destDir string, tag string,layerData io.Reader) error {
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
		target := filepath.Join(destDir, header.Name)
		log.Printf("Creating %s\n", target)
		if err := ExtractTarFile(target, header, tarReader); err != nil {
			return err
		}
	}
	return ChangeOwnershipImage(destDir)
}