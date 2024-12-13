package request

import (
	"net/http"
	"time"
	"fmt"
	"io/ioutil"
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