package internal

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"time"

	config "github.com/Ryse245/blogAggregator/internal/config"
	"github.com/Ryse245/blogAggregator/internal/database"
	"github.com/google/uuid"
)

func Read() config.Config {
	url := config.GetConfigFilePath()
	file, err := os.Open(url)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
	}
	defer file.Close()
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

func HandlerGetFeeds(s *config.State, cmd config.Command) error {
	feeds, err := s.DbPtr.GetFeeds(context.Background())
	for _, feed := range feeds {
		userName, err := s.DbPtr.GetUserFromID(context.Background(), feed.UserID)
		if err != nil {
			return err
		}
		fmt.Printf("Feed Name: %s, URL: %s, Owner: %s\n", feed.Name, feed.Url, userName.Name)
	}
	return err
}

func HandlerGetFeed(s *config.State, cmd config.Command) error {

	if len(cmd.Args) == 0 {
		return fmt.Errorf("No arguments provided")
	}
	/*
		url := "https://www.wagslane.dev/index.xml" //Stupid hard-coding for Boot Dev
		feed, err := FetchFeed(context.Background(), url)
		if err != nil {
			fmt.Println("Error in Fetch Feed call")
			return err
		}
		fmt.Printf("%v", feed)
		return nil*/
	time_between_str := cmd.Args[0]
	time_between, err := time.ParseDuration(time_between_str)
	if err != nil {
		return err
	}
	fmt.Printf("Collecting feeds every %v\n", time_between)
	ticker := time.NewTicker(time_between)
	for ; ; <-ticker.C {
		fmt.Println("Collecting new feed now...")
		ScrapeFeeds(s, cmd)
	}
}

func MiddlewareLoggedIn(handler func(s *config.State, cmd config.Command, user database.User) error) func(*config.State, config.Command) error {
	return func(s *config.State, cmd config.Command) error {
		currentUser, err := s.DbPtr.GetUser(context.Background(), s.ConfigPtr.Current_User_Name)
		if err != nil {
			return err
		}
		return handler(s, cmd, currentUser)
	}
}

func HandlerAddFeed(s *config.State, cmd config.Command, user database.User) error {
	if len(cmd.Args) < 2 {
		fmt.Println("Not enough arguments provided")
		os.Exit(1)
	}
	feedName := cmd.Args[0]
	feedUrl := cmd.Args[1]
	params := database.CreateFeedParams{ID: int32(uuid.New()[0]), CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: feedName,
		Url: feedUrl, UserID: user.ID}
	feed, err := s.DbPtr.CreateFeed(context.Background(), params)
	if err != nil {
		return err
	}
	fmt.Printf("%v", feed)

	newArg := [1]string{feedUrl}
	newCmd := config.Command{Name: cmd.Name, Args: newArg[:]}
	err = HandlerCreateFeedFollow(s, newCmd, user)
	return err
}

func FetchFeed(ctx context.Context, feedURL string) (*config.RSSFeed, error) {
	request, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		fmt.Println("Error in New Request With Context")
		return nil, err
	}
	request.Header.Set("User-Agent", "gator")
	var client http.Client
	response, err := client.Do(request)
	if err != nil {
		fmt.Println("Error in Client Do")
		return nil, err
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error in Read All")
		return nil, err
	}
	var feed config.RSSFeed
	err = xml.Unmarshal(body, &feed)
	if err != nil {
		fmt.Println("Error in Unmarshal")
		return nil, err
	}

	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	for i := 0; i < len(feed.Channel.Item); i++ {
		feed.Channel.Item[i].Title = html.UnescapeString(feed.Channel.Item[i].Title)
		feed.Channel.Item[i].Description = html.UnescapeString(feed.Channel.Item[i].Description)
	}

	return &feed, nil
}

func HandlerCreateFeedFollow(s *config.State, cmd config.Command, user database.User) error {
	if len(cmd.Args) < 1 {
		fmt.Println("Not enough arguments provided")
		os.Exit(1)
	}
	url := cmd.Args[0]
	feed, err := s.DbPtr.GetFeedFromUrl(context.Background(), url)
	if err != nil || feed.Url == "" {
		fmt.Println("Error in Get Feed From Url")
		return err
	}
	params := database.CreateFeedFollowParams{ID: int32(uuid.New()[0]), CreatedAt: time.Now(), UpdatedAt: time.Now(), UserID: user.ID, FeedID: feed.ID}
	data, err := s.DbPtr.CreateFeedFollow(context.Background(), params)
	if err != nil {
		return err
	}
	fmt.Printf("Feed: %s\n", data.FeedName)
	fmt.Printf("User: %s\n", data.UserName)
	return nil
}

func HandlerGetFeedFollowsFromUser(s *config.State, cmd config.Command, user database.User) error {
	follows, err := s.DbPtr.GetFeedFollowsFromUserId(context.Background(), user.ID)
	if err != nil {
		return err
	}
	fmt.Printf("Current User: %s\n", s.ConfigPtr.Current_User_Name)
	for _, row := range follows {
		fmt.Printf("- Following Feed: %s\n", row.FeedName)
	}
	return nil
}

func HandlerUnfollowFeed(s *config.State, cmd config.Command, user database.User) error {
	feedID, err := s.DbPtr.GetFeedFromUrl(context.Background(), cmd.Args[0])
	if err != nil {
		return err
	}
	param := database.DeleteFeedFollowsParams{UserID: user.ID, FeedID: feedID.ID}
	err = s.DbPtr.DeleteFeedFollows(context.Background(), param)
	if err != nil {
		return err
	}
	return nil
}

func ScrapeFeeds(s *config.State, cmd config.Command) error {
	nextFeed, err := s.DbPtr.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}
	err = s.DbPtr.MarkFeedFetched(context.Background(), nextFeed.ID)
	feedData, err := FetchFeed(context.Background(), nextFeed.Url)
	if err != nil {
		return err
	}
	for _, item := range feedData.Channel.Item {
		fmt.Printf("Item Title: %s\n", item.Title)
	}
	return nil
}
