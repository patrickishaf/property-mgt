package cpaas

type ticker struct {
	client *Client
}

func NewTicker(client *Client) *ticker {
	return &ticker{
		client: client,
	}
}

func (this *ticker) Run() {
	/*
	 */
}
