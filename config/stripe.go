package config

// StripeConfig defines Stripe webhook and Connect onboarding settings.
type StripeConfig struct {
	SecretKey                         string `env:"STRIPE_SECRET_KEY,required"`
	CheckoutWebhookSecret             string `env:"STRIPE_PAYMENT_WEBHOOK_SECRET,required"`
	FreelancerOnboardingWebhookSecret string `env:"STRIPE_FREELANCER_ONBOARDING_WEBHOOK_SECRET,required"`
	ConnectReturnURL                  string `env:"STRIPE_CONNECT_RETURN_URL,required"`
	ConnectRefreshURL                 string `env:"STRIPE_CONNECT_REFRESH_URL,required"`
	ConnectCountry                    string `env:"STRIPE_CONNECT_COUNTRY" envDefault:"US"`
	SkipTransfers                     bool   `env:"STRIPE_SKIP_TRANSFERS" envDefault:"false"`
}
