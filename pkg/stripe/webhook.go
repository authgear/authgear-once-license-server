package stripe

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/client"
	"github.com/stripe/stripe-go/v82/webhook"
)

var ErrUnknownEvent = errors.New("pkgstripe: unknown event")

type ConstructEventOptions struct {
	SigningSecret string
	MarkerValue   string
}

func ConstructEvent(ctx context.Context, client *client.API, r *http.Request, opts ConstructEventOptions) (*stripe.Event, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	sig := r.Header.Get("Stripe-Signature")
	e, err := webhook.ConstructEvent(body, sig, opts.SigningSecret)
	if err != nil {
		return nil, err
	}

	if e.Type != stripe.EventTypeCheckoutSessionCompleted {
		return &e, ErrUnknownEvent
	}

	checkoutSessionID := GetEventDataID(&e)
	checkoutSession, err := client.CheckoutSessions.Get(checkoutSessionID, &stripe.CheckoutSessionParams{
		Expand: []*string{
			stripe.String("line_items"),
		},
	})
	if err != nil {
		return nil, err
	}

	marker := checkoutSession.Metadata[MetadataKeyMarker]
	if marker == opts.MarkerValue {
		return &e, nil
	}

	return &e, ErrUnknownEvent
}

func IsWebhookClientError(err error) bool {
	switch {
	case errors.Is(err, webhook.ErrInvalidHeader):
		return true
	case errors.Is(err, webhook.ErrNoValidSignature):
		return true
	case errors.Is(err, webhook.ErrNotSigned):
		return true
	case errors.Is(err, webhook.ErrTooOld):
		return true
	default:
		return false
	}
}

// Example of the event.
// {
//   "account": "",
//   "api_version": "2025-03-31.basil",
//   "created": 1745579518,
//   "data": {
//     "previous_attributes": null,
//     "object": {
//       "id": "cs_test_a11hRgOSk6QtPrKyWrprW1DdsK8wtVH1NOZ39OQsxsZQhfJhtqGxo2QvMG",
//       "object": "checkout.session",
//       "adaptive_pricing": {
//         "enabled": true
//       },
//       "after_expiration": null,
//       "allow_promotion_codes": null,
//       "amount_subtotal": 29900,
//       "amount_total": 29900,
//       "automatic_tax": {
//         "enabled": false,
//         "liability": null,
//         "provider": null,
//         "status": null
//       },
//       "billing_address_collection": null,
//       "cancel_url": "https://www.authgear.com/payment-unsuccessful",
//       "client_reference_id": null,
//       "client_secret": null,
//       "collected_information": {
//         "shipping_details": null
//       },
//       "consent": null,
//       "consent_collection": null,
//       "created": 1745579492,
//       "currency": "usd",
//       "currency_conversion": null,
//       "custom_fields": [],
//       "custom_text": {
//         "after_submit": null,
//         "shipping_address": null,
//         "submit": null,
//         "terms_of_service_acceptance": null
//       },
//       "customer": "cus_SC8R9AbEdGa1gO",
//       "customer_creation": "always",
//       "customer_details": {
//         "address": {
//           "city": null,
//           "country": "HK",
//           "line1": null,
//           "line2": null,
//           "postal_code": null,
//           "state": null
//         },
//         "email": "louischan@oursky.com",
//         "name": "Louis Chan",
//         "phone": null,
//         "tax_exempt": "none",
//         "tax_ids": []
//       },
//       "customer_email": null,
//       "discounts": [],
//       "expires_at": 1745665892,
//       "invoice": null,
//       "invoice_creation": {
//         "enabled": false,
//         "invoice_data": {
//           "account_tax_ids": null,
//           "custom_fields": null,
//           "description": null,
//           "footer": null,
//           "issuer": null,
//           "metadata": {},
//           "rendering_options": null
//         }
//       },
//       "livemode": false,
//       "locale": null,
//       "metadata": {},
//       "mode": "payment",
//       "payment_intent": "pi_3RHkAYBv9FIDZu7y0bUK7o8S",
//       "payment_link": null,
//       "payment_method_collection": "if_required",
//       "payment_method_configuration_details": {
//         "id": "pmc_1Om4HzBv9FIDZu7yqvJpTGJf",
//         "parent": null
//       },
//       "payment_method_options": {
//         "card": {
//           "request_three_d_secure": "automatic"
//         }
//       },
//       "payment_method_types": [
//         "card",
//         "link"
//       ],
//       "payment_status": "paid",
//       "permissions": null,
//       "phone_number_collection": {
//         "enabled": false
//       },
//       "recovered_from": null,
//       "saved_payment_method_options": {
//         "allow_redisplay_filters": [
//           "always"
//         ],
//         "payment_method_remove": null,
//         "payment_method_save": null
//       },
//       "setup_intent": null,
//       "shipping_address_collection": null,
//       "shipping_options": [],
//       "status": "complete",
//       "submit_type": null,
//       "subscription": null,
//       "success_url": "https://www.authgear.com/payment-confirmed",
//       "total_details": {
//         "amount_discount": 0,
//         "amount_shipping": 0,
//         "amount_tax": 0
//       },
//       "ui_mode": "hosted",
//       "url": null,
//       "wallet_options": null,
//       "shipping_cost": null
//     }
//   },
//   "id": "evt_1RHkAmBv9FIDZu7y5SYx6Xu3",
//   "livemode": false,
//   "object": "event",
//   "pending_webhooks": 2,
//   "request": {
//     "id": "",
//     "idempotency_key": ""
//   },
//   "type": "checkout.session.completed"
// }

func GetCustomerID(e *stripe.Event) (string, bool) {
	id, ok := e.Data.Object["customer"].(string)
	if !ok {
		return "", false
	}
	return id, true
}

func GetCustomerEmail(e *stripe.Event) (string, bool) {
	customer_details, ok := e.Data.Object["customer_details"].(map[string]interface{})
	if !ok {
		return "", false
	}
	email, ok := customer_details["email"].(string)
	if !ok {
		return "", false
	}
	return email, true
}

func GetEventDataID(e *stripe.Event) string {
	id := e.Data.Object["id"].(string)
	if id == "" {
		panic(fmt.Errorf("stripe event data has no ID"))
	}
	return id
}
