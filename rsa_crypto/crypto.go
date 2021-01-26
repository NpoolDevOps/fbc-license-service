package rsa_crypto


import (
    // "crypto"
    "crypto/rand"
    "crypto/rsa"
    // "crypto/sha256"
    "crypto/x509"
    // "encoding/hex"
    "encoding/pem"
    "errors"
)


type RsaCrypto struct {
    Pubkey  []byte
    Privkey []byte
    Keylen  int
}


func GenerateRsakey(keylen int) ([]byte, []byte){
    // 生成私钥文件
    privateKey, err := rsa.GenerateKey(rand.Reader, keylen)
    if err != nil {
        panic(err)
    }
    derStream := x509.MarshalPKCS1PrivateKey(privateKey)
    block := &pem.Block{
        Type:  "RSA PRIVATE KEY",
        Bytes: derStream,
    }
    prvkey := pem.EncodeToMemory(block)
    publicKey := &privateKey.PublicKey
    derPkix, err := x509.MarshalPKIXPublicKey(publicKey)
    if err != nil {
        panic(err)
    }
    block = &pem.Block{
        Type:  "PUBLIC KEY",
        Bytes: derPkix,
    }
    pubkey := pem.EncodeToMemory(block)
    
    return pubkey, prvkey
}


func NewRsaCrypto(keylen int) *RsaCrypto{

    pubkey, prvkey := GenerateRsakey(keylen)

    return &RsaCrypto{
        Pubkey:  pubkey,
        Privkey: prvkey,
        Keylen:  keylen,
    }
}


func NewRsaCryptoWithParam(pubkey []byte, privkey []byte) *RsaCrypto{

    return &RsaCrypto{
        Pubkey: pubkey,
        Privkey: privkey,
    }
}


func (self *RsaCrypto) Encrypt(content []byte) ([]byte, error){
    //解密pem格式的公钥
    block, _ := pem.Decode(self.Pubkey)
    if block == nil {
        // panic(errors.New("public key error"))
        return nil, errors.New("public key error")
    }
    // 解析公钥
    pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
    if err != nil {
        // panic(err)
        return nil, err
    }
    // 类型断言
    pub := pubInterface.(*rsa.PublicKey)
    //加密
    ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, pub, content)
    if err != nil {
        // panic(err)
        return nil, err
    }

    return ciphertext, nil
}


func (self *RsaCrypto) Decrypt(ciphertext []byte) ([]byte, error){

    //获取私钥
    block, _ := pem.Decode(self.Privkey)
    if block == nil {
        //panic(errors.New("private key error!"))
        return nil, errors.New("private key error")
    }
    //解析PKCS1格式的私钥
    priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
    if err != nil {
        // panic(err)
        return nil, err
    }
    // 解密
    data, err := rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext)
    if err != nil {
        // panic(err)
        return nil, err
    }
    return data, nil
}


func (self *RsaCrypto) GetPubkey() []byte{

    return self.Pubkey
}


func (self *RsaCrypto) GetPrivkey() []byte{

    return self.Privkey
}

