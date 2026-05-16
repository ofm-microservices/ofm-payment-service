package yugabyte

// DBErrorTranslator maps storage-driver failures into domain-aware repository
// errors.
type DBErrorTranslator interface {
	TranslateCreatePaymentIntentError(err error) error
	TranslateFindPaymentIntentError(err error) error
	TranslateUpdatePaymentIntentError(err error) error
}
