package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"slices"
)

const (
	jsonPath         = "./dev-result.json"
	lowerScenarioID  = 6500
	higherScenarioID = 7000
)

type PricingResult struct {
	ResultsMap Results `json:"results"`
}

/*
	type PricingResult struct {
		ResultsMap Results `json:"results"`
	}
*/

type Results struct {
	Main      MainResult                `json:"main"`
	Scenarios map[uint32]ScenarioResult `json:"scenarios"`
}

type MainResult struct {
	NPV struct {
		Value  float64 `json:"value"`
		Status string  `json:"status"`
	} `json:"NPV"`
	NotionalNPV struct {
		Value  float64 `json:"value"`
		Status string  `json:"status"`
	} `json:"notionalNPV"`
}

type ScenarioResult struct {
	Value  float64 `json:"value"`
	Status string  `json:"status"`
	id     uint32
}

func main() {
	content, err := os.ReadFile(jsonPath)
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}

	fmt.Println("File: ", jsonPath)
	fmt.Println("---------------------")

	var results PricingResult
	// var results Results

	err = json.Unmarshal(content, &results)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}

	scenarioValues := results.ResultsMap.Scenarios
	// scenarioValues := results.Scenarios
	maxScenarioValue := math.Inf(-1)
	maxScenarioID := uint32(0)

	maxScenarioValue2 := math.Inf(-1)
	maxScenarioID2 := uint32(0)

	for scenarioId, scenarioResult := range scenarioValues {
		if scenarioId < lowerScenarioID || scenarioId > higherScenarioID {
			continue
		}

		if scenarioResult.Value > maxScenarioValue {
			maxScenarioValue2 = maxScenarioValue
			maxScenarioID2 = maxScenarioID

			maxScenarioValue = scenarioResult.Value
			maxScenarioID = scenarioId

			continue
		}

		if scenarioResult.Value > maxScenarioValue2 {
			maxScenarioValue2 = scenarioResult.Value
			maxScenarioID2 = scenarioId
		}
	}

	fmt.Println("Max scenario ID: ", maxScenarioID)
	fmt.Println("Max scenario value: ", maxScenarioValue)

	fmt.Println("Second max scenario ID: ", maxScenarioID2)
	fmt.Println("Second max scenario value: ", maxScenarioValue2)

	fmt.Println("---------------------")

	minScenarioValue := math.Inf(1)
	minScenarioID := uint32(0)

	minScenarioValue2 := math.Inf(1)
	minScenarioID2 := uint32(0)

	for scenarioId, scenarioResult := range scenarioValues {
		if scenarioId < lowerScenarioID || scenarioId > higherScenarioID {
			continue
		}

		if scenarioResult.Value < minScenarioValue {
			minScenarioValue2 = minScenarioValue
			minScenarioID2 = maxScenarioID

			minScenarioValue = scenarioResult.Value
			minScenarioID = scenarioId

			continue
		}

		if scenarioResult.Value < minScenarioValue2 {
			minScenarioValue2 = scenarioResult.Value
			minScenarioID2 = scenarioId
		}
	}

	fmt.Println("Second min scenario ID: ", minScenarioID2)
	fmt.Println("Second min scenario value: ", minScenarioValue2)

	fmt.Println("Min scenario ID: ", minScenarioID)
	fmt.Println("Min scenario value: ", minScenarioValue)

	// -----------------------------
	scenarios := make([]ScenarioResult, 0, len(scenarioValues))
	for scenarioId, scenarioResult := range scenarioValues {
		if scenarioId < lowerScenarioID || scenarioId > higherScenarioID {
			continue
		}
		scenarioResult.id = scenarioId
		scenarios = append(scenarios, scenarioResult)
	}

	slices.SortFunc(scenarios, func(a ScenarioResult, b ScenarioResult) int {
		if a.Value < b.Value {
			return -1
		}

		if a.Value > b.Value {
			return 1
		}

		return 0
	})

	fmt.Println("10 highest scenarios: ", scenarios[len(scenarios)-10:])
	fmt.Println("10 lowest scenarios: ", scenarios[:10])
	fmt.Println("---------------------")
	// -----------------------------
	nbScenarios := 50
	lowScenarios := scenarios[:nbScenarios]

	var es float64
	for _, scenario := range lowScenarios {
		// fmt.Println("Low scenario: ", scenario.id, scenario.Value)
		es += (scenario.Value) / float64(nbScenarios)
	}

	v := lowScenarios[len(lowScenarios)-1].Value/results.ResultsMap.Main.NPV.Value - 1
	es = es/results.ResultsMap.Main.NPV.Value - 1

	// v := lowScenarios[len(lowScenarios)-1].Value/results.Main.NPV.Value - 1
	// es = es/results.Main.NPV.Value - 1
	vol := 0.0
	for _, scenario := range scenarios {
		vol += math.Pow(scenario.Value-results.ResultsMap.Main.NPV.Value, 2)
		// vol += math.Pow(scenario.Value-results.Main.NPV.Value, 2)
	}
	vol = math.Sqrt(vol / float64(len(scenarios)))

	fmt.Println("VaR: ", v)
	fmt.Println("ES: ", es)
	fmt.Println("Vol: ", vol)
}
