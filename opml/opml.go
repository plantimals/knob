package opml

import (
	"encoding/xml"

	"github.com/plantimals/go-opml/opml"
)

type OPMLParser struct {
	Path string `json:"path"`
	OPML *Opml  `json:"opml"`
}

type Opml struct {
	Title string `xml:"head>title"`
	Lists []List `xml:"body>outline"`
}

type List struct {
	XMLName xml.Name `xml:"outline"`
	Title   string   `xml:"title,attr"`
	Feeds   []Feed   `xml:"body>outline>outline"`
}

type Feed struct {
	XMLName xml.Name `xml:"outline"`
	Title   string   `xml:"title,attr"`
	Url     string   `xml:"xmlUrl,attr"`
	Link    string   `xml:"htmlUrl,attr"`
}

func NewOPMLParser() *OPMLParser {
	return &OPMLParser{}
}

// func (p *OPMLParser) Parse(path string) (*Opml, error) {
// 	f, err := os.ReadFile(path)
// 	if err != nil {
// 		return nil, err
// 	}
// 	p.Path = path
// 	p.OPML = &Opml{}
// 	if err := xml.Unmarshal(f, p.OPML); err != nil {
// 		return nil, err
// 	}
// 	return p.OPML, nil
// }

func (p *OPMLParser) Parse(path string) (*Opml, error) {
	doc, err := opml.NewOPMLFromFile(path)
	if err != nil {
		return nil, err
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
	return answer, nil
}
