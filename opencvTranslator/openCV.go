package opencvTranslator

import (
	"MagicMirrorGo/data"
	"bufio"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

type Heartrate_Comms struct {
	FrontendChan           chan data.GenericData
	Done                   chan struct{}
	AndroidChan            chan data.GenericData
	LEDChan                chan data.LED
	HRThreshold            int
	TimeToDisplayHeartbeat bool

	// Used to redisplay the exercise when the heartrate isn't high enough
	Exercise string
}

type FromOpenCV struct {
	Face bool     `json:"face"`
	Bpm  *float64 `json:"bpm,omitempty"`
}

func (c *Heartrate_Comms) Stop() {
	c.Done <- struct{}{}
}

func (c *Heartrate_Comms) Start() {
	// test command
	done := make(chan struct{}, 1)
	cmd := exec.Command("./Heartbeat", "-max", "10", "-min", "5", "-facedet", "deep", "-rppg", "g", "-f", "1", "-r", "1", "-gui", "false")
	//cmd := exec.Command("./Heartbeat", "-max", "10", "-min", "5", "-facedet", "deep", "-rppg", "g", "-f", "1", "-r", "1")
	//cmd := exec.Command("./Heartbeat -max 10 -min 5 -facedet deep -rppg g -f 1 -r 1 -gui false")

	//if err != nil {
	//	panic(err)
	//}
	//defer file.Close()
	//errFile, err := os.Create("opencverrOutput.log")
	//if err != nil {
	//	panic(err)
	//}
	//defer errFile.Close()
	//cmd.Stdout = file
	//cmd.Stderr = errFile
	r, _ := cmd.StdoutPipe()

	scanner := bufio.NewScanner(r)
	go func() {
		sendBpm := false
		seenThem := false
		startDisplayingDone := make(chan struct{}, 1)
		go func() {
			for {
				select {
				case <-time.After(time.Second * 30):
					gd := data.GenericData{
						"start_displaying_heartrate",
						60,
					}
					c.FrontendChan <- gd
					sendBpm = true
					select {
					case <-time.After(time.Second * 20):
						gd = data.GenericData{
							Method: "heartrate_not_high_enough",
							Data:   false,
						}
						c.FrontendChan <- gd
						sendBpm = false
						<-time.After(time.Second * 5)
						gd = data.GenericData{
							Method: "display_exercise",
							Data:   c.Exercise,
						}

					case <-startDisplayingDone:
						return
					}

				case <-startDisplayingDone:
					return
				}
			}

		}()

		for scanner.Scan() {
			line := scanner.Bytes()
			from := FromOpenCV{}
			err := json.Unmarshal(line, &from)
			fmt.Printf("Got something from openCV %s\n", string(line))
			if err != nil {
				fmt.Printf("Could not deserizlze %s \n", string(line))
				continue
			}
			if !seenThem && from.Face {
				seenThem = true
				gd := data.GenericData{
					Method: "quiet_alarm",
					Data:   true,
				}
				c.AndroidChan <- gd
			}

			if sendBpm && from.Bpm != nil {
				gd := data.GenericData{
					Method: "change_heartrate",
					Data:   int(*from.Bpm),
				}
				c.FrontendChan <- gd
				if int(*from.Bpm) > c.HRThreshold {
					fmt.Printf("We are at threshold! All done!! Heartrate %.2f, threshold %d\n", *from.Bpm, c.HRThreshold)
					startDisplayingDone <- struct{}{}
					stopAlarm := data.GenericData{
						Method: "stop_alarm",
						Data:   true,
					}
					c.FrontendChan <- stopAlarm
					c.AndroidChan <- stopAlarm
					c.LEDChan <- data.LED{
						R: 0,
						G: 0,
						B: 0,
						W: 0,
					}
					done <- struct{}{}
					err := cmd.Process.Kill()
					if err != nil {
						panic(err)
					}
				}
			}

			if from.Bpm != nil {
				fmt.Printf("Got face:%t, BPM: %.2f\n", from.Face, *from.Bpm)
			} else {
				fmt.Printf("Got face:%t\n", from.Face)
			}

		}
	}()

	err := cmd.Run()

	fmt.Printf("We were killed?\n")
	if err != nil {
		_, ok := err.(*exec.ExitError)
		fmt.Printf("The actual error: %v", err)
		if !ok {
			panic(err)
		}
	}
	<-done
	fmt.Println("Horray!!!")
}
