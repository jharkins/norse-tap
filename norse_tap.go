package main

import (
	"net/url"
	"log"
	"github.com/gorilla/websocket"
	"time"
	"sync"
	"encoding/json"
)

func main() {
	u := url.URL{
		Scheme: "ws",
		Host:   "map.norsecorp.com",
		Path:   "/socketcluster",
	}

	var wg sync.WaitGroup
	wg.Add(1)
	tap(1, u, &wg)
	wg.Wait()
}

type Message struct {
	Event string `json:"event"`
	Data struct {
		Channel string `json:"channel"`
		Data []struct {
			VectorID     int64 `json:"vector_id"`
			Latitude     float64 `json:"latitude"`
			Longitude    float64 `json:"longitude"`
			Countrycode  string `json:"countrycode"`
			Country      string `json:"country"`
			City         string `json:"city"`
			Latitude2    float64 `json:"latitude2"`
			Longitude2   float64 `json:"longitude2"`
			Countrycode2 string `json:"countrycode2"`
			Country2     string `json:"country2"`
			City2        string `json:"city2"`
			Md5          string `json:"md5"`
			Dport        int `json:"dport"`
			Svc          int `json:"svc"`
			Type         string `json:"type"`
			Org          string `json:"org"`
			Zerg         string `json:"zerg"`
		} `json:"data"`
	} `json:"data"`
	Cid int `json:"cid"`
}

func tap(tapID int, u url.URL, wg *sync.WaitGroup) {
	log.Printf("%d - connecting to %s", tapID, u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("%d - dial error", tapID)
	}
	defer c.Close()

	sendCh := make(chan string)
	rcvCh := make(chan []byte)
	errCh := make(chan error)

	go func() {
		for m := range sendCh {
			err := c.WriteMessage(websocket.TextMessage, []byte(m))
			if err != nil {
				log.Println("write error:", err)
				return
			}
		}
	}()

	go func() {
		for e := range errCh {
			log.Fatalln("errCh:", e)
		}
	}()

	log.Println("Handshake!")
	sendCh <- "{\"event\":\"#handshake\",\"data\":{\"authToken\":null},\"cid\":1}"
	log.Println("Subscribe!")
	sendCh <- "{\"event\":\"#subscribe\",\"data\":{\"channel\":\"global\"},\"cid\":2}"

	go func() {
		defer c.Close()
		defer wg.Done()
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				errCh <- err
				return
			}
			rcvCh <- message
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case m := <-rcvCh:
			if string(m) == "#1" {
				sendCh <- "#2"
				break
			}

			var f Message
			err := json.Unmarshal(m, &f)
			if err != nil {
				errCh <- err
				break
			}

			log.Printf("%+v", f)

		case t := <-ticker.C:
			sendCh <- t.String()
		}
	}
}
