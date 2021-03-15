package types

import (
	"github.com/google/uuid"
)

type StartUpReq struct {
	ClientSn string
	SystemSn string
}

type ExchangeKeyInput struct {
	PublicKey string `json:"public_key"`
}

type ExchangeKeyOutput struct {
	SessionId uuid.UUID `json:"session_id"`
	PublicKey string    `json:"public_key"`
}
