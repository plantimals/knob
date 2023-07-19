package feed

type DRSS struct {
	Name  string `json:"name"`
	Lists []List `json:"lists"`
}

type List struct {
	Name  string `json:"name"`
	Feeds []Feed `json:"feeds"`
}

type Feed struct {
	Name string `json:"name"`
	Url  string `json:"url"`
	Link string `json:"link"`
}
