package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	nostr "github.com/fiatjaf/go-nostr"
)

type RelayPolicy struct {
	Read  bool
	Write bool
}

func (rp *RelayPolicy) ShouldRead(f nostr.Filters) bool {
	return rp.Read
}

func (rp *RelayPolicy) ShouldWrite(e *nostr.Event) bool {
	return rp.Write
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: nostr <path>")
		os.Exit(1)
	}
	path := os.Args[1]
	f, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	pool := nostr.NewRelayPool()

	p := &RelayPolicy{
		Read:  true,
		Write: true,
	}

	err = pool.Add("wss://nostr.drss.io", p)
	if err != nil {
		log.Fatal(err)
	}

	k := os.Getenv("NOSTR_KEY")
	pool.SecretKey = &k

	event, statuses, _ := pool.PublishEvent(&nostr.Event{
		CreatedAt: time.Now(),
		Kind:      nostr.KindTextNote,
		Tags:      make(nostr.Tags, 0),
		Content:   string(f),
	})

	log.Printf("pubkey: %s", event.PubKey)
	log.Printf("event:  %s", event.ID)
	log.Printf("sig:    %s", event.Sig)

	for status := range statuses {
		switch status.Status {
		case nostr.PublishStatusSent:
			fmt.Printf("Sent event %s to '%s'.\n", event.ID, status.Relay)
		case nostr.PublishStatusFailed:
			fmt.Printf("Failed to send event %s to '%s'.\n", event.ID, status.Relay)
		case nostr.PublishStatusSucceeded:
			fmt.Printf("Event seen %s on '%s'.\n", event.ID, status.Relay)
		}
	}
}
