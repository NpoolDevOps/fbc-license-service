package types

import (
	"github.com/google/uuid"
	"time"
)

type ExchangeKeyInput struct {
	Spec      string `json:"spec"`
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
	ClientUser   string `json:"client_user"`
	ClientPasswd string `json:"client_passwd"`
	ClientSN     string `json:"client_sn"`
	NetworkType  string `json:"network_type"`
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

type HeartbeatV1Output HeartbeatOutput

type MyClientsInput struct {
	AuthCode string `json:"auth_code"`
}

type ClientInfo struct {
	Id          uuid.UUID `gorm:"column:id;primary_key" json:"id"`
	ClientUser  string    `gorm:"column:client_user" json:"client_user"`
	ClientSn    string    `gorm:"column:client_sn" json:"client_sn"`
	Status      string    `gorm:"column:status" json:"status"`
	CreateTime  time.Time `gorm:"column:create_time" json:"create_time"`
	ModifyTime  time.Time `gorm:"column:modify_time" json:"modify_time"`
	NetworkType string    `gorm:"-" json:"network_type"`
}

type UserInfo struct {
	Id           uuid.UUID `gorm:"column:id;primary_key" json:"id"`
	Username     string    `gorm:"column:username" json:"username"`
	ValidateDate time.Time `gorm:"column:validate_date" json:"validate_date"`
	Quota        int       `gorm:"column:quota" json:"quota"`
	Count        int       `gorm:"column:count" json:"count"`
	CreateTime   time.Time `gorm:"column:create_time" json:"create_time"`
	ModifyTime   time.Time `gorm:"column:modify_time" json:"modify_time"`
}

type MyClientsOutput struct {
	SuperUser   bool         `json:"super_user"`
	VisitorOnly bool         `json:"visitor_only"`
	Users       []UserInfo   `json:"users"`
	Clients     []ClientInfo `json:"clients"`
}

type UpdateAuthInput struct {
	AuthCode     string `json:"auth_code"`
	Username     string `json:"username"`
	Quota        int    `json:"quota"`
	ValidateDate int    `json:"validate_time"`
}

type ClientInfoInput struct {
	Id uuid.UUID `json:"id"`
}

type ClientInfoOutput = ClientInfo
