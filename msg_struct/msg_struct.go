package msg_struct


import (
    "guard_server/rsa_crypto"
)


type RsaPair struct {
    RemoteRsa *rsa_crypto.RsaCrypto
    LocalRsa *rsa_crypto.RsaCrypto
}


type StartUpReq struct{
    ClientSn string
    SystemSn string
}

