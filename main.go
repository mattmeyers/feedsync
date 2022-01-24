package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/mattmeyers/feedsync/pocket"
	"github.com/mmcdole/gofeed"
	"github.com/urfave/cli/v2"
)

type handler struct {
	conf   config
	client *pocket.Client
}

func main() {
	conf, err := openConfig()
	if err != nil {
		log.Fatal(err)
	}

	h := handler{
		conf:   conf,
		client: pocket.NewClient(conf.ConsumerKey, conf.AccessToken),
	}

	app := &cli.App{
		Name:  "feedsync",
		Usage: "synchronize RSS/Atom feeds to your Pocket",
		Commands: []*cli.Command{
			{
				Name:   "list",
				Usage:  "list all subscriptions",
				Action: h.handleList,
			},
			{
				Name:   "add",
				Usage:  "begin synchronizing a new feed",
				Action: h.handleAdd,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "sync",
						Usage: "run the sync operation after adding the new feed",
					},
				},
			},
			{
				Name:   "remove",
				Usage:  "stop synchronizing a feed",
				Action: h.handleRemove,
			},
			{
				Name:   "sync",
				Usage:  "run the sync process",
				Action: h.handleSync,
			},
			{
				Name:   "authenticate",
				Usage:  "retrieve an access token",
				Action: h.handleAuthenticate,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "overwrite",
						Usage: "overwrite the existing access token",
					},
					&cli.StringFlag{
						Name:  "consumer-key",
						Usage: "Pocket app's consumer key",
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func (h *handler) handleList(c *cli.Context) error {
	for i, item := range h.conf.Feeds {
		fmt.Printf("[%d] %s\n", i+1, item.Link)
	}

	return nil
}

func (h *handler) handleAdd(c *cli.Context) error {
	l := c.Args().First()
	if l == "" {
		return errors.New("link required")
	}

	for _, item := range h.conf.Feeds {
		if item.Link == l {
			return errors.New("already subscribed")
		}
	}

	h.conf.Feeds = append(h.conf.Feeds, feedEntry{Link: l})

	if c.Bool("sync") {
		err := h.handleSync(c)
		if err != nil {
			return err
		}
	} else {
		_, err := saveConfig(h.conf)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *handler) handleRemove(c *cli.Context) error {
	if !c.Args().Present() {
		return errors.New("index required")
	}

	i, err := strconv.Atoi(c.Args().First())
	if err != nil {
		return errors.New("invalid index")
	}

	i--
	if i < 0 || i > len(h.conf.Feeds)-1 {
		return errors.New("invalid index")
	}

	h.conf.Feeds = append(h.conf.Feeds[:i], h.conf.Feeds[i+1:]...)

	_, err = saveConfig(h.conf)
	if err != nil {
		return err
	}

	return nil
}

func (h *handler) handleSync(c *cli.Context) error {
	for i, feed := range h.conf.Feeds {
		fp := gofeed.NewParser()
		f, err := fp.ParseURL(feed.Link)
		if err != nil {
			return err
		}

		lastLink := feed.LastLink

		for j, entry := range f.Items {
			if entry.Link == feed.LastLink {
				break
			}

			err := h.addArticle(entry.Link)
			if err != nil {
				return err
			}

			if lastLink == "" {
				lastLink = entry.Link
				break
			}

			if j == 0 {
				lastLink = entry.Link
			}
		}

		h.conf.Feeds[i].LastLink = lastLink
	}

	_, err := saveConfig(h.conf)
	if err != nil {
		return err
	}

	return nil
}

func (h *handler) addArticle(url string) error {
	return h.client.AddLink(url)
}

func (h *handler) handleAuthenticate(c *cli.Context) error {
	if h.conf.AccessToken != "" && !c.Bool("overwrite") {
		return errors.New("access token already set, use --overwrite to replace")
	}

	ck := c.String("consumer-key")
	if ck == "" {
		ck = h.conf.ConsumerKey
		if ck == "" {
			return errors.New("consumer key must be provided in config file or --consumer-key")
		}
	}

	err := authFlow(ck)
	if err != nil {
		return err
	}

	return nil
}
