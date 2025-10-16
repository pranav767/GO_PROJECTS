package payment

import (
	"os"

	"github.com/stripe/stripe-go/v76"
)

func InitStripe() {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
}
