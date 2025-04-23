package stripe

import (
	"context"

	stripe "github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/client"
)

type CheckoutSessionParams struct {
	SuccessURL string
	CancelURL  string
	PriceID    string
}

func NewCheckoutSession(ctx context.Context, client *client.API, params *CheckoutSessionParams) (*stripe.CheckoutSession, error) {
	sessParams := &stripe.CheckoutSessionParams{
		SuccessURL: stripe.String(params.SuccessURL),
		CancelURL:  stripe.String(params.CancelURL),
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(params.PriceID),
				Quantity: stripe.Int64(1),
			},
		},
	}

	sess, err := client.CheckoutSessions.New(sessParams)
	if err != nil {
		return nil, err
	}

	return sess, nil
}
