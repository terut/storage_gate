package store

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Buckets []Bucket `json:"buckets"`
}

type Bucket struct {
	Name  string `json:"name"`
	Token string `json:"token"`
}

type store struct {
	Buckets map[string]Bucket
}

var instance *store = newStore()

func newStore() *store {
	file, err := ioutil.ReadFile("./config/app.json")
	if err != nil {
		panic(err)
	}
	var config Config
	json.Unmarshal(file, &config)

	buckets := map[string]Bucket{}
	for _, bucket := range config.Buckets {
		buckets[bucket.Token] = bucket
	}

	return &store{Buckets: buckets}
}

func GetInstance() *store {
	return instance
}
