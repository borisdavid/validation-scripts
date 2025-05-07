package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

const (
	//cerberusHost = "http://cerberus.service.consul"

	cerberusHost = "https://api.dev.edge-lab.ch/cerberus"
	token        = "eyJraWQiOiJzZXNhbWUtZWRnZWxhYi1hcGktMTY1NjkzNDkzNSIsInR5cCI6IkpXVCIsImFsZyI6IlJTMjU2In0.eyJqdGkiOiI2ZmFkYmE2My1lMTBlLTRhYTktYTgyZi04ODJiYzM5NTQ1MjgiLCJzdWIiOiJhdXRoMHw2MTcxMGFkZTJjYWVlODAwNzFiOGNkNTQiLCJpc3MiOiJodHRwczovL2ludGVybmFsLWFwaS5kZXYuZWRnZS1sYWIuY2gvc2VzYW1lLyIsImF1ZCI6Imh0dHBzOi8vYXBpLmRldi5lZGdlLWxhYi5jaCIsImlhdCI6MTc0NDE5Nzc5OSwiZXhwIjoxNzQ0MjMzNzk5LCJuYmYiOjE3NDQxOTc3OTksImh0dHBzOi8vZWRnZWxhYi5jaC9vcmdhbml6YXRpb24iOiI3NGIyYmUyNS05NDk3LTRhZjAtODAzYi1jYTZkNTg3ODUzMzEifQ.HnYSR-wbPcALtN-8sHOvyotOY9TXAShw_VPlK0ivSgYRIo6poinAqr2l2GRB70W2yC6Va7s4pQj7QbuRYEtTVltuWXYn-UT5zIc_BjuA20RC0Uy-KQWkXHY0p18lZbd_DCk3zMahA1OS5anG9hTS9gvMEIcymtMeRaB2X_54Rihyh70TjVdzmFoY_JQsZgdjwp4hse-7BP7NARDdyqbOT85ec7WKfZNsdB-xNfpKQzLbRKdxjoFT1Nj24qbmR3Zg4RCeMI8bSZmuAn6Qzob839Qu-fNMj_S3GNhmhVmPqXone6NLp4BMEmBMb1xhczzZ50k5q6MIaRKrCLiwyyndmQ"
)

type InputIsins struct {
	Input []string `json:"input"`
}

type output struct {
	ISIN string
	ID   string
}

func main() {
	// ==================================
	// Open the JSON file
	jsonFile, err := os.Open("input.json")
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer jsonFile.Close()

	// Read the file's content
	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	// Parse JSON into struct
	var inputIsins InputIsins
	if err := json.Unmarshal(byteValue, &inputIsins); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	// ===================================
	// Call cerberus API http.
	ids := make([]output, 0, len(inputIsins.Input))
	for _, isin := range inputIsins.Input {
		// Call the API with the ISIN
		// Example: fmt.Printf("Calling API with ISIN: %s\n", isin)
		// Here you would implement the actual API call logic

		id, err := callCerberus(isin)
		if err != nil {
			log.Printf("Error calling API for ISIN %s: %v", isin, err)
			ids = append(ids, output{ISIN: isin})

			continue
		}

		// Append the ID to the list
		ids = append(ids, output{ISIN: isin, ID: id})
	}

	// Create the CSV file
	file, err := os.Create("output.csv")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	// Create CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Optional: write header
	writer.Write([]string{"Key", "Value"})

	// Write map entries
	for _, value := range ids {
		err := writer.Write([]string{value.ISIN, value.ID})
		if err != nil {
			fmt.Println("Error writing to CSV:", err)

			return
		}
	}

	fmt.Println("CSV file written successfully.")
}

func callCerberus(isin string) (string, error) {
	// Implement the API call logic here
	// For example, using the net/http package to make a GET request
	// and return the response or an error
	url := fmt.Sprintf("%s/assets/isin/%s?view=full", cerberusHost, isin)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("could not create http request for url %s: %w", url, err)
	}
	req.Header.Set("x-internal-service", "validation")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("could not send the request: %w", err)
	}

	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("could not read response body: %w", err)
	}

	var output struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(raw, &output); err != nil {
		return "", fmt.Errorf("received non-JSON response: %s, error was: %w", string(raw), err)
	}

	return output.ID, nil
}
