package domain

type GMessagePart struct {
	Text string `json:"text"`
}

type GMessage struct {
	Role  string         `json:"role"`
	Parts []GMessagePart `json:"parts"`
}
