package main

import (
	"encoding/json"
	"os"
)

type Scenario struct {
	ID uint32 `json:"id"`
}

func main() {
	scenarioMap := make(map[uint32]Scenario, 500)
	for i := 6500; i < 7000; i++ {
		scenarioMap[uint32(i)] = Scenario{ID: uint32(i)}
	}

	file, _ := json.MarshalIndent(scenarioMap, "", " ")
	_ = os.WriteFile("scenarios_map.json", file, 0644)
}
