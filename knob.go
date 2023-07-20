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
)

var genkeys, pause bool
var path, relay, input string
var waitTime int = 30

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
	var priv, pub string
	var err error
	if genkeys {
		priv, pub = GenerateKeysShow()
		if input == "" {
			os.Exit(0)
		}
	} else {
		priv = os.Getenv("NOSTR_KEY")
		pub, err = nostr.GetPublicKey(priv)
		if err != nil {
			panic(err)
		}
	}
	OPML := opml.NewOPMLParser()
	if err := OPML.Parse(path); err != nil {
		panic(err)
	}
	events, err := OPML.FeedEventsFromOpml(path, priv, pub)
	if err != nil {
		panic(err)
	}
	err = PublishEvents(events)
	if err != nil {
		panic(err)
	}
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
	for _, event := range events {
		ctx := context.Background()
		for _, url := range []string{"wss://relay.damus.io"} {
			relay, err := nostr.RelayConnect(ctx, url)
			if err != nil {
				fmt.Println(err)
				continue
			}
			_, err = relay.Publish(ctx, *event)
			if err != nil {
				fmt.Println(err)
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
