// Package fetch provides functions to fetch web pages.
package fetch

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// Get requests a url with the given client and returns the bytes
// of the response body if successful.
// A response with a non-2XX/3XX status code is considered an error.
func Get(client *http.Client, url string) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Fetch.Get error while using client.Get: %#v\nLink: %v", err, url)
	}
	defer resp.Body.Close()
	if resp.StatusCode > 399 || resp.StatusCode < 200 {
		return nil, fmt.Errorf("Fetch.Get returned with a non-2XX/3XX status code: %d\nLink: %v", resp.StatusCode, url)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Fetch.Get error in reading response body: %v", err)
	}
	return body, nil
}
