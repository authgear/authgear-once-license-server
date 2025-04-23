package stripe

import (
	"errors"
	"io"
	"net/http"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
)

func ConstructEvent(r *http.Request, signingSecret string) (*stripe.Event, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	sig := r.Header.Get("Stripe-Signature")
	e, err := webhook.ConstructEvent(body, sig, signingSecret)
	if err != nil {
		return nil, err
	}

	return &e, nil
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

func GetCustomerEmail(e *stripe.Event) (string, bool) {
	// {
	//   "account": "",
	//   "api_version": "2025-03-31.basil",
	//   "created": 1745401714,
	//   "data": {
	//     "previous_attributes": null,
	//     "object": {
	//       "id": "cs_test_a1RXNfa1BzgWY8zPhlnPtzQg1KtlyIaE3GwXVK3eALVCItsAoIYEgFA8z9",
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
	//       "cancel_url": "https://authgear.com",
	//       "client_reference_id": null,
	//       "client_secret": null,
	//       "collected_information": null,
	//       "consent": null,
	//       "consent_collection": null,
	//       "created": 1745401699,
	//       "currency": "usd",
	//       "currency_conversion": null,
	//       "custom_fields": [],
	//       "custom_text": {
	//         "after_submit": null,
	//         "shipping_address": null,
	//         "submit": null,
	//         "terms_of_service_acceptance": null
	//       },
	//       "customer": null,
	//       "customer_creation": "if_required",
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
	//       "expires_at": 1745488099,
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
	//       "payment_intent": "pi_3RGzuvBv9FIDZu7y1syjAnvK",
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
	//       "saved_payment_method_options": null,
	//       "setup_intent": null,
	//       "shipping_address_collection": null,
	//       "shipping_options": [],
	//       "status": "complete",
	//       "submit_type": null,
	//       "subscription": null,
	//       "success_url": "https://authgear.com",
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
	//   "id": "evt_1RGzuwBv9FIDZu7yKp4kzCho",
	//   "livemode": false,
	//   "object": "event",
	//   "pending_webhooks": 2,
	//   "request": {
	//     "id": "",
	//     "idempotency_key": ""
	//   },
	//   "type": "checkout.session.completed"
	// }
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

func GetCheckoutSessionID(e *stripe.Event) (string, bool) {
	id, ok := e.Data.Object["id"].(string)
	if !ok {
		return "", false
	}
	return id, true
}
