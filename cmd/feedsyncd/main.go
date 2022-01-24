package main

import (
	"fmt"
	"os"

	"github.com/mattmeyers/feedsync/http"
	"github.com/mattmeyers/feedsync/store/bolt"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(argv []string) error {
	db, err := bolt.Open("./feedsync-dev.bolt")
	if err != nil {
		return err
	}

	feedStore := bolt.NewFeedStore(db)

	return http.NewServer(feedStore).ListenAndServe(":8080")
}
