package eventbroker

import app "payment-service/internal/application"

// MessageHandler is the application-layer broker handler contract.
type MessageHandler = app.MessageHandler

// EventBroker is the application-layer broker contract.
type EventBroker = app.EventBroker
