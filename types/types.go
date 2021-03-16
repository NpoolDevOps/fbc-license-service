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

type CommonInput struct {
	SessionId uuid.UUID `json:"session_id"`
}

type ClientLoginInput struct {
	CommonInput
	ClientUser string `json:"client_user"`
	ClientSN   string `json:"client_sn"`
}

type ClientLoginOutput struct {
	ClientUuid uuid.UUID `json:"client_uuid"`
}

type HeartbeatInput struct {
	CommonInput
	ClientUuid uuid.UUID `json:"client_uuid"`
}

type HeartbeatOutput struct {
	ShouldStop bool `json:"should_stop"`
}
