package opencvTranslator

import (
	"MagicMirrorGo/data"
	"fmt"
	"time"
)

type Heartrate_Comms_Mock struct {
	Heartrate_Comms
}

func (h *Heartrate_Comms_Mock) Start() {
	fmt.Println("Starting MOCK")
	<-time.After(15 * time.Second)
	gd := data.GenericData{
		Method: "quiet_alarm",
		Data:   true,
	}

	fmt.Println("Should be stopping now")
	h.AndroidChan <- gd

	<-time.After(10 * time.Second)

	startDisplayingHeartrate := data.GenericData{
		Method: "start_displaying_heartrate",
		Data:   "--",
	}

	h.FrontendChan <- startDisplayingHeartrate

	<-time.After(10 * time.Second)

	heartrates := []int{
		60, 63, 72, 65, 75, 80, 77, 85, 93, 105,
	}
	for _, i := range heartrates {
		gd := data.GenericData{
			Method: "change_heartrate",
			Data:   i,
		}
		h.FrontendChan <- gd
		fmt.Println(i)
		<-time.After(500 * time.Millisecond)
	}

	h.TimeToDisplayHeartbeat = true
	stopAlarm := data.GenericData{
		Method: "stop_alarm",
		Data:   true,
	}
	fmt.Println("Should be done now")
	h.FrontendChan <- stopAlarm
	h.AndroidChan <- stopAlarm
	h.LEDChan <- data.LED{
		R: 0,
		G: 0,
		B: 0,
		W: 0,
	}
}
