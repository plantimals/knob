package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	nostr "github.com/fiatjaf/go-nostr"
	log "github.com/sirupsen/logrus"
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

var genkeys, pause bool
var path, relay, input string

func initFlags() {
	flag.BoolVar(&genkeys, "genkeys", false, "set to generate a new pub/priv key pair")
	flag.BoolVar(&pause, "pause", false, "set to pause for 10 seconds to allow debugger to attach")
	flag.StringVar(&relay, "relay", "wss://nostr.drss.io", "url of a relay to send to")
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

	//setup keys
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

	//set up relay
	pool := nostr.NewRelayPool()

	p := &RelayPolicy{
		Read:  true,
		Write: true,
	}

	err = pool.Add(relay, p)
	if err != nil {
		log.Fatal(err)
	}

	pool.SecretKey = &priv

	//allow attachment of debugger
	if pause {
		time.Sleep(10 * time.Second)
	}

	events := make(chan *nostr.Event)
	//process events
	go func(events chan *nostr.Event) {
		for evt := range events {
			evt.PubKey = pub
			err = evt.Sign(priv)
			if err != nil {
				panic(err)
			}
			ok, err := evt.CheckSignature()
			if err != nil {
				panic(err)
			}
			if !ok {
				log.Errorf("event failed signature check")
			} else {
				log.Info("event passed signature check")
			}

			event, statuses, _ := pool.PublishEvent(evt)

			wait := time.Tick(10 * time.Second)
		forLoop:
			for {
				select {
				case status := <-statuses:
					switch status.Status {
					case nostr.PublishStatusSent:
						fmt.Printf("Sent event %s to '%s'.\n", event.ID, status.Relay)
					case nostr.PublishStatusFailed:
						fmt.Printf("Failed to send event %s to '%s'.\n", event.ID, status.Relay)
					case nostr.PublishStatusSucceeded:
						fmt.Printf("Event seen %s on '%s'.\n", event.ID, status.Relay)
						break forLoop
					}
				case <-wait:
					log.Errorf("timeout exceeded for event: %s", event.ID)
					break forLoop
				}
			}

			ShowEvent(event)
		}

	}(events)

	if input != "" {
		EventFromInput(input, pub, events)
	} else if strings.HasSuffix(path, ".json") {
		EventsFromJson(path, events)
	} else if strings.HasSuffix(path, ".md") || strings.HasSuffix(path, ".txt") {
		EventFromText(path, pub, events)
	} else {
		fmt.Println("no inputs found")
	}
	fmt.Println("finished")
	close(events)
	time.Sleep(1 * time.Second)
}

func ShowEvent(evt *nostr.Event) {
	b, err := json.MarshalIndent(evt, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

func EventFromText(path, pk string, events chan *nostr.Event) {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	events <- &nostr.Event{
		CreatedAt: time.Now(),
		Kind:      nostr.KindTextNote,
		Tags:      make(nostr.Tags, 0),
		Content:   string(f),
		PubKey:    pk,
	}
}

func EventFromInput(input, pk string, events chan *nostr.Event) {
	events <- &nostr.Event{
		CreatedAt: time.Now(),
		Kind:      nostr.KindTextNote,
		Tags:      make(nostr.Tags, 0),
		Content:   input,
		PubKey:    pk,
	}
}

func EventsFromJson(path string, events chan *nostr.Event) {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	buf := bufio.NewScanner(f)
	for buf.Scan() {
		var evt *nostr.Event
		payload := []byte(buf.Text())
		json.Unmarshal(payload, &evt)
		if err != nil {
			panic(err)
		}
		evt.CreatedAt = time.Now()
		events <- evt
	}
}
