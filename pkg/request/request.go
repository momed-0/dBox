package request

import (
	"net/http"
	"time"
	"fmt"
	"log"
	"io/ioutil"
	"encoding/json"
	"dBox/pkg/model"
	"dBox/pkg/utils"
	"path"
)

func HttpClient() *http.Client {
	client := &http.Client{Timeout: 10 * time.Second}
	return client
}

func SendRequest(client *http.Client, method string, endpoint string,token string) ([]byte, error) {
	/*
		Send API request and return the response body
		Args:
			client : http client
			method : request method. Currently supports only GET
			endpoint: endpoint url
	*/
	req, err := http.NewRequest(method, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s request: %v", method, err)
	}
	if token != "" {
		req.Header.Add("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request to the server: %v", err)
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading API response: %v", err)
	}
	return responseBody, nil
}

// createHTTPRequest creates an HTTP request with the necessary headers
func CreateManifestHTTPRequest(method, url, authToken string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}
	return req, nil
}

// executeHTTPRequest executes the HTTP request and returns the response
func ExecuteManifestHTTPRequest(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return resp, nil
}

// getAuthToken retrieves the authorization token for pulling images
func GetAuthToken(imageName string) (model.AuthResponse, error) {
	authEndpoint := fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:library/%s:pull", imageName)
	client := HttpClient()
	auth, err := SendRequest(client, http.MethodGet, authEndpoint, "")
	if err != nil {
		return model.AuthResponse{}, fmt.Errorf("failed to fetch auth token: %w", err)
	}

	var authData model.AuthResponse
	if err := json.Unmarshal(auth, &authData); err != nil {
		return model.AuthResponse{}, fmt.Errorf("error parsing auth token: %w", err)
	}
	return authData, nil
}

// fetchManifestList retrieves the manifest list for a specific image and tag
func FetchManifestList(imageName, tag, authToken string) (model.ManifestList, error) {
	manifestURL := fmt.Sprintf("https://registry-1.docker.io/v2/library/%s/manifests/%s", imageName, tag)
	
	req, err := CreateManifestHTTPRequest("GET", manifestURL, authToken)
	if err != nil {
		return model.ManifestList{}, err
	}

	resp, err := ExecuteManifestHTTPRequest(req)
	if err != nil {
		return model.ManifestList{}, err
	}
	defer resp.Body.Close()

	var manifestList model.ManifestList
	if err := utils.ParseJSONResponse(resp, &manifestList); err != nil {
		return model.ManifestList{}, err
	}

	return manifestList, nil
}

// fetchLayer handles the download and extraction of a single layer
func FetchLayer(endpoint, imageName, digest, authToken string) error {
	digestURL := fmt.Sprintf("%s%s", endpoint, digest)
	log.Println("Fetching layer from:", digestURL)
	req, err := CreateManifestHTTPRequest("GET", digestURL, authToken)
	if err != nil {
		return err
	}

	resp, err := ExecuteManifestHTTPRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	destination := path.Join(model.IMAGE_DIR, imageName)
	if err := utils.ExtractLayer(destination, resp.Body); err != nil {
		return fmt.Errorf("failed to extract layer: %v", err)
	}
	return nil
}

func FetchEachLayer(layers []model.Config,imageName string,authToken string) error {

	endpoint := fmt.Sprintf("https://registry-1.docker.io/v2/library/%s/blobs/", imageName)

	for _, layer := range layers {
		if err := FetchLayer(endpoint, imageName, layer.Digest, authToken); err != nil {
			return err
		}
	}
	log.Println("All layers extracted successfully to:", path.Join(model.IMAGE_DIR, imageName))
	return nil
}


// fetchManifest retrieves the image manifest from the registry
func FetchManifest(imageName string,digestsha string,authToken string) error{
	manifestURL := fmt.Sprintf("https://registry-1.docker.io/v2/library/%s/manifests/%s", imageName, digestsha)

	req, err := CreateManifestHTTPRequest("GET", manifestURL, authToken)
	if err != nil {
		return err
	}

	resp, err := ExecuteManifestHTTPRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var manifest model.Manifest
	if err :=  utils.ParseJSONResponse(resp, &manifest); err != nil {
		return err
	}

	return FetchEachLayer(manifest.Layers, imageName, authToken)
}