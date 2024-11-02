package utils_test

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/cylScripter/chest/utils"
	"log"
	"testing"
)

func TestAesEncrypt(t *testing.T) {
	origData := []byte("Hello World") // 待加密的数据
	key := []byte("ABCDEFGHIJKLMNOP") // 加密的密钥
	log.Println("原文：", string(origData))

	passwordManager := utils.NewPasswordManager(key, utils.NewAesEncryptCBC(key))

	log.Println("------------------ CBC模式 --------------------")

	fmt.Println(passwordManager.EncryptPassword(string(origData)))
	dePass, err := passwordManager.DecryptPassword(passwordManager.EncryptPassword(string(origData)))
	fmt.Println(dePass, "err:", err)
	fmt.Println(passwordManager.ValidatePassword("jjkkii", passwordManager.EncryptPassword(string(origData))))

	log.Println("------------------ ECB模式 --------------------")

	encrypted := utils.AesEncryptECB(origData, key)
	log.Println("密文(hex)：", hex.EncodeToString(encrypted))
	log.Println("密文(base64)：", base64.StdEncoding.EncodeToString(encrypted))
	decrypted := utils.AesDecryptECB(encrypted, key)
	log.Println("解密结果：", string(decrypted))

	log.Println("------------------ CFB模式 --------------------")
	encrypted = utils.AesEncryptCFB(origData, key)
	log.Println("密文(hex)：", hex.EncodeToString(encrypted))
	log.Println("密文(base64)：", base64.StdEncoding.EncodeToString(encrypted))
	decrypted = utils.AesDecryptCFB(encrypted, key)
	log.Println("解密结果：", string(decrypted))
}
