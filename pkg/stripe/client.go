package stripe

import (
	"github.com/stripe/stripe-go/v82/client"
)

func NewClient(secretKey string) *client.API {
	return client.New(secretKey, nil)
}
