package stripe

import (
	"context"

	stripe "github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/client"
)

func GetCustomer(ctx context.Context, client *client.API, customerID string) (*stripe.Customer, error) {
	customer, err := client.Customers.Get(customerID, &stripe.CustomerParams{})
	if err != nil {
		return nil, err
	}
	return customer, nil
}
