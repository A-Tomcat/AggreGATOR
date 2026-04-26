package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/atomcat/AggreGATOR/internal/config"
	"github.com/atomcat/AggreGATOR/internal/database"
	_ "github.com/lib/pq"
)

/*
goose -dir sql/schema postgres "postgres://atomcat:lulatsch12@localhost:5432/gator?sslmode=disable" up
*/

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	db, err := sql.Open("postgres", cfg.DbUrl)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer db.Close()
	s := state{
		cfg: &cfg,
		db:  database.New(db),
	}
	cmds := commands{
		Cmds: make(map[string]func(*state, command) error),
	}
	cmds.Register("login", handlerLogin)
	cmds.Register("register", handlerRegister)
	cmds.Register("reset", handlerReset)
	cmds.Register("users", handlerGetUsers)
	cmds.Register("agg", handlerAgg)
	cmds.Register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.Register("feeds", handlerFeeds)
	cmds.Register("follow", middlewareLoggedIn(handlerFollow))
	cmds.Register("following", middlewareLoggedIn(handlerFollowing))
	cmds.Register("unfollow", middlewareLoggedIn(handlerUnfollow))
	cmds.Register("browse", middlewareLoggedIn(handlerBrowse))

	args := os.Args
	if len(args) < 2 {
		fmt.Println("Not enough arguments. No Command given.")
		os.Exit(1)
	}

	cmd := command{name: args[1], args: args[2:]}

	if err = cmds.Run(&s, cmd); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
