package utils

import "encoding/hex"

// PasswordManager 密码管理器
type PasswordManager struct {
	Key       []byte
	Encrypter Encrypter
}

func NewPasswordManager(key []byte, encrypter Encrypter) *PasswordManager {
	return &PasswordManager{
		Key:       key,
		Encrypter: encrypter,
	}
}

func (manager *PasswordManager) EncryptPassword(password string) string {
	encrypted := manager.Encrypter.Encrypt([]byte(password))
	return hex.EncodeToString(encrypted)
}

func (manager *PasswordManager) DecryptPassword(encryptedPassword string) (string, error) {
	encrypted, err := hex.DecodeString(encryptedPassword)
	if err != nil {
		return "", err
	}
	return string(manager.Encrypter.Decrypt(encrypted)), nil
}

// ValidatePassword 校验密码 返回是否正确
//
// password:明文
//
// encryptedPassword:密文
func (manager *PasswordManager) ValidatePassword(password, encryptedPassword string) bool {
	dePassword, err := manager.DecryptPassword(encryptedPassword)
	if err != nil {
		return false
	}
	return dePassword == password
}
