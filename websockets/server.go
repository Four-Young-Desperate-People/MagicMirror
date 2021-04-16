package websockets

import (
	"MagicMirrorGo/data"
	"MagicMirrorGo/opencvTranslator"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Server struct {
	// These channels represent who the receivers are
	frontendChannel chan data.GenericData
	androidChannel  chan data.GenericData
	ledChannel      chan data.LED
	hc              *opencvTranslator.Heartrate_Comms_Mock
}

func NewServer() *Server {
	// Make them buffered channels to remove deadlock
	ser := Server{
		frontendChannel: make(chan data.GenericData, 1000),
		androidChannel:  make(chan data.GenericData, 1000),
		ledChannel:      make(chan data.LED, 10),
	}
	return &ser
}

func ping(returnInt int, conn *websocket.Conn) error {
	gd := data.GenericData{
		Method: "pong",
		Data:   returnInt,
	}
	err := conn.WriteJSON(gd)
	return err
}

func (s *Server) FrontendWebsocket(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Got a connected Frontend websocket")
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}
	done := make(chan struct{})

	go func() {
		for {
			g := data.GenericData{}
			err := ws.ReadJSON(&g)
			if err != nil {
				done <- struct{}{}
				return
			}
		}
	}()

	// Before we do anything else, we need to give the frontend the defaults
	file, err := ioutil.ReadFile("defaultDisplays.json")
	if err != nil {
		panic(err)
	}
	displays := map[string]data.ModuleDisplayJSON{}
	err = json.Unmarshal(file, &displays)
	if err != nil {
		panic(err)
	}
	gd := data.GenericData{
		Method: "update_modules_display",
		Data:   displays,
	}

	// Don't block the front end goroutine
	go func() {
		fmt.Println("Wait a bit before we send module defaults to frontend")
		<-time.After(10 * time.Second)
		fmt.Println("Sent to frontend")
		err = ws.WriteJSON(gd)
		if err != nil {
			panic(err)
		}
	}()

	for {
		select {
		case gd := <-s.frontendChannel:
			err := ws.WriteJSON(gd)
			if err != nil {
				panic(err)
			}

		case <-done:
			fmt.Println("Frontend websocket disconnected.")
			return
		}
	}
}

func (s *Server) AndroidWebSocket(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}

	fmt.Println("Got a connected Android websocket")

	websocketChan := make(chan data.GenericData)
	done := make(chan struct{})

	go func() {
		for {
			g := data.GenericData{}
			err := ws.ReadJSON(&g)
			if err != nil {
				close(websocketChan)
				done <- struct{}{}
				return
			}
			websocketChan <- g
		}
	}()

	for {
		select {
		case gd := <-s.androidChannel:
			err := ws.WriteJSON(gd)
			if err != nil {
				panic(err)
			}
		case g := <-websocketChan:
			fmt.Printf("Got some data %s %v\n", g.Method, g.Data)
			if g.Method == "ping" {
				// fmt.Println("Got a Ping request")
				// It may seem dumb that I'm marshaling everything here, and it is. But go gets confused with ints
				// and floats, so I need to manually specify that "return int" is an int here
				returnInt := 0
				marsh, err := json.Marshal(g.Data)
				if err != nil {
					panic(err)
				}
				err = json.Unmarshal(marsh, &returnInt)
				if err != nil {
					panic(err)
				}

				err = ping(returnInt, ws)
				if err != nil {
					panic(err)
				}
			} else if g.Method == "alarm_start" {
				fmt.Println("Got a alarm_start")
				as := data.AlarmStartJSON{}
				marsh, err := json.Marshal(g.Data)
				if err != nil {
					panic(err)
				}
				err = json.Unmarshal(marsh, &as)
				if err != nil {
					panic(err)
				}

				gd := data.GenericData{
					Method: "display_exercise",
					Data:   as.ExerciseToDo,
				}

				s.ledChannel <- as.LED
				s.frontendChannel <- gd
				h := opencvTranslator.Heartrate_Comms{
					FrontendChan:           s.frontendChannel,
					AndroidChan:            s.androidChannel,
					Done:                   make(chan struct{}, 1),
					LEDChan:                s.ledChannel,
					HRThreshold:            as.HRThreshold,
					TimeToDisplayHeartbeat: false,
					Exercise:               as.ExerciseToDo,
				}
				s.hc = &opencvTranslator.Heartrate_Comms_Mock{
					h,
				}

				go s.hc.Start()
			} else if g.Method == "update_modules" {
				fmt.Println("Got a Update Module Display")
				displays := map[string]data.ModuleDisplayJSON{}
				marsh, err := json.Marshal(g.Data)
				if err != nil {
					panic(err)
				}
				err = json.Unmarshal(marsh, &displays)
				if err != nil {
					panic(err)
				}

				file, _ := json.Marshal(displays)
				err = ioutil.WriteFile("defaultDisplays.json", file, 0644)
				if err != nil {
					panic(err)
				}

				fmt.Println("Saved module display into file")

				gd := data.GenericData{
					Method: "update_modules_display",
					Data:   displays,
				}

				s.frontendChannel <- gd
				fmt.Println("Saved modules saved")

			} else if g.Method == "get_modules_display" {
				file, err := ioutil.ReadFile("defaultDisplays.json")
				if err != nil {
					panic(err)
				}
				displays := map[string]data.ModuleDisplayJSON{}
				err = json.Unmarshal(file, &displays)
				if err != nil {
					panic(err)
				}
				gd := data.GenericData{
					Method: "get_modules_display_result",
					Data:   displays,
				}
				fmt.Printf("Sending %v\n", gd)
				err = ws.WriteJSON(gd)
				if err != nil {
					panic(err)
				}
			}

		case <-done:
			fmt.Println("Android web socket closed. We are done here")
			return
		}
	}
}

func (s *Server) LEDWebsocket(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}

	fmt.Println("Got a connected LED websocket")

	// No websocket channel here as we are not expecting any data to come through the ALARM start
	done := make(chan struct{})

	go func() {
		for {
			g := data.GenericData{}
			err := ws.ReadJSON(&g)
			if err != nil {
				done <- struct{}{}
				return
			}
		}
	}()

	for {
		select {
		case led := <-s.ledChannel:
			err = ws.WriteJSON(led.W)
			if err != nil {
				panic(err)
			}
		case <-done:
			fmt.Println("LED websocket disconnected")
			return
		}
	}
}

func Run() {
	server := NewServer()
	http.HandleFunc("/frontend", server.FrontendWebsocket)
	http.HandleFunc("/android", server.AndroidWebSocket)
	http.HandleFunc("/led", server.LEDWebsocket)
	err := http.ListenAndServe(":3683", nil)
	if err != nil {
		panic(err)
	}
}

// A test client. Must be called within its own go routine
func Client() {
	time.Sleep(2 * time.Second)
	fmt.Println("Client will now attempt to connect")
	c, _, err := websocket.DefaultDialer.Dial("ws://localhost:3683/android", nil)
	if err != nil {
		panic(err)
	}
	defer c.Close()
	fmt.Println("Client connected")
	as := data.AlarmStartJSON{
		LED: data.LED{
			R: 0,
			G: 0,
			B: 0,
			W: 0,
		},
		ExerciseToDo: "something",
		HRThreshold:  60,
	}
	gd := data.GenericData{
		Method: "alarm_start",
		Data:   as,
	}
	err = c.WriteJSON(gd)
	if err != nil {
		panic(err)
	}
	<-time.After(1 * time.Minute)

	fmt.Println("We are now dead....")

}
