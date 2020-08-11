package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/mmcdole/gofeed"
)

type app struct {
	conf   config
	client *http.Client
}

func main() {
	conf, err := openConfig()
	if err != nil {
		log.Fatal(err)
	}

	a := &app{
		conf:   conf,
		client: &http.Client{Timeout: 10 * time.Second},
	}

	for i, feed := range conf.Feeds {
		fp := gofeed.NewParser()
		f, err := fp.ParseURL(feed.Link)
		if err != nil {
			log.Fatal(err)
		}

		lastLink := feed.LastLink

		for j, entry := range f.Items {
			if entry.Link == feed.LastLink {
				break
			}

			err := a.addArticle(entry.Link)
			if err != nil {
				log.Fatal(err)
			}

			if j == 0 {
				lastLink = entry.Link
			}
		}

		conf.Feeds[i].LastLink = lastLink
	}

	_, err = saveConfig(conf)
	if err != nil {
		log.Fatal(err)
	}
}

type config struct {
	ConsumerKey string      `json:"consumer_key"`
	AccessToken string      `json:"access_token"`
	Username    string      `json:"username"`
	Feeds       []feedEntry `json:"feeds"`
}

type feedEntry struct {
	Link     string `json:"link"`
	LastLink string `json:"last_link"`
}

func openConfig() (conf config, err error) {
	cDir, err := os.UserConfigDir()
	if err != nil {
		return conf, err
	}

	buf, err := ioutil.ReadFile(cDir + "/rss2pocket/config.json")
	if err != nil {
		return conf, err
	}

	err = json.Unmarshal(buf, &conf)
	if err != nil {
		return config{}, err
	}

	return conf, err
}

func saveConfig(c config) (string, error) {
	cDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	filename := cDir + "/rss2pocket/config.json"

	buf, err := json.Marshal(c)
	if err != nil {
		return "", err
	}

	err = ioutil.WriteFile(filename, buf, 0644)
	if err != nil {
		return "", err
	}

	return filename, nil
}

func (a *app) addArticle(url string) error {
	body, err := json.Marshal(struct {
		URL         string `json:"url"`
		ConsumerKey string `json:"consumer_key"`
		AccessToken string `json:"access_token"`
	}{
		URL:         url,
		ConsumerKey: a.conf.ConsumerKey,
		AccessToken: a.conf.AccessToken,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://getpocket.com/v3/add", bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("X-Accept", "application/json")

	_, err = a.client.Do(req)
	if err != nil {
		return err
	}

	return nil
}

func authFlow(consumerKey string) {
	// We open a listener to grab a random open port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}

	err = listener.Close()
	if err != nil {
		log.Fatal(err)
	}

	doneChan := make(chan struct{})

	s := &http.Server{
		Addr:    listener.Addr().String(),
		Handler: handleAuth(doneChan),
	}
	fmt.Printf("Listening on %s\n", s.Addr)

	go func() {
		s.ListenAndServe()
	}()

	client := &http.Client{}

	reqBody := struct {
		ConsumerKey string `json:"consumer_key"`
		RedirectURI string `json:"redirect_uri"`
	}{
		ConsumerKey: consumerKey,
		RedirectURI: "http://" + s.Addr,
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("POST", "https://getpocket.com/v3/oauth/request", bytes.NewBuffer(b))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("X-Accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	} else if res.StatusCode >= 300 {
		log.Fatalf("got status code %d", res.StatusCode)
	}

	resBody := struct {
		Code string `json:"code"`
	}{}

	err = json.NewDecoder(res.Body).Decode(&resBody)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Go to https://getpocket.com/auth/authorize?request_token=%s&redirect_uri=%s to continue\n", resBody.Code, "http://"+s.Addr)

	<-doneChan
	err = s.Close()
	if err != nil {
		log.Fatal(err)
	}

	authorizeReqBody := struct {
		ConsumerKey string `json:"consumer_key"`
		Code        string `json:"code"`
	}{
		ConsumerKey: consumerKey,
		Code:        resBody.Code,
	}

	b, err = json.Marshal(authorizeReqBody)
	if err != nil {
		log.Fatal(err)
	}

	req, err = http.NewRequest("POST", "https://getpocket.com/v3/oauth/authorize", bytes.NewBuffer(b))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("X-Accept", "application/json")

	res, err = client.Do(req)
	if err != nil {
		log.Fatal(err)
	} else if res.StatusCode >= 300 {
		log.Fatalf("got status code %d", res.StatusCode)
	}

	authorizeResBody := struct {
		AccessToken string `json:"access_token"`
		Username    string `json:"username"`
	}{}
	err = json.NewDecoder(res.Body).Decode(&authorizeResBody)
	if err != nil {
		log.Fatal(err)
	}

	cFile, err := saveConfig(config{
		ConsumerKey: consumerKey,
		AccessToken: authorizeResBody.AccessToken,
		Username:    authorizeResBody.Username,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Access token: %s\n", authorizeResBody.AccessToken)
	fmt.Printf("Username: %s\n", authorizeResBody.Username)
	fmt.Printf("\nConfiguration written to %s\n", cFile)
}

func handleAuth(done chan<- struct{}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		done <- struct{}{}
	})
}
