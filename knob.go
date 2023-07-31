package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	nostr "github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
	"github.com/plantimals/knob/opml"
	log "github.com/sirupsen/logrus"
)

var genkeys, pause bool
var path, relay, input string
var waitTime int = 60
var relays = []string{
	"wss://relay.damus.io",
}

func initFlags() {
	flag.BoolVar(&genkeys, "genkeys", false, "set to generate a new pub/priv key pair")
	flag.BoolVar(&pause, "pause", false, "set to pause for 10 seconds to allow debugger to attach")
	flag.StringVar(&relay, "relay", "wss://nostr.wine", "url of a relay to send to")
	flag.StringVar(&path, "file", "", "file path to .md .txt or .json")
	flag.StringVar(&input, "input", "", "input content")
	flag.Parse()
}

func GenerateKeysShow() (string, string) {
	priv := nostr.GeneratePrivateKey()
	pub, err := nostr.GetPublicKey(priv)
	if err != nil {
		panic(err)
	}
	fmt.Printf("{\"private_key\":\"%s\",\"public_key\":\"%s\"}\n", priv, pub)
	return priv, pub
}

func main() {
	initFlags()
	var priv string
	var err error
	if genkeys {
		priv, _ = GenerateKeysShow()
		if input == "" {
			os.Exit(0)
		}
	} else {
		priv = os.Getenv("NOSTR_KEY")
		if err != nil {
			panic(err)
		}
	}
	OPML := opml.NewOPMLParser()
	if err := OPML.Parse(path, priv, relays); err != nil {
		panic(err)
	}
	events := OPML.Events()

	ShowEvents(events)

	err = PublishEvents(events)
	if err != nil {
		panic(err)
	}

	fmt.Printf("drss id: %s\n", OPML.OPML.NAddr)
}

func ShowEvent(evt *nostr.Event) {
	b, err := json.MarshalIndent(evt, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
	note := EncodeEvent(evt)
	fmt.Printf("encoded: %s\n", note)
}

func ShowEvents(events []*nostr.Event) {
	for _, evt := range events {
		if evt.Kind == 1063 {
			fmt.Printf("%d\n", evt.Kind)
		} else {
			fmt.Printf("%d %s %s\n", evt.Kind, evt.Tags[0][0], evt.Tags[0][1])
		}
	}
}

func EncodeEvent(evt *nostr.Event) string {
	note, err := nip19.EncodeNote(evt.ID)
	if err != nil {
		panic(err)
	}
	return note
}

func EventFromJson(path string, pk string) *nostr.Event {
	payload, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	var evt *nostr.Event
	err = json.Unmarshal(payload, &evt)
	if err != nil {
		panic(err)
	}
	evt.CreatedAt = nostr.Now()
	return evt
}

func PublishEvents(events []*nostr.Event) error {
	notes := make(map[string]int, 0)
	for _, url := range relays {
		ctx := context.Background()
		relay, err := nostr.RelayConnect(ctx, url)
		if err != nil {
			log.Info("failed to connect to relay")
			log.Error(err)
			continue
		}
		for _, event := range events {
			ShowEvent(event)
			_, err = relay.Publish(ctx, *event)
			if err != nil {
				log.Info("failed to publish")
				log.Error(err)
				time.Sleep(time.Duration(waitTime) * time.Second)
				continue
			}
			notes[EncodeEvent(event)] = 1
			fmt.Printf("published to %s\n", url)
			fmt.Printf("waiting for %d seconds\n", waitTime)
			time.Sleep(time.Duration(waitTime) * time.Second)
		}
	}
	fmt.Printf("published %d events\n", len(notes))
	for note := range notes {
		fmt.Printf("%s\n", note)
	}
	return nil
}
