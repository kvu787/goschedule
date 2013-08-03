package fetch

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func Get(client *http.Client, url string) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("Error: %+v\n", err)
		log.Printf("Error in fetching: %v\n", url)
		return nil, errors.New("Error in request")
	}
	if resp.StatusCode > 399 || resp.StatusCode < 200 {
		return nil, fmt.Errorf("Error in fetching: %v\nFetch returned with a non-2XX/3XX status code\n", url)
	}
	// log.Printf("Fetched with status %d\nLink: %v", resp.Status, url)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("Error in reading fetched response body")
	}
	return body, nil
}
