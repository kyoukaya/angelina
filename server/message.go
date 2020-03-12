package server

type messageT struct {
	client  *Client
	payload []byte
}
