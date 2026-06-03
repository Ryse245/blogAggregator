package main

import (
	internal "github.com/Ryse245/blogAggregator/internal"
	//config "github.com/Ryse245/blogAggregator/internal/config"
)

func main() {
	configGet := internal.Read()
	internal.PrintConfig(configGet)
	configGet.SetUser("lane")
	configGet = internal.Read()
	internal.PrintConfig(configGet)
}
