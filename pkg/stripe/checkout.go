package stripe

import (
	"context"

	stripe "github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/client"
)

const MetadataKeyMarker = "authgear_once_license_server"

type CheckoutSessionParams struct {
	MarkerValue string
	SuccessURL  string
	CancelURL   string
	PriceID     string
}

func NewCheckoutSession(ctx context.Context, client *client.API, params *CheckoutSessionParams) (*stripe.CheckoutSession, error) {
	sessParams := &stripe.CheckoutSessionParams{
		SuccessURL: stripe.String(params.SuccessURL),
		CancelURL:  stripe.String(params.CancelURL),
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		// Always create a customer so that we can get a customer_id back.
		CustomerCreation: stripe.String(string(stripe.CheckoutSessionCustomerCreationAlways)),
		// Allow the use of promotion codes.
		// https://linear.app/authgear/issue/DEV-2756/allow-promo-codes-coupon-to-be-used-in-once
		AllowPromotionCodes: stripe.Bool(true),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(params.PriceID),
				Quantity: stripe.Int64(1),
			},
		},
		Metadata: map[string]string{
			MetadataKeyMarker: params.MarkerValue,
		},
	}

	sess, err := client.CheckoutSessions.New(sessParams)
	if err != nil {
		return nil, err
	}

	return sess, nil
}
