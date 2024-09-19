package main

import (
	"encoding/csv"
	"os"

	log "github.com/sirupsen/logrus"
)

const (
	// environment = "DEV"
	environment = "PROD"

	tokenDEV  = "eyJraWQiOiJzZXNhbWUtZWRnZWxhYi1hcGktMTY1NjkzNDkzNSIsInR5cCI6IkpXVCIsImFsZyI6IlJTMjU2In0.eyJqdGkiOiI1NzRhZjkzZS02YzM3LTQyNjMtYjNkZi04NTU3MWMyYjdkYjgiLCJzdWIiOiJhdXRoMHw2MTcxMGFkZTJjYWVlODAwNzFiOGNkNTQiLCJpc3MiOiJodHRwczovL2ludGVybmFsLWFwaS5kZXYuZWRnZS1sYWIuY2gvc2VzYW1lLyIsImF1ZCI6Imh0dHBzOi8vYXBpLmRldi5lZGdlLWxhYi5jaCIsImlhdCI6MTcyNjczMDY0MiwiZXhwIjoxNzI2NzY2NjQyLCJuYmYiOjE3MjY3MzA2NDIsImh0dHBzOi8vZWRnZWxhYi5jaC9vcmdhbml6YXRpb24iOiIxMmFjOTIzYi0zMTNkLTRlMDktOWNmNS0wZDQzZTFmNjMwZWMifQ.rLV6111EFAoNXsB0eMhaLb7xCjKC5aRaVv4ueO-vuARNF-Ej7ZQsI9J8eRmEfcbSoU5qj4DRS9nhpCCMMDPc7eZ6nMSRdHdIRyqYACH7ulYdA_Y6__Sr41zuNu7mTVLaWZFgqUUyYmaVkWmfORPeJa9uJQp9GdwZHFNbIO9JvuZIjldCiKP7twkHPbAF6t6z4DJoptEnMvB0AAXahzB4sUKVR-ifOJtqpyIoQoNYx_ceGseRKRgn74hstp_7C9TQ8z9B3Mh8NEnvXSxGppjO3pCz5CtIxDqhCY3icoTMN2P8i0bRuCDJSQAjYH81HySw7R0mjn3p_mZtTJTApr_mew"
	tokenPROD = "eyJraWQiOiJzZXNhbWUtZWRnZWxhYi1hcGktMTY1NzE4NDA1MSIsInR5cCI6IkpXVCIsImFsZyI6IlJTMjU2In0.eyJqdGkiOiI4NTU5Zjc5Ni04NDdmLTQ1MTctOTZkZi1lMmU2NjM3OGE5ZGMiLCJzdWIiOiJhdXRoMHw2MTcxMGFkZTJjYWVlODAwNzFiOGNkNTQiLCJpc3MiOiJodHRwczovL2ludGVybmFsLWFwaS5lZGdlbGFiLmNoL3Nlc2FtZS8iLCJhdWQiOiJodHRwczovL2FwaS5lZGdlbGFiLmNoIiwiaWF0IjoxNzI2NzMwNjEzLCJleHAiOjE3MjY3NjY2MTMsIm5iZiI6MTcyNjczMDYxMywiaHR0cHM6Ly9lZGdlbGFiLmNoL29yZ2FuaXphdGlvbiI6IjEyYWM5MjNiLTMxM2QtNGUwOS05Y2Y1LTBkNDNlMWY2MzBlYyJ9.Mn3Exz3vx9pIkBpiav-OQvbc29GaXGzhP01Jayi19VvsX_6Rj36Jv6xJActcbWhcqPA7kSl2gKM_z3cHDEASHaKM-hbINY8-lRDuyynaE6meYaQCDdAnHJddtDjkZnvCfX4v_vztw4cJTkiWjX-2ojMuAd4nMr5pTNCvtw24nxuI7FGzHUt8haj4nmI62BH5PBXqQ1HRlTxTv9x79ebFGcS0Jzjm-HVDCDGh1C46N0QyPyYsPqys8jAPBYoy6flu4HlppSepd9kwFfyKpFDEm6SLpHLObKzs5QwCg3QG8nU-XCGJpZCGNPlA2qOS6FSuTDPf0DXRJUZOctR_10znQw"

	snapshotDEV  = "2024-09-19T00:30:04Z"
	snapshotPROD = "2024-09-18T19:30:05Z"
)

func main() {
	file, err := os.Open("input.csv")
	if err != nil {
		log.Fatal("Error while reading the file", err)
	}

	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal("Error while reading the file", err)
	}

	assetIDs := make([]string, 0)
	for _, record := range records[1:] {
		assetIDs = append(assetIDs, record[0])
	}

	log.Info("Starting the process by fetching liquidity in MD")
	outputMD, err := requestMarketdata(assetIDs)
	if err != nil {
		log.Fatal("Error while calling request", err)
	}

	// -----------------------------------------

	log.Info("fetch requests in Adam")
	outputAdam, err := requestAdam(assetIDs)
	if err != nil {
		log.Fatal("Error while calling Adam request", err)
	}

	log.Info("fetch responses in Eve")
	outputEve, err := requestEve(outputAdam)
	if err != nil {
		log.Fatal("Error while calling Eve request", err)
	}

	// -----------------------------------------

	log.Info("fetch requests in Arcanist")
	outputArcanist, outputArcanistMD, err := requestArcanist(outputMD)
	if err != nil {
		log.Fatal("Error while calling Arcanist request", err)
	}

	log.Info("fetch responses in Recco")
	outputRecco, err := requestRecco(assetIDs)
	if err != nil {
		log.Fatal("Error while calling Recco request", err)
	}

	// -----------------------------------------

	log.Info("Save results in csv")
	// Write output csv.
	err = outputToCsv(outputMD, outputEve, outputArcanist, outputArcanistMD, outputRecco)
	if err != nil {
		log.Fatalf("Error converting results to csv: %v", err)
	}
}
