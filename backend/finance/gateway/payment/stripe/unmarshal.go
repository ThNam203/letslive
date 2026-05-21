package stripe

import (
	"encoding/json"
	"fmt"

	stripego "github.com/stripe/stripe-go/v82"
)

// unmarshalDataObject decodes event.Data.Raw into the typed Stripe object so
// downstream code does not have to deal with the generic map.
func unmarshalDataObject(event stripego.Event, dest any) error {
	if err := json.Unmarshal(event.Data.Raw, dest); err != nil {
		return fmt.Errorf("stripe webhook payload: %w", err)
	}
	return nil
}
