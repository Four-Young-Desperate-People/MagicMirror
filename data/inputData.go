package data

import "encoding/json"

type GenericData struct {
	Method string      `json:"method"`
	Data   interface{} `json:"data"`
}

type LED struct {
	R int `json:"r"`
	G int `json:"g"`
	B int `json:"b"`
	W int `json:"w"`
}

type AlarmStartJSON struct {
	LED          LED    `json:"LED"`
	ExerciseToDo string `json:"ExerciseToDo"`
	HRThreshold  int    `json:"HRThreshold"`
}

type ModuleDisplayJSON struct {
	Visible  string `json:"visible"`
	Position string `json:"position"`
}

var mapOfData = map[string]interface{}{
	"alarm_start":      AlarmStartJSON{},
	"display_exercise": "",
	"quiet_alarm":      false,
	"ping":             0,
}

var mapOfFuncs = map[string]func(interface{}){
	"alarm_start": AlamrStart,
}

func GetData(g GenericData) (interface{}, error) {
	dataType := mapOfData[g.Method]
	marsh, err := json.Marshal(g.Data)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(marsh, &dataType)
	if err != nil {
		return nil, err
	}
	return dataType, nil
}
