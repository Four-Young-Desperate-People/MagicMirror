package main

import (
	"MagicMirrorGo/websockets"
	"fmt"
)

func main() {
	fmt.Println("Hello World")
	//go websockets.Client()
	websockets.Run()
}
