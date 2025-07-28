package shorten

type CreatingShortLinksDTOIn struct {
	URL string `json:"url"`
}

type CreatingShortLinksDTOOut struct {
	Result string `json:"result"`
}
