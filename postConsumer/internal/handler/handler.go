package handler

import (
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"log/slog"
)

type Handler struct {
	log *slog.Logger
}

func NewHandler(log slog.Logger) *Handler {
	return &Handler{
		log: &log,
	}
}

func (h *Handler) HandleMessage(message []byte, offset kafka.Offset) error {
	const op = "handler.HandleMessage"
	h.log.With("op", op).Info("Message from kafka with offset "+offset.String(), slog.String("message", string(message)))
	return nil
}
