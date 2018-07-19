package message

// Envelope encapsulate message and its data
type Envelope struct {
	// t   int
	ID       int    `json:"id"`
	Username string `json:"username"` // TODO: the sender username or id
	Msg      string `json:"message"`
}
