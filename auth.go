package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
)

func authFlow(consumerKey string) error {
	// We open a listener to grab a random open port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}

	err = listener.Close()
	if err != nil {
		return err
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
		return err
	}

	req, err := http.NewRequest("POST", "https://getpocket.com/v3/oauth/request", bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("X-Accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return err
	} else if res.StatusCode >= 300 {
		log.Fatalf("got status code %d", res.StatusCode)
	}

	resBody := struct {
		Code string `json:"code"`
	}{}

	err = json.NewDecoder(res.Body).Decode(&resBody)
	if err != nil {
		return err
	}

	fmt.Printf("Go to https://getpocket.com/auth/authorize?request_token=%s&redirect_uri=%s to continue\n", resBody.Code, "http://"+s.Addr)

	<-doneChan
	err = s.Close()
	if err != nil {
		return err
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
		return err
	}

	req, err = http.NewRequest("POST", "https://getpocket.com/v3/oauth/authorize", bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("X-Accept", "application/json")

	res, err = client.Do(req)
	if err != nil {
		return err
	} else if res.StatusCode >= 300 {
		log.Fatalf("got status code %d", res.StatusCode)
	}

	authorizeResBody := struct {
		AccessToken string `json:"access_token"`
		Username    string `json:"username"`
	}{}
	err = json.NewDecoder(res.Body).Decode(&authorizeResBody)
	if err != nil {
		return err
	}

	cFile, err := saveConfig(config{
		ConsumerKey: consumerKey,
		AccessToken: authorizeResBody.AccessToken,
		Username:    authorizeResBody.Username,
	})
	if err != nil {
		return err
	}

	log.Printf("Username: %s\n", authorizeResBody.Username)
	log.Printf("Access token: %s\n", authorizeResBody.AccessToken)
	log.Printf("Configuration written to %s\n", cFile)

	return nil
}

const authSuccessTmpl string = `<!DOCTYPE html>
<html lang="en">

<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>feedsync</title>
  <style>
    * {
      margin: 0;
      padding: 0;
    }

    .header {
      padding: 30px;
      width: 100%;
      text-align: left;
      background: #282828;
      color: white;
      font-size: 24px;
    }

    .content {
      text-align: center;
      margin-top: 16rem;
      font-size: 1.5rem;
    }
  </style>
</head>

<body>
  <header class="header">
    <h1>feedsync</h1>
  </header>
  <div class="content">
    <h2>Authentication Complete</h2>
    <div>Please return to your terminal.</div>
  </div>
</body>

</html>
`

func handleAuth(done chan<- struct{}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(authSuccessTmpl))
		done <- struct{}{}
	})
}
