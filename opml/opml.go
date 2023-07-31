package opml

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
	"github.com/plantimals/go-opml/opml"
	log "github.com/sirupsen/logrus"
)

type OPMLParser struct {
	Path   string `json:"path"`
	OPML   *Opml  `json:"opml"`
	Priv   string
	Pub    string
	Relays []string
}

type Opml struct {
	Title    string    `xml:"head>title"`
	Outlines []Outline `xml:"body>outline"`
	Event    *nostr.Event
	NAddr    string
}

type Outline struct {
	XMLName  xml.Name  `xml:"outline"`
	Title    string    `xml:"title,attr"`
	Outlines []Outline `xml:"body>outline>outline,omitempty"`
	Url      string    `xml:"xmlUrl,attr,omitempty"`
	Link     string    `xml:"htmlUrl,attr,omitempty"`
	Event    *nostr.Event
	NAddr    string
}

// type Feed struct {
// 	XMLName xml.Name `xml:"outline"`
// 	Title   string   `xml:"title,attr"`
// 	Url     string   `xml:"xmlUrl,attr"`
// 	Link    string   `xml:"htmlUrl,attr"`
// 	Event   *nostr.Event
// }

func NewOPMLParser() *OPMLParser {
	return &OPMLParser{}
}

func (p *OPMLParser) Parse(path string, key string, relays []string) error {
	doc, err := opml.NewOPMLFromFile(path)
	if err != nil {
		return err
	}
	if strings.Contains(key, "nsec") {
		pref, k, err := nip19.Decode(key)
		if err != nil {
			log.Error(err)
			return err
		}
		if pref != "nsec" {
			return fmt.Errorf("invalid key prefix: %s", pref)
		}
		key = k.(string)
	}
	p.Priv = key
	pub, err := nostr.GetPublicKey(key)
	if err != nil {
		log.Error(err)
		return err
	}
	p.Pub = pub
	p.Relays = relays
	fmt.Printf("priv: %s\tpub: %s\n", p.Priv, p.Pub)
	answer := &Opml{
		Title: doc.Head.Title,
	}
	drssEvent := &nostr.Event{
		CreatedAt: nostr.Now(),
		Kind:      30001,
		PubKey:    p.Pub,
	}
	drssEvent.Tags = append(drssEvent.Tags, []string{"d", answer.Title})

	for _, o := range doc.Body.Outlines {
		outline := Outline{
			Title: o.Title,
		}
		for _, f := range o.Outlines {
			ol := p.parseOutline(&f)
			outline.Outlines = append(outline.Outlines, *ol)
		}
		event := &nostr.Event{
			CreatedAt: nostr.Now(),
			Kind:      30001,
			PubKey:    p.Pub,
		}
		event.Tags = append(event.Tags, []string{"d", outline.Title})
		naddr, err := nip19.EncodeEntity(p.Pub, 30001, outline.Title, p.Relays)
		if err != nil {
			log.Error(err)
			panic(err)
		}
		outline.NAddr = naddr

		for _, o := range outline.Outlines {
			if o.Event.Kind == 1063 {
				event.Tags = append(event.Tags, []string{"e", o.Event.ID})
			} else {
				event.Tags = append(event.Tags, []string{"a", fmt.Sprintf("%d:%s:%s", 30001, p.Pub, o.Title), p.Relays[0]})
			}
		}
		event.Sign(p.Priv)
		outline.Event = event
		answer.Outlines = append(answer.Outlines, outline)

		drssEvent.Tags = append(drssEvent.Tags, []string{"a", fmt.Sprintf("%d:%s:%s", 30001, p.Pub, outline.Title), relays[0]})
	}

	naddr, err := nip19.EncodeEntity(p.Pub, 30001, answer.Title, p.Relays)
	if err != nil {
		log.Error(err)
		panic(err)
	}
	drssEvent.Sign(p.Priv)
	answer.NAddr = naddr
	answer.Event = drssEvent
	p.OPML = answer
	return nil
}

func (p *OPMLParser) parseOutline(outline *opml.Outline) *Outline {
	ol := Outline{
		Title: outline.Title,
	}
	if outline.XMLURL != "" {
		ol.Url = outline.XMLURL
		ol.Link = outline.HTMLURL
		event := &nostr.Event{
			Content:   outline.Title,
			CreatedAt: nostr.Now(),
			Kind:      1063,
			PubKey:    p.Pub,
		}
		event.Tags = append(event.Tags, []string{"url", outline.XMLURL})
		event.Tags = append(event.Tags, []string{"m", "application/rss+xml"})
		event.Tags = append(event.Tags, []string{"link", outline.HTMLURL})
		event.Sign(p.Priv)
		ol.Event = event
	} else {
		for _, o := range outline.Outlines {
			o := p.parseOutline(&o)
			ol.Outlines = append(ol.Outlines, *o)
		}
		event := &nostr.Event{
			CreatedAt: nostr.Now(),
			Kind:      30001,
			PubKey:    p.Pub,
		}
		naddr, err := nip19.EncodeEntity(p.Pub, 30001, ol.Title, p.Relays)
		if err != nil {
			log.Error(err)
			panic(err)
		}
		ol.NAddr = naddr
		event.Tags = append(event.Tags, []string{"d", outline.Title})
		for _, o := range ol.Outlines {
			if o.Event.Kind == 1063 {
				event.Tags = append(event.Tags, []string{"e", o.Event.ID})
			} else {
				event.Tags = append(event.Tags, []string{"a", fmt.Sprintf("%d:%s:%s", 30001, p.Pub, o.Title), p.Relays[0]})
			}
		}
		event.Sign(p.Priv)
		ol.Event = event
	}
	return &ol
}

func (p *OPMLParser) Events() []*nostr.Event {
	var events []*nostr.Event
	events = append(events, p.OPML.Event)
	for _, o := range p.OPML.Outlines {
		events = append(events, getOutlineEvents(o)...)
	}
	return events
}

func getOutlineEvents(outline Outline) []*nostr.Event {
	var events []*nostr.Event
	if outline.Event != nil {
		events = append(events, outline.Event)
	} else {
		fmt.Println(outline.NAddr)
		fmt.Printf("empty outline event: %s\n", outline.Title)
	}
	for _, o := range outline.Outlines {
		events = append(events, getOutlineEvents(o)...)
	}
	return events
}

// func (p *OPMLParser) FeedEventsFromOpml(path string, priv string, pk string) ([]*nostr.Event, error) {
// 	op := NewOPMLParser()
// 	if err := op.Parse(path); err != nil {
// 		return nil, err
// 	}
// 	var events []*nostr.Event
// 	for _, list := range p.OPML.Lists {
// 		for _, feed := range list.Feeds {
// 			evt := &nostr.Event{
// 				Content:   feed.Title,
// 				CreatedAt: nostr.Now(),
// 				Kind:      1063,
// 				PubKey:    pk,
// 			}
// 			evt.Tags = append(evt.Tags, []string{"url", feed.Url})
// 			evt.Tags = append(evt.Tags, []string{"m", "application/rss+xml"})
// 			evt.Tags = append(evt.Tags, []string{"link", feed.Link})
// 			evt.Sign(priv)
// 			events = append(events, evt)
// 		}
// 	}
// 	return events, nil
// }
