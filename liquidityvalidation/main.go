package main

import (
	"encoding/csv"
	"os"

	log "github.com/sirupsen/logrus"
)

const (
	environment = "DEV"
	// environment = "PROD"

	tokenDEV  = "eyJraWQiOiJzZXNhbWUtZWRnZWxhYi1hcGktMTY1NjkzNDkzNSIsInR5cCI6IkpXVCIsImFsZyI6IlJTMjU2In0.eyJqdGkiOiJmNWFmZDBmNy02YTkwLTRmYWUtYTJjOC1hZDA3Y2FlZDYxZmYiLCJzdWIiOiJhdXRoMHw2MTcxMGFkZTJjYWVlODAwNzFiOGNkNTQiLCJpc3MiOiJodHRwczovL2ludGVybmFsLWFwaS5kZXYuZWRnZS1sYWIuY2gvc2VzYW1lLyIsImF1ZCI6Imh0dHBzOi8vYXBpLmRldi5lZGdlLWxhYi5jaCIsImlhdCI6MTcyNzc3NTIwMSwiZXhwIjoxNzI3ODExMjAwLCJuYmYiOjE3Mjc3NzUyMDEsImh0dHBzOi8vZWRnZWxhYi5jaC9vcmdhbml6YXRpb24iOiIxMmFjOTIzYi0zMTNkLTRlMDktOWNmNS0wZDQzZTFmNjMwZWMifQ.p5hIY4PumNkLx-Ki8nZl4JNweI-qJOMRX_rG1WiHw7818-Oc4wJ-MlKsPvKViXUa3Sc4ZSC6fe_WTmxz8UcqoOWeeSDHonKl-_OU_Tt0Zc85LChIpamgOEDqkjkAT44p_4FKbzbUdcey_kp3JGklWJOZo_vrlnvey5k_N3TyRyMXEuG7ZV4f1dYw8IA2MRQIqVDfKBIA4c7NHOnAmDacAGubGemjesZLpRv3HOc4ybkLhzs934Q0SzHb-VOu6NTU0q6yjHsv7mgyvJ_WlPdT9dj5Z26Sjt5ptHgDIP05s-I1_Ix8Rypf69yoa9rhKOFPdLne8gCSS1_qRMtdmSBl6A"
	tokenPROD = "eyJraWQiOiJzZXNhbWUtZWRnZWxhYi1hcGktMTY1NzE4NDA1MSIsInR5cCI6IkpXVCIsImFsZyI6IlJTMjU2In0.eyJqdGkiOiI1M2EzMGE4Ni05OTI0LTRkNzEtYTFhOC1kY2JmZWM4Y2Q0ZTQiLCJzdWIiOiJhdXRoMHw2MTcxMGFkZTJjYWVlODAwNzFiOGNkNTQiLCJpc3MiOiJodHRwczovL2ludGVybmFsLWFwaS5lZGdlbGFiLmNoL3Nlc2FtZS8iLCJhdWQiOiJodHRwczovL2FwaS5lZGdlbGFiLmNoIiwiaWF0IjoxNzI3Njg2MzM0LCJleHAiOjE3Mjc3MjIzMzQsIm5iZiI6MTcyNzY4NjMzNCwiaHR0cHM6Ly9lZGdlbGFiLmNoL29yZ2FuaXphdGlvbiI6IjEyYWM5MjNiLTMxM2QtNGUwOS05Y2Y1LTBkNDNlMWY2MzBlYyJ9.IJv4dc8Fz3ildiaMZCBf3rmASp3UNV_DLLiuJMdtqogRvzIHmT4-QNrP6T850nK9SUNbZWdPqnYrqjjq9vn0gj1YSZj_ku55OWT4YqgDFz9PfjMgJRd6XCGfJXSNB4k2_lYDJ8EljGxDO7p0YEcGa6LAwKv9XrJvF9K6gAc294tjklcSTvoXaa7K8jpyQeqdXGIr5N4o-NQmv-LE3wZEHHoGOmBP2KVWGdPpuk7fS1gOUqVEuGEOORtuYYchTTO8_Oqg_vdzXYuhrg2jOIyx6b1o0LmwrZu49zMFWmAq8U_noMv5WtrhEmrKn9etGrUYKtZO1TVK-AAZLexJDVTBCw"

	snapshotDEV  = "2024-09-30T00:30:04Z"
	snapshotPROD = "2024-09-29T19:30:04Z"
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
