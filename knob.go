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

func initFlags() {
	flag.BoolVar(&genkeys, "genkeys", false, "set to generate a new pub/priv key pair")
	flag.BoolVar(&pause, "pause", false, "set to pause for 10 seconds to allow debugger to attach")
	flag.StringVar(&relay, "relay", "wss://nostr.wine", "url of a relay to send to")
	flag.StringVar(&path, "file", "", "file path to .md .txt or .json")
	flag.StringVar(&input, "input", "", "input content")
	flag.Parse()
}

func genkeysShow() (string, string) {
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
		priv, pub = genkeysShow()
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
	events, err := FeedEventsFromOpml(path, priv, pub)
	if err != nil {
		panic(err)
	}

	for _, event := range events {
		ShowEvent(event)
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

			fmt.Printf("published to %s\n", url)
		}
	}
	time.Sleep(30 * time.Second)

	fmt.Println("close events and start wg.Wait()")
	fmt.Println("closing")
}

func ShowEvent(evt *nostr.Event) {
	b, err := json.MarshalIndent(evt, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
	note, err := nip19.EncodeNote(evt.ID)
	if err != nil {
		panic(err)
	}
	fmt.Printf("encoded: %s\n", note)
}

func EventsFromJson(path string, pk string) *nostr.Event {
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

func FeedEventsFromOpml(path string, priv string, pk string) ([]*nostr.Event, error) {
	op := opml.NewOPMLParser()

	opml, err := op.Parse(path)
	if err != nil {
		return nil, err
	}
	var events []*nostr.Event
	for _, list := range opml.Lists {
		for _, feed := range list.Feeds {
			evt := &nostr.Event{
				Content:   feed.Title,
				CreatedAt: nostr.Now(),
				Kind:      1063,
				PubKey:    pk,
			}
			evt.Tags = append(evt.Tags, []string{"url", feed.Url})
			evt.Tags = append(evt.Tags, []string{"m", "application/rss+xml"})
			evt.Tags = append(evt.Tags, []string{"link", feed.Link})
			evt.Sign(priv)
			events = append(events, evt)
		}
	}
	return events, nil
}
