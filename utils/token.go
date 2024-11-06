package utils

import (
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

type Token struct {
	ExpiredAt int64
	Signature string // 签名
}

func NewToken(userId string, expiredAt int64) *Token {
	return &Token{ExpiredAt: expiredAt, Signature: userId + "\000" + strconv.FormatInt(expiredAt, 10)}
}

func (token *Token) Sign(key []byte) string {
	aesEncrypt := NewAesEncrypt(key)
	encryptedToken, err := aesEncrypt.Encrypt([]byte(token.Signature))
	if err != nil {
		return ""
	}
	return hex.EncodeToString(encryptedToken)
}

func ValidateToken(tokenStr string, key []byte) (bool, string) {
	aesEncrypt := NewAesEncrypt(key)
	decryptedToken, err := hex.DecodeString(tokenStr)
	if err != nil {
		return false, ""
	}
	tokenBuf, err := aesEncrypt.Decrypt(decryptedToken)
	if err != nil {
		return false, ""
	}
	signatureData := strings.Split(string(tokenBuf), "\000")
	if len(signatureData) != 2 {
		return false, ""
	}
	expiredAt, err := strconv.ParseInt(signatureData[1], 10, 64)
	if err != nil {
		return false, ""
	}
	if expiredAt < time.Now().Unix() {
		return false, ""
	}
	return true, signatureData[0]
}
