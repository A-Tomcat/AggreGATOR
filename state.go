package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/atomcat/AggreGATOR/internal/config"
	"github.com/atomcat/AggreGATOR/internal/database"
	"github.com/google/uuid"
)

type state struct {
	cfg *config.Config
	db  *database.Queries
}

type command struct {
	name string
	args []string
}

type commands struct {
	Cmds map[string]func(*state, command) error
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("No username given.\n")
	}
	username := cmd.args[0]

	name, _ := s.db.GetUser(context.Background(), username)
	if name.Name != username {
		os.Exit(1)
	}

	err := s.cfg.SetUser(username)
	if err != nil {
		return err
	}
	fmt.Printf("Username has been set to: %s\n", username)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("No username given. Cannot register.\n")
	}
	username := cmd.args[0]

	var u database.CreateUserParams
	u.ID = uuid.New()
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	u.Name = username
	newUser, err := s.db.CreateUser(context.Background(), u)
	if err != nil {
		return err
	}
	err = s.cfg.SetUser(username)
	if err != nil {
		return err
	}
	fmt.Printf("User: %s created.\nID: %d\nCreatedAT: %v\nUpdatedAT: %v\n", newUser.Name, newUser.ID, newUser.CreatedAt, newUser.UpdatedAt)

	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Wrong agrument given. Need time_between_reqs.\n")
	}
	time_between_reqs, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return err
	}
	fmt.Printf("Collecting feeds every %v\n", time_between_reqs)
	ticker := time.NewTicker(time_between_reqs)
	user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		return err
	}
	for ; ; <-ticker.C {
		err = scrapeFeeds(s, user)
		if err != nil {
			fmt.Printf("Error %v\n", err)
		}
	}
	return nil
}

func (c *commands) Run(s *state, cmd command) error {
	if command, ok := c.Cmds[cmd.name]; ok {
		err := command(s, cmd)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("Command does not exist.\n")
	}
	return nil
}

func (c *commands) Register(name string, f func(*state, command) error) {
	c.Cmds[name] = f
}

func handlerReset(s *state, cmd command) error {
	err := s.db.Reset(context.Background())
	if err != nil {
		return fmt.Errorf("Database not reset. Error: %v\n", err)
	}
	fmt.Println("Table successfully reset.")
	return nil
}

func handlerGetUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	for _, user := range users {
		name := user.Name
		if name == s.cfg.CurrentUserName {
			name = name + " (current)"
		}
		fmt.Printf("* %s\n", name)
	}

	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("No Feed-Name and/or url given.\n")
	}

	feed_name := cmd.args[0]
	feed_url := cmd.args[1]
	feedParams := database.CreateFeedsParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      feed_name,
		Url:       feed_url,
		UserID:    user.ID,
	}
	newFeed, err := s.db.CreateFeeds(context.Background(), feedParams)
	if err != nil {
		return err
	}

	params := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    newFeed.ID,
	}
	_, err = s.db.CreateFeedFollow(context.Background(), params)
	if err != nil {
		return err
	}
	fmt.Printf("ID: %v\nCreatedAT: %v\nUpdatedAT: %v\nName: %s, URL: %s, UserID: %v",
		newFeed.ID, newFeed.CreatedAt, newFeed.UpdatedAt, newFeed.Name, newFeed.Url, newFeed.UserID)

	return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}
	for _, feed := range feeds {
		user, err := s.db.GettUserFromID(context.Background(), feed.UserID)
		if err != nil {
			return err
		}
		fmt.Printf("Feed: %s\nURL: %s\nCreator: %s", feed.Name, feed.Url, user.Name)
	}

	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Wrong argument given. Need URL.\n")
	}
	feed, err := s.db.GetFeedFromUrl(context.Background(), cmd.args[0])
	params := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}
	newFollow, err := s.db.CreateFeedFollow(context.Background(), params)
	if err != nil {
		return err
	}
	fmt.Printf("Feed_Name: %s\nUsername: %s\n", newFollow.FeedName, newFollow.UserName)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	follows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return err
	}
	fmt.Printf("User: %s\nFollowing Feeds:\n", s.cfg.CurrentUserName)

	for _, feed := range follows {
		fmt.Printf("- %s", feed.FeedName)
	}

	return nil
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return err
		}
		return handler(s, cmd, user)
	}
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Wrong argument given. Need URL of Feed to unfollow.\n")
	}
	feed, err := s.db.GetFeedFromUrl(context.Background(), cmd.args[0])
	if err != nil {
		return err
	}
	params := database.UnfollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}
	err = s.db.Unfollow(context.Background(), params)
	if err != nil {
		return err
	}
	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	limit := 2
	if len(cmd.args) == 1 {
		num, err := strconv.Atoi(cmd.args[0])
		if err != nil {
			return fmt.Errorf("invalid limit: %w\n", err)
		}
		limit = num
	}
	params := database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	}
	posts, err := s.db.GetPostsForUser(context.Background(), params)
	if err != nil {
		return err
	}
	fmt.Printf("Posts for User: %s:\n", user.Name)
	for _, post := range posts {
		fmt.Printf(" - Title: %s\n", post.Title)
		fmt.Printf("- URL: %s\n", post.Url)
		fmt.Printf("- Description: %s\n", post.Description.String)
		fmt.Printf(" --- \n")
	}
	return nil
}
