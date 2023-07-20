package opml

import (
	"encoding/xml"

	"github.com/nbd-wtf/go-nostr"
	"github.com/plantimals/go-opml/opml"
)

type OPMLParser struct {
	Path string `json:"path"`
	OPML *Opml  `json:"opml"`
}

type Opml struct {
	Title string `xml:"head>title"`
	Lists []List `xml:"body>outline"`
	Event *nostr.Event
}

type List struct {
	XMLName xml.Name `xml:"outline"`
	Title   string   `xml:"title,attr"`
	Feeds   []Feed   `xml:"body>outline>outline"`
	Event   *nostr.Event
}

type Feed struct {
	XMLName xml.Name `xml:"outline"`
	Title   string   `xml:"title,attr"`
	Url     string   `xml:"xmlUrl,attr"`
	Link    string   `xml:"htmlUrl,attr"`
	Event   *nostr.Event
}

func NewOPMLParser() *OPMLParser {
	return &OPMLParser{}
}

func (p *OPMLParser) Parse(path string) error {
	doc, err := opml.NewOPMLFromFile(path)
	if err != nil {
		return err
	}
	answer := &Opml{
		Title: doc.Head.Title,
	}

	for _, o := range doc.Body.Outlines {
		list := List{
			Title: o.Title,
		}
		for _, f := range o.Outlines {
			list.Feeds = append(list.Feeds, Feed{
				Title: f.Title,
				Url:   f.XMLURL,
				Link:  f.HTMLURL,
			})
		}
		answer.Lists = append(answer.Lists, list)
	}
	p.OPML = answer
	return nil
}

func (p *OPMLParser) FeedEventsFromOpml(path string, priv string, pk string) ([]*nostr.Event, error) {
	op := NewOPMLParser()
	if err := op.Parse(path); err != nil {
		return nil, err
	}
	var events []*nostr.Event
	for _, list := range p.OPML.Lists {
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
