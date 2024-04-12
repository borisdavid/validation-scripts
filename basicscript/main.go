package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
)

const (
	jsonPath = "./20240110-result.json"
)

type PricingResult struct {
	ResultsMap Results `json:"results"`
}

type Results struct {
	Main      MainResult                `json:"main"`
	Scenarios map[uint32]ScenarioResult `json:"scenarios"`
}

type MainResult struct {
	NPV struct {
		Value  float64 `json:"value"`
		Status string  `json:"status"`
	} `json:"NPV"`
}

type ScenarioResult struct {
	Value  float64 `json:"value"`
	Status string  `json:"status"`
}

func main() {
	content, err := os.ReadFile(jsonPath)
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}

	fmt.Println("File: ", jsonPath)
	fmt.Println("---------------------")

	var results PricingResult

	err = json.Unmarshal(content, &results)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}

	scenarioIDList := make([]int, 0, len(results.ResultsMap.Scenarios))
	for id := range results.ResultsMap.Scenarios {
		scenarioIDList = append(scenarioIDList, int(id))
	}

	sort.Ints(scenarioIDList)

	file, _ := json.MarshalIndent(scenarioIDList, "", " ")
	// _ = os.WriteFile(fmt.Sprintf("aab_pricing_scalpel_%f.json", creditRate.Value(1.0)), file, 0644)
	_ = os.WriteFile("scenarios.json", file, 0644)
}
