package main

import (
	log "github.com/sirupsen/logrus"
)

const token = "eyJraWQiOiJzZXNhbWUtZWRnZWxhYi1hcGktMTY1NzE4NDA1MSIsInR5cCI6IkpXVCIsImFsZyI6IlJTMjU2In0.eyJqdGkiOiIzNGI5YTBiMC0xZTk3LTRkOGMtYTg3NC02NzI5YzFmOWUxZTYiLCJzdWIiOiJhdXRoMHw2MTcxMGFkZTJjYWVlODAwNzFiOGNkNTQiLCJpc3MiOiJodHRwczovL2ludGVybmFsLWFwaS5lZGdlbGFiLmNoL3Nlc2FtZS8iLCJhdWQiOiJodHRwczovL2FwaS5lZGdlbGFiLmNoIiwiaWF0IjoxNzQ2NjA3NzQwLCJleHAiOjE3NDY2NDM3NDAsIm5iZiI6MTc0NjYwNzc0MCwiaHR0cHM6Ly9lZGdlbGFiLmNoL29yZ2FuaXphdGlvbiI6IjEyYWM5MjNiLTMxM2QtNGUwOS05Y2Y1LTBkNDNlMWY2MzBlYyJ9.fyqJAkwJ3pL4QQOtxlX8ZFQ6hVqCA5bFr5wpslhrNMilWWHR_xqj0bMX7_lezxjypyVbEQv2FBr9Eelo9rkfmmy2gI7t7E2dcEjU3dmUcnxsSQiSgxVpANbkA4q3JYIZVjF4ryvnyQ74xUvKRc9dEGkA3US9iuvXR4kvC0ygUgn8_QYmO-6ZAEKhV-Xg5cqn4ItsZ-tOY_29qZRVXG6WBme6A6o-avVGI1NY_7Z4b7q1ojsZHD_MeOEq1K-rnKzAwmNB4omgelFfnkKONX24-XJ5KlskaWf5_ELc_QZDODs36RmfgvFxjT4W3ka46JMwsa57Z4RcubcuJ7r9ihAt9w"

func main() {
	// Fetch issuers.
	issuers, err := IssuerList()
	if err != nil {
		log.Fatalf("could not fetch issuers: %v", err)
	}

	log.Infof("Fetched %d issuers", len(issuers))

	ratings := make(chan *IssuerCreditRating, len(issuers))

	go func() {
		for i, issuer := range issuers {
			if i%200 == 0 {
				log.Infof("Fetched %d issuers (%f%%)", i, float64(i)/float64(len(issuers))*100)
			}

			rating, err := FetchIssuerCreditRating(issuer)
			if err != nil {
				//log.Errorf("could not fetch issuer credit rating for %s: %v", issuer, err)

				continue
			}

			log.Infof("Fetched issuer credit rating for %s : short term %s, long term %s", issuer, rating.ShortTerm.Rating, rating.LongTerm.Rating)

			ratings <- rating
		}

		close(ratings)
	}()

	// Write results in CSV.
	err = outputToCsv("./output/", ratings)
	if err != nil {
		log.Fatalf("could not write results in CSV: %v", err)
	}
}
