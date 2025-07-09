package shorten

type CreatingShortLinksDTOIn struct {
	Url string `json:"url"`
}

type CreatingShortLinksDTOOut struct {
	Result string `json:"result"`
}
