package utils

import "encoding/hex"

// PasswordManager 密码管理器
type PasswordManager struct {
	encrypt IEncrypt
}

func NewPasswordManager(encrypt IEncrypt) *PasswordManager {
	return &PasswordManager{
		encrypt: encrypt,
	}
}

func (manager *PasswordManager) EncryptPassword(password string) string {
	encrypted, err := manager.encrypt.Encrypt([]byte(password))
	if err != nil {
		return ""
	}
	return hex.EncodeToString(encrypted)
}

func (manager *PasswordManager) DecryptPassword(encryptedPassword string) (string, error) {
	encrypted, err := hex.DecodeString(encryptedPassword)
	if err != nil {
		return "", err
	}
	password, err := manager.encrypt.Decrypt(encrypted)
	if err != nil {
		return "", err
	}
	return string(password), nil
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
