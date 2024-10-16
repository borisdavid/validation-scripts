package main

import (
	log "github.com/sirupsen/logrus"
)

const token = "eyJraWQiOiJzZXNhbWUtZWRnZWxhYi1hcGktMTY1NzE4NDA1MSIsInR5cCI6IkpXVCIsImFsZyI6IlJTMjU2In0.eyJqdGkiOiI0MTJjYTY2Ny0zOTU2LTRlODAtYTcyYS1hMDAxY2NmMTNlZTgiLCJzdWIiOiJhdXRoMHw2MTcxMGFkZTJjYWVlODAwNzFiOGNkNTQiLCJpc3MiOiJodHRwczovL2ludGVybmFsLWFwaS5lZGdlbGFiLmNoL3Nlc2FtZS8iLCJhdWQiOiJodHRwczovL2FwaS5lZGdlbGFiLmNoIiwiaWF0IjoxNzI3NjM1NTIyLCJleHAiOjE3Mjc2NzE1MjEsIm5iZiI6MTcyNzYzNTUyMiwiaHR0cHM6Ly9lZGdlbGFiLmNoL29yZ2FuaXphdGlvbiI6IjEyYWM5MjNiLTMxM2QtNGUwOS05Y2Y1LTBkNDNlMWY2MzBlYyJ9.n4vt6CistVnSDZtq8YPgeQhRo8awNrD12g0ZcjmmqwGblfqKmqPyMUfVkaKpEJPipeWmr2fHLdFxIghNGnr6a7C5R5zeS4gppltJukiOqI_tSYIxp8DuilodVe6yJ7hffYVjoLcafDjNcMbxVwAVC3j4Huz0VK_cN0RfOyIRQ2ljAccyJpmu75xhD4bQJkAIuJQiuYHMLT1itS200K3spS1Sz4uwJMdNctuorhU2G4NJMHTk0lhvc3pEuJ579LzGuF5LW1hTQ0cs-qB3YkBCRNGFau-qQgiXy3hYt3iifbv5BFWzV-mhX_GTGoZjTq8_4uUlv8ac1xtCR-aCJhJ_DQ"

func main() {
	// Fetch issuers.
	issuers, err := IssuerList()
	if err != nil {
		log.Fatalf("could not fetch issuers: %v", err)
	}

	// issuers := []string{"48097721-bfec-48f6-b10d-18d830d2e18a"}

	// Fetch issuers descriptions.
	descriptions, err := IssuerDescriptions(issuers)
	if err != nil {
		log.Fatalf("could not fetch issuers descriptions: %v", err)
	}

	// Write results in CSV.
	err = outputToCsv("./output/", descriptions)
	if err != nil {
		log.Fatalf("could not write results in CSV: %v", err)
	}
}
