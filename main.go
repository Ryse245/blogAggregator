package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"

	internal "github.com/Ryse245/blogAggregator/internal"

	config "github.com/Ryse245/blogAggregator/internal/config"
	"github.com/Ryse245/blogAggregator/internal/database"
)

func main() {
	configGet := internal.Read()

	db, err := sql.Open("postgres", configGet.Db_Url)
	dbQueries := database.New(db)

	blogState := config.State{ConfigPtr: &configGet, DbPtr: dbQueries}
	blogCommands := config.Commands{CommandMap: map[string]func(*config.State, config.Command) error{}}
	//Add commands here
	blogCommands.Register("login", internal.HandlerLogin)
	blogCommands.Register("register", internal.HandlerRegister)
	blogCommands.Register("reset", internal.HanddlerReset)
	blogCommands.Register("users", internal.HandlerGetUsers)
	blogCommands.Register("agg", internal.HandlerGetFeed)
	blogCommands.Register("addfeed", internal.HandlerAddFeed)
	blogCommands.Register("feeds", internal.HandlerGetFeeds)

	fullArguments := os.Args

	if len(fullArguments) < 2 {
		fmt.Println("Too few arguments")
		os.Exit(1)
	}
	justCommand := fullArguments[1]
	additionalArgs := fullArguments[2:]
	commandData := config.Command{Name: justCommand, Args: additionalArgs}
	err = blogCommands.Run(&blogState, commandData)
	if err != nil {
		fmt.Printf("Error on running command: %v\n", err)
		os.Exit(1)
	}
}
