package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

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

	buf, err := ioutil.ReadFile(cDir + "/feedsync/config.json")
	if _, ok := err.(*os.PathError); ok {
		err = saveDefaultConfig()
		return conf, err
	} else if err != nil {
		return conf, err
	}

	err = json.Unmarshal(buf, &conf)
	if err != nil {
		return config{}, err
	}

	return conf, err
}

func saveDefaultConfig() error {
	cDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	err = os.MkdirAll(cDir+"/feedsync", 0755)
	if err != nil {
		return err
	}

	buf, err := json.Marshal(config{})
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(cDir+"/feedsync/config.json", buf, 0644)
	if err != nil {
		return err
	}

	return nil
}

func saveConfig(c config) (string, error) {
	cDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	filename := cDir + "/feedsync/config.json"

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
