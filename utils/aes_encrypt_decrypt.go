package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
)

// Encrypter 加密接口
type IEncrypt interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
}

type AesEncrypt struct {
	key []byte
}

func NewAesEncrypt(key []byte) *AesEncrypt {
	return &AesEncrypt{key: key}
}

func (a *AesEncrypt) Encrypt(origData []byte) ([]byte, error) {
	key := a.adjustKeyLength()
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	//AES分组长度为128位，所以blockSize=16，单位字节
	blockSize := block.BlockSize()
	origData = pKCS5Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize]) //初始向量的长度必须等于块block的长度16字节
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

func (a *AesEncrypt) Decrypt(encrypted []byte) ([]byte, error) {
	key := a.adjustKeyLength()
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	//AES分组长度为128位，所以blockSize=16，单位字节
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize]) //初始向量的长度必须等于块block的长度16字节
	origData := make([]byte, len(encrypted))
	blockMode.CryptBlocks(origData, encrypted)
	origData = pKCS5UnPadding(origData)
	return origData, nil
}

func (a *AesEncrypt) adjustKeyLength() []byte {
	const keyLength = 16
	if len(a.key) < keyLength {
		// 如果密钥长度不足16个字节，在后面补0字节
		padding := make([]byte, keyLength-len(a.key))
		return append(a.key, padding...)
	} else if len(a.key) > keyLength {
		// 如果密钥长度超过16个字节，只取前16个字节
		return a.key[:keyLength]
	}
	// 密钥长度正好为16个字节，直接返回
	return a.key
}

func pKCS5Padding(plaintext []byte, blockSize int) []byte {
	padding := blockSize - len(plaintext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(plaintext, padtext...)
}

func pKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	if unpadding > length {
		return nil
	}
	return origData[:(length - unpadding)]
}
