package main

import (
	"encoding/csv"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

const (
	// environment = "DEV"
	environment = "PROD"

	tokenDEV  = "eyJraWQiOiJzZXNhbWUtZWRnZWxhYi1hcGktMTY1NjkzNDkzNSIsInR5cCI6IkpXVCIsImFsZyI6IlJTMjU2In0.eyJqdGkiOiIyMTE0MzdiMS0zY2ZkLTRmOGMtOGU4NC1kOThkMGYxYzgwYjkiLCJzdWIiOiJhdXRoMHw2MTcxMGFkZTJjYWVlODAwNzFiOGNkNTQiLCJpc3MiOiJodHRwczovL2ludGVybmFsLWFwaS5kZXYuZWRnZS1sYWIuY2gvc2VzYW1lLyIsImF1ZCI6Imh0dHBzOi8vYXBpLmRldi5lZGdlLWxhYi5jaCIsImlhdCI6MTcyNjgxOTU1NywiZXhwIjoxNzI2ODU1NTU3LCJuYmYiOjE3MjY4MTk1NTcsImh0dHBzOi8vZWRnZWxhYi5jaC9vcmdhbml6YXRpb24iOiIxMmFjOTIzYi0zMTNkLTRlMDktOWNmNS0wZDQzZTFmNjMwZWMifQ.WfY9EmRe6hDp7QDvsijHekAKWlnqsb8WIjRgSFJWutL4OqV9faUdLodpy9dDSSjS4KfiiuugA-EQ_IAO8yte8ctRKPTNuzOjrENQHNVTyGULi2x6_4JbqNjQp9N1Nt2j7BCcf7QSbooji2qEG6ZXwr3Adpf0mjE2EM7JZFur6nMnPG7PqsKKQzKVUVcNm0FnlfhiWKKGx_QwOSMUFyMQ5ef-38RqK7nFB_djC-eufvAqINgY_GDMvUrdFy928zm0LorKXqF2E_haGPXbnksXh9o9izk6BnXoHC5io2WH2DYeclbaxMH3Rlq-wrFE5Fpxe2alifci3dQrCOV1c0ZfJg"
	tokenPROD = "eyJraWQiOiJzZXNhbWUtZWRnZWxhYi1hcGktMTY1NzE4NDA1MSIsInR5cCI6IkpXVCIsImFsZyI6IlJTMjU2In0.eyJqdGkiOiI5NzZjYTQ0My1lNzNlLTRiMjktOGJiMi1mZDE2MTM3OWYwMWMiLCJzdWIiOiJhdXRoMHw2MTcxMGFkZTJjYWVlODAwNzFiOGNkNTQiLCJpc3MiOiJodHRwczovL2ludGVybmFsLWFwaS5lZGdlbGFiLmNoL3Nlc2FtZS8iLCJhdWQiOiJodHRwczovL2FwaS5lZGdlbGFiLmNoIiwiaWF0IjoxNzI2ODE5NTg2LCJleHAiOjE3MjY4NTU1ODYsIm5iZiI6MTcyNjgxOTU4NiwiaHR0cHM6Ly9lZGdlbGFiLmNoL29yZ2FuaXphdGlvbiI6IjEyYWM5MjNiLTMxM2QtNGUwOS05Y2Y1LTBkNDNlMWY2MzBlYyJ9.KOn5wZHF7gydITcSJ7FkKR8QDm7aPFnGJfumH4_k8w1MleWzLubqjCkiUFdaayy-8x1IUrkNr1FRyuiEbY2-Ihoqye5Cb-w5NmDMOliA8Y-23gG5LGKtYJ9eh80fslPWjG8SsddgeiU075v5n8bvx-cB1V3LfKBTsIpC697FNp1tY0Ynk4JSGEHtwEWuS7knWOpqFJic-E3CifPpy4FmhmcZEUPC3uJ0aKF1nRjCOcin3-pzBTYPiEgpJl28xGx_Qpwt6xZCZnTTEOww2GeWotlKDkwNSk3IVtYYMF55NJUrFMp5HlBzrKKiDboZQncc4HBm9cO5evFz7h15_KIuMw"
)

func main() {
	// Load the issuer list.
	issuerIDs, err := readInputIssuers("input.csv")
	if err != nil {
		log.Fatal("Error while reading the file", err)
	}

	log.Infof("issuers : %v", issuerIDs)

	// Calibrate termstructures.
	calibratedCurves, err := calibrateCreditCurves(issuerIDs)
	if err != nil {
		log.Fatal("Error while calibrating term structures", err)
	}
	fmt.Println(calibratedCurves)

	// Load credit curves from Scalpel directly.
	creditCurves, err := requestCreditCurves(issuerIDs)
	if err != nil {
		log.Fatal("Error while fetching credit curves", err)
	}

	// Save in CSV format.
	err = outputToCsv("./output/", creditCurves)
	if err != nil {
		log.Fatal("Error while saving the output", err)
	}
}

func readInputIssuers(inputPath string) ([]string, error) {
	// Load the input file.
	// CSV format containing a list of issuers.
	file, err := os.Open(inputPath)
	if err != nil {
		log.Fatal("Error while reading the file", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal("Error while reading the file", err)
	}

	issuerIDs := make([]string, 0)
	for _, record := range records {
		issuerIDs = append(issuerIDs, record[0])
	}

	return issuerIDs, nil
}
