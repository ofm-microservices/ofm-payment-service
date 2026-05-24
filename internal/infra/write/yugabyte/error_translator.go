package yugabyte

import "fmt"

type dbErrorTranslator struct{}

// NewDBErrorTranslator constructs the payment write-model error translator.
func NewDBErrorTranslator() DBErrorTranslator { return &dbErrorTranslator{} }

func (t *dbErrorTranslator) TranslateCreatePaymentIntentError(err error) error {
	return fmt.Errorf("create payment intent: %w", err)
}
func (t *dbErrorTranslator) TranslateFindPaymentIntentError(err error) error {
	return fmt.Errorf("find payment intent: %w", err)
}
func (t *dbErrorTranslator) TranslateUpdatePaymentIntentError(err error) error {
	return fmt.Errorf("update payment intent: %w", err)
}
func (t *dbErrorTranslator) TranslateCreatePaymentReleaseError(err error) error {
	return fmt.Errorf("create payment release: %w", err)
}
func (t *dbErrorTranslator) TranslateFindPaymentReleaseError(err error) error {
	return fmt.Errorf("find payment release: %w", err)
}
func (t *dbErrorTranslator) TranslateUpdatePaymentReleaseError(err error) error {
	return fmt.Errorf("update payment release: %w", err)
}
