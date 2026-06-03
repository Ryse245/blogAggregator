package internal

import (
	"encoding/json"
	"fmt"
	"os"

	config "github.com/Ryse245/blogAggregator/internal/config"
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
