package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

// Encrypter 加密接口
type Encrypter interface {
	Encrypt(plaintext []byte) []byte
	Decrypt(ciphertext []byte) []byte
}

type AesEncryptCBC struct {
	key []byte
}

func NewAesEncryptCBC(key []byte) *AesEncryptCBC {
	return &AesEncryptCBC{key: key}
}

func (cbc *AesEncryptCBC) Encrypt(origData []byte) []byte {
	block, _ := aes.NewCipher(cbc.key)
	blockSize := block.BlockSize()                                  // 获取秘钥块的长度
	origData = pkcs5Padding(origData, blockSize)                    // 补全码
	blockMode := cipher.NewCBCEncrypter(block, cbc.key[:blockSize]) // 加密模式
	encrypted := make([]byte, len(origData))                        // 创建数组
	blockMode.CryptBlocks(encrypted, origData)                      // 加密
	return encrypted
}

func (cbc *AesEncryptCBC) Decrypt(encrypted []byte) []byte {
	block, _ := aes.NewCipher(cbc.key)                              // 分组秘钥
	blockSize := block.BlockSize()                                  // 获取秘钥块的长度
	blockMode := cipher.NewCBCDecrypter(block, cbc.key[:blockSize]) // 加密模式
	decrypted := make([]byte, len(encrypted))                       // 创建数组
	blockMode.CryptBlocks(decrypted, encrypted)                     // 解密
	decrypted = pkcs5UnPadding(decrypted)                           // 去除补全码
	return decrypted
}

func AesDecryptCBC(encrypted []byte, key []byte) (decrypted []byte) {
	block, _ := aes.NewCipher(key)                              // 分组秘钥
	blockSize := block.BlockSize()                              // 获取秘钥块的长度
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize]) // 加密模式
	decrypted = make([]byte, len(encrypted))                    // 创建数组
	blockMode.CryptBlocks(decrypted, encrypted)                 // 解密
	decrypted = pkcs5UnPadding(decrypted)                       // 去除补全码
	return decrypted
}
func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padText...)
}

func pkcs5UnPadding(origData []byte) []byte {
	length := len(origData)
	unfading := int(origData[length-1])
	return origData[:(length - unfading)]
}

func AesEncryptECB(origData []byte, key []byte) (encrypted []byte) {
	newCipher, _ := aes.NewCipher(generateKey(key))
	length := (len(origData) + aes.BlockSize) / aes.BlockSize
	plain := make([]byte, length*aes.BlockSize)
	copy(plain, origData)
	pad := byte(len(plain) - len(origData))
	for i := len(origData); i < len(plain); i++ {
		plain[i] = pad
	}
	encrypted = make([]byte, len(plain))
	// 分组分块加密
	for bs, be := 0, newCipher.BlockSize(); bs <= len(origData); bs, be = bs+newCipher.BlockSize(), be+newCipher.BlockSize() {
		newCipher.Encrypt(encrypted[bs:be], plain[bs:be])
	}

	return encrypted
}
func AesDecryptECB(encrypted []byte, key []byte) (decrypted []byte) {
	newCipher, _ := aes.NewCipher(generateKey(key))
	decrypted = make([]byte, len(encrypted))
	//
	for bs, be := 0, newCipher.BlockSize(); bs < len(encrypted); bs, be = bs+newCipher.BlockSize(), be+newCipher.BlockSize() {
		newCipher.Decrypt(decrypted[bs:be], encrypted[bs:be])
	}

	trim := 0
	if len(decrypted) > 0 {
		trim = len(decrypted) - int(decrypted[len(decrypted)-1])
	}

	return decrypted[:trim]
}
func generateKey(key []byte) (genKey []byte) {
	genKey = make([]byte, 16)
	copy(genKey, key)
	for i := 16; i < len(key); {
		for j := 0; j < 16 && i < len(key); j, i = j+1, i+1 {
			genKey[j] ^= key[i]
		}
	}
	return genKey
}

func AesEncryptCFB(origData []byte, key []byte) (encrypted []byte) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	encrypted = make([]byte, aes.BlockSize+len(origData))
	iv := encrypted[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(encrypted[aes.BlockSize:], origData)
	return encrypted
}
func AesDecryptCFB(encrypted []byte, key []byte) (decrypted []byte) {
	block, _ := aes.NewCipher(key)
	if len(encrypted) < aes.BlockSize {
		panic("ciphertext too short")
	}
	iv := encrypted[:aes.BlockSize]
	encrypted = encrypted[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(encrypted, encrypted)
	return encrypted
}
