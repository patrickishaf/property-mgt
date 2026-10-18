package bachs

import "errors"

type Client struct{}

func CreateClient() *Client {
	return &Client{}
}

func (this *Client) InitiateCheckout(dto createCheckoutSessionBody) (*checkoutRes, error) {
	return nil, errors.New("")
}

func (this *Client) GetCheckout() (*checkoutRes, error) {
	return nil, errors.New("")
}
