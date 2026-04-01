package main

import "jarvis/config"

func main() {
	app := NewApp(config.Load())
	defer app.Close()
	app.Run()
}
