package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	config "github.com/Ryse245/blogAggregator/internal/config"
	"github.com/Ryse245/blogAggregator/internal/database"
	"github.com/google/uuid"
)

func Read() config.Config {
	url := config.GetConfigFilePath()
	file, err := os.Open(url)
	defer file.Close()
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
	}
	var parseConfig config.Config
	parser := json.NewDecoder(file)
	if err = parser.Decode(&parseConfig); err != nil {
		fmt.Printf("Error parsing: %v\n", err)
	}

	return parseConfig
}

func PrintConfig(cfg config.Config) {
	fmt.Printf("Database URL: %s\n", cfg.Db_Url)
	fmt.Printf("Current User Name: %s\n", cfg.Current_User_Name)
}

func HandlerLogin(s *config.State, cmd config.Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("No arguments provided")
	}
	dbUser, _ := s.DbPtr.GetUser(context.Background(), cmd.Args[0])
	if dbUser.Name == "" {
		fmt.Printf("Cannot login to user %s: user not in database\n", cmd.Args[0])
		os.Exit(1)
	}

	s.ConfigPtr.SetUser(cmd.Args[0])
	fmt.Println("User has been set")
	return nil
}

func HandlerRegister(s *config.State, cmd config.Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("No arguments provided")
	}
	dbUser, _ := s.DbPtr.GetUser(context.Background(), cmd.Args[0])
	if dbUser.Name != "" {
		fmt.Printf("Error when registering: user %s found", dbUser.Name)
		os.Exit(1)
	}

	createUserParams := database.CreateUserParams{ID: int32(uuid.New()[0]), CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: cmd.Args[0]}
	user, err := s.DbPtr.CreateUser(context.Background(), createUserParams)
	if err != nil {
		return fmt.Errorf("Error in creating user: %v\n", err)
	}
	s.ConfigPtr.SetUser(user.Name)
	fmt.Printf("User %s created\nData: \nCreated at: %v\nUpdated at: %v\nUser ID: %v\n", user.Name, user.CreatedAt, user.UpdatedAt, user.ID)
	return nil
}

func HanddlerReset(s *config.State, cmd config.Command) error {
	err := s.DbPtr.DeleteUsers(context.Background())
	return err
}

func HandlerGetUsers(s *config.State, cmd config.Command) error {
	users, err := s.DbPtr.GetUsers(context.Background())
	for _, user := range users {
		userName := user.Name
		if userName == s.ConfigPtr.Current_User_Name {
			userName += " (current)"
		}
		fmt.Printf("* %s\n", userName)
	}
	return err
}
