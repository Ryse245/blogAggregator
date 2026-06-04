package main

import (
	"fmt"
	"os"

	internal "github.com/Ryse245/blogAggregator/internal"
	config "github.com/Ryse245/blogAggregator/internal/config"
)

func main() {
	configGet := internal.Read()
	blogState := config.State{ConfigPtr: &configGet}
	blogCommands := config.Commands{CommandMap: map[string]func(*config.State, config.Command) error{}}
	//Add commands here
	blogCommands.Register("login", internal.HandlerLogin)

	fullArguments := os.Args

	if len(fullArguments) < 2 {
		fmt.Println("Too few arguments")
		os.Exit(1)
	}
	justCommand := fullArguments[1]
	additionalArgs := fullArguments[2:]
	commandData := config.Command{Name: justCommand, Args: additionalArgs}
	err := blogCommands.Run(&blogState, commandData)
	if err != nil {
		fmt.Printf("Error on running command: %v\n", err)
		os.Exit(1)
	}
}
