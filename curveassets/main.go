package main

import (
	"encoding/csv"
	"log"
	"os"
)

const (
	// environment = "DEV"
	environment = "PROD"

	tokenDEV  = "eyJraWQiOiJzZXNhbWUtZWRnZWxhYi1hcGktMTY1NjkzNDkzNSIsInR5cCI6IkpXVCIsImFsZyI6IlJTMjU2In0.eyJqdGkiOiJmZDJkYjMwMS01NWZmLTQ5YzQtOGMzNS05YTkyMGZhNmU5NjAiLCJzdWIiOiJhdXRoMHw2MTcxMGFkZTJjYWVlODAwNzFiOGNkNTQiLCJpc3MiOiJodHRwczovL2ludGVybmFsLWFwaS5kZXYuZWRnZS1sYWIuY2gvc2VzYW1lLyIsImF1ZCI6Imh0dHBzOi8vYXBpLmRldi5lZGdlLWxhYi5jaCIsImlhdCI6MTcyNzg2OTcxOCwiZXhwIjoxNzI3OTA1NzE4LCJuYmYiOjE3Mjc4Njk3MTgsImh0dHBzOi8vZWRnZWxhYi5jaC9vcmdhbml6YXRpb24iOiIxMmFjOTIzYi0zMTNkLTRlMDktOWNmNS0wZDQzZTFmNjMwZWMifQ.c-EpkP3HBAuavmQmgzQ3dVKJ0uHlTQ6yTh724-8T1jY1I4Bw4a67ongzEN0aePJTi44YL03WRHxrsBPjRpj6M_3PXMy8GjjiE1qrvWjGdYAQi4pDdCRKcvlDNC57naXXALhRaLPbAOLN-lIX0fEZMRpo6XRdre7PAayQt1Fmyq93w_1mF5RfCSFnxzVjaQj7Ja8nzwBKf0HDCo_1hQZ0mb-2ASMvLfPAz6tp1_eYko5fta1R73vg1Q1eBaeHrmG6gRiZp26EcNww4Lf9fxVyeYcfaKxiqrAfHUSuMl9qG8DDWBfIlKcbLxGnJL7RCc1rC2omGHGu9XYa60EzJbeG4g"
	tokenPROD = "eyJraWQiOiJzZXNhbWUtZWRnZWxhYi1hcGktMTY1NzE4NDA1MSIsInR5cCI6IkpXVCIsImFsZyI6IlJTMjU2In0.eyJqdGkiOiIwNjZjZTQwZi0yOWRkLTQzNDktOTY4My02ODViMmNlNTUxZTYiLCJzdWIiOiJhdXRoMHw2MTcxMGFkZTJjYWVlODAwNzFiOGNkNTQiLCJpc3MiOiJodHRwczovL2ludGVybmFsLWFwaS5lZGdlbGFiLmNoL3Nlc2FtZS8iLCJhdWQiOiJodHRwczovL2FwaS5lZGdlbGFiLmNoIiwiaWF0IjoxNzI3ODY5Njc1LCJleHAiOjE3Mjc5MDU2NzQsIm5iZiI6MTcyNzg2OTY3NSwiaHR0cHM6Ly9lZGdlbGFiLmNoL29yZ2FuaXphdGlvbiI6IjEyYWM5MjNiLTMxM2QtNGUwOS05Y2Y1LTBkNDNlMWY2MzBlYyJ9.d0W8OL3ip-7VzeLsa_j-wqTw2mV0UdSQZeWZuHEGCzxAsOv15W9IYgu42LN7VKATDh7RmEP7FdTsRj3N9ZF9vHB9v7LhOObOOAONT-Zj8kX9drx0--m92NfQ33cdSgO3FDsS4Ztr1K7E822dzFJkw2DxhUPrFuijnGU9_HnYyRKv_l2ZGf1wf2X_nrT4Il6JLd3fvsDxdjvteZ6DIrAirRP6GX_6QsBd3rNlUevTNmIW9BgKpj_yCL_VhtgaOBeGxtqHbBQa7Dh4iIXLx8QjfM10D7-ZkSLhPMNza1AMmF1Rm7GLrnI2S2a5zjhumC3JIV_NRtQ9sx-KD3AzKGvPzQ"
)

func main() {
	// Read issuers.
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

	issuersIDs := make([]string, 0)
	for _, record := range records[1:] {
		issuersIDs = append(issuersIDs, record[0])
	}

	// Fetch curve assets.
	assetsCountChan, err := CurveAssetsCount(issuersIDs)
	if err != nil {
		log.Fatalf("could not fetch curve assets: %v", err)
	}

	// Write results in CSV.
	err = outputToCsv("./output/", assetsCountChan, len(issuersIDs))
	if err != nil {
		log.Fatalf("could not write results in CSV: %v", err)
	}
}
