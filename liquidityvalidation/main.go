package main

import (
	"encoding/csv"
	"os"

	log "github.com/sirupsen/logrus"
)

const (
	// environment = "DEV"
	environment = "PROD"

	tokenDEV  = "eyJraWQiOiJzZXNhbWUtZWRnZWxhYi1hcGktMTY1NjkzNDkzNSIsInR5cCI6IkpXVCIsImFsZyI6IlJTMjU2In0.eyJqdGkiOiJmNWFmZDBmNy02YTkwLTRmYWUtYTJjOC1hZDA3Y2FlZDYxZmYiLCJzdWIiOiJhdXRoMHw2MTcxMGFkZTJjYWVlODAwNzFiOGNkNTQiLCJpc3MiOiJodHRwczovL2ludGVybmFsLWFwaS5kZXYuZWRnZS1sYWIuY2gvc2VzYW1lLyIsImF1ZCI6Imh0dHBzOi8vYXBpLmRldi5lZGdlLWxhYi5jaCIsImlhdCI6MTcyNzc3NTIwMSwiZXhwIjoxNzI3ODExMjAwLCJuYmYiOjE3Mjc3NzUyMDEsImh0dHBzOi8vZWRnZWxhYi5jaC9vcmdhbml6YXRpb24iOiIxMmFjOTIzYi0zMTNkLTRlMDktOWNmNS0wZDQzZTFmNjMwZWMifQ.p5hIY4PumNkLx-Ki8nZl4JNweI-qJOMRX_rG1WiHw7818-Oc4wJ-MlKsPvKViXUa3Sc4ZSC6fe_WTmxz8UcqoOWeeSDHonKl-_OU_Tt0Zc85LChIpamgOEDqkjkAT44p_4FKbzbUdcey_kp3JGklWJOZo_vrlnvey5k_N3TyRyMXEuG7ZV4f1dYw8IA2MRQIqVDfKBIA4c7NHOnAmDacAGubGemjesZLpRv3HOc4ybkLhzs934Q0SzHb-VOu6NTU0q6yjHsv7mgyvJ_WlPdT9dj5Z26Sjt5ptHgDIP05s-I1_Ix8Rypf69yoa9rhKOFPdLne8gCSS1_qRMtdmSBl6A"
	tokenPROD = "eyJraWQiOiJzZXNhbWUtZWRnZWxhYi1hcGktMTY1NzE4NDA1MSIsInR5cCI6IkpXVCIsImFsZyI6IlJTMjU2In0.eyJqdGkiOiJjZTBiY2VkNi1mZTBiLTQ1ODEtOWNlNy03ZGQ1ODg4YjhkY2UiLCJzdWIiOiJhdXRoMHw2MTcxMGFkZTJjYWVlODAwNzFiOGNkNTQiLCJpc3MiOiJodHRwczovL2ludGVybmFsLWFwaS5lZGdlbGFiLmNoL3Nlc2FtZS8iLCJhdWQiOiJodHRwczovL2FwaS5lZGdlbGFiLmNoIiwiaWF0IjoxNzI4OTA5MjUzLCJleHAiOjE3Mjg5NDUyNTMsIm5iZiI6MTcyODkwOTI1MywiaHR0cHM6Ly9lZGdlbGFiLmNoL29yZ2FuaXphdGlvbiI6IjEyYWM5MjNiLTMxM2QtNGUwOS05Y2Y1LTBkNDNlMWY2MzBlYyJ9.UvdynNdySoYbSVzR6ceY2_WzAcGmyxdq5_eoW_aYn27U6ydRjhD31Sb_EgrR09uqCyPgXcAYLOnzc64ErBydCGhdf8OeRXK-70oKWR8wvxbxxS-4rkGnCiWVsSCuiqH6bXboHB2HF4N-AyqnGUzRjbkpH_nH4pMKpVY8nkUSPuh8Ht9ihHQPomiKi5sPx3wo0rm_x9Mqupnld2e63r1jh_skrkbGN1Nt4njTVLl6AFuNVl1iiiwfVWOLGtW0ElhlucXwxFto297077eSd6KVKmoyH5nxfHfPhPA7oaVczzsl27NX2TwLNPwdNO_93C64V0NLxgKd3x8ruxL460N0SA"

	snapshotDEV  = "2024-10-14T00:30:04Z"
	snapshotPROD = "2024-10-13T19:30:05Z"
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
