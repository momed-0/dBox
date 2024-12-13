package image

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"encoding/json"
	"strings"
	"dBox/pkg/filesystem"
	"dBox/pkg/request"
	"path"
	"archive/tar"
	"io"
	"os"
	"path/filepath"
	"compress/gzip"
	"os/exec"
)

const ARCH = "amd64"
const OS = "linux"

type AuthResponse struct {
	Token string `json:"token"`
}

type Platform struct {
	Architecture string   `json:"architecture"`
	OS           string   `json:"os"`
}

type ManifestFatList struct {
	Digest    string   `json:"digest"`
	Platform  Platform `json:"platform"`
}

type ManifestList struct {
	Manifests     []ManifestFatList `json:"manifests"`
}

type Config struct {
	MediaType string `json:"mediaType"`
	Size      int    `json:"size"`
	Digest    string `json:"digest"`
}

type Manifest struct {
	Config Config `json:"config"`
	Layers []Config `json:"layers"`
}

func fetchManifest(imageName string,digestsha string,authToken string) error{
	manifestURL := fmt.Sprintf("https://registry-1.docker.io/v2/library/%s/manifests/%s", imageName, digestsha)

	// Create a new HTTP request
	req, err := http.NewRequest("GET", manifestURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	// Add necessary headers
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}
		// Perform the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	// Check for successful status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp)
	}

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	var manifest Manifest
	err = json.Unmarshal(body, &manifest)
	if err != nil {
		log.Fatalf("Error unmarshalling manifest JSON: %v", err)
	}
	//fmt.Println(manifest)
	return fetchEachLayer(manifest.Layers,imageName, authToken)
}

const IMAGE_DIR = "./images/"

func fetchEachLayer(layers []Config,imageName string,authToken string) error {

	endpoint := fmt.Sprintf("https://registry-1.docker.io/v2/library/%s/blobs/", imageName)

	for _,layer := range layers {
		digest := layer.Digest
		digestURL := fmt.Sprintf(endpoint+digest)
		fmt.Println("Fetching layer from:", digestURL)

		// Make the HTTP request
		req, err := http.NewRequest("GET", digestURL, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+authToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to fetch layer: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to fetch layer, status: %s", resp.Status)
		}
		destination := path.Join(IMAGE_DIR, imageName)
		// Decompress and extract the layer
		err = extractLayer(destination, resp.Body)
		if err != nil {
			return fmt.Errorf("failed to extract layer: %v", err)
		}
	}
	fmt.Println("All layers extracted successfully to:", path.Join(IMAGE_DIR, imageName))
	return nil
}

func extractLayer(destDir string, layerData io.Reader) error {
	// Handle gzip decompression
	gz, err := gzip.NewReader(layerData)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %v", err)
	}
	defer gz.Close()

	// Untar the contents
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
		// Extract files based on header type
		target := filepath.Join(destDir, header.Name)
		fmt.Printf("Creating %s\n", target)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %v", err)
			}
		case tar.TypeReg:
			file, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file: %v", err)
			}
			if _, err := io.Copy(file, tarReader); err != nil {
				file.Close()
				return fmt.Errorf("failed to write file: %v", err)
			}
			file.Close()
		case tar.TypeSymlink:
			// Handle symbolic links explicitly
			targetLink := filepath.Join(destDir, header.Name)
			// Check if the target already exists
			if _, err := os.Lstat(targetLink); err == nil {
				// Target exists, skip creating symlink or handle conflict
				fmt.Println("Symlink already exists, skipping:", targetLink)
				continue // Skip this case, or handle it in another way
			}
			
			if err := os.Symlink(header.Linkname, targetLink); err != nil {
				return fmt.Errorf("failed to create symlink: %v", err)
			}
		}
		
	}
	return changeOwnershipImage(destDir)
}

func changeOwnershipImage(imageDir string) error {
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

func InitPull(imageName string,tag string) {

	if filesystem.FindImagesSaved(imageName,tag) == true {
		fmt.Printf("Image: %s with tag %s already exists\n",imageName,tag)
		return
	}
	authEndpoint := fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:library/%s:pull",imageName)
	

	client := request.HttpClient()
	auth, err := request.SendRequest(client, http.MethodGet, authEndpoint, "")
	if err != nil {
		log.Fatalf("Failed to fetch auth token: %v", err)
	}

	var authData AuthResponse
	if err := json.Unmarshal(auth, &authData); err != nil {
		log.Fatalf("Error parsing auth token: %v", err)
	}

	//fetch the manifest
	manifestURL := fmt.Sprintf("https://registry-1.docker.io/v2/library/%s/manifests/%s", imageName, tag)

	req, err := http.NewRequest(http.MethodGet, manifestURL, nil)
	if err != nil {
		log.Fatalf("Failed to create GET request for manifest: %v", err)
	}
	req.Header.Add("Authorization", "Bearer "+authData.Token)
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.list.v2+json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to fetch manifest: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		manifestBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Error reading manifest response: %v", err)
		}
		var manifestList ManifestList
		err = json.Unmarshal(manifestBody, &manifestList)
		if err != nil {
			log.Fatalf("Error unmarshalling manifest JSON: %v", err)
		}

		// var digests []string
		// Print out the layer digests
		for _, manifest := range manifestList.Manifests {
			//fmt.Printf("Layer Digest: %s\n", manifest.Digest)
			// fmt.Println(manifest.Platform)
			if manifest.Platform.Architecture == ARCH && manifest.Platform.OS == OS{
				err := fetchManifest(imageName, manifest.Digest,authData.Token)
				if err != nil {
					fmt.Println("Error:", err)
				}
				
			}
		}
		// for _,uniqueBlob := range digests {
		// 	layerURL := fmt.Sprintf("https://registry-1.docker.io/v2/library/%s/blobs/%s",imageName,uniqueBlob)
		// 	fmt.Println(layerURL)
		// }
	} else {
		log.Fatalf("Failed to fetch manifest: status code %d", resp.StatusCode)
	}
}

func extractHashValue(digest string) string {
	// Split the string by the colon (":") separator
	parts := strings.Split(digest, ":")
	if len(parts) > 1 {
		return parts[1] // Return the part after 'sha256:'
	}
	return ""
}


