package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	endpoint := "http://scalpel.service.consul/credit/batches/75dedb36-973f-432e-82b5-ccaa355021fd" // Replace with your actual endpoint

	for {
		err := makeHTTPRequest(endpoint)
		if err != nil {
			fmt.Printf("Error making HTTP request: %v\n", err)
		}

		time.Sleep(30 * time.Second)
	}
}

func makeHTTPRequest(endpoint string) error {
	resp, err := http.Get(endpoint)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP request failed with status code %d", resp.StatusCode)
	}

	// Read and print the body of the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %v", err)
	}

	fmt.Printf("Time: %s, HTTP request successful! Response Body:\n%s\n", time.Now(), body)
	return nil
}
