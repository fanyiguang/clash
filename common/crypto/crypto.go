package crypto

import (
	"encoding/base64"
	"time"

	"github.com/forgoer/openssl"
)

// AecEcb128Pkcs7EncryptWithDefaultKey 默认key是 `buzzword` + `日期`，如 `buzzword2020-01-01`
func AecEcb128Pkcs7EncryptWithDefaultKey(plaintext []byte) (string, error) {
	defaultKeyPrefix := "buzzword"

	key := make([]byte, 0, 16)
	key = append(key, defaultKeyPrefix...)
	key = append(key, []byte(time.Now().Format("20060102"))...)

	encryptData, err := openssl.AesECBEncrypt(plaintext, key, openssl.PKCS7_PADDING)
	if err != nil {
		return "", err
	}
	base64Data := base64.StdEncoding.EncodeToString(encryptData)

	return base64Data, nil
}

func AecEcb128Pkcs7Encrypt(plaintext []byte, key []byte) ([]byte, error) {
	return openssl.AesECBEncrypt(plaintext, key, openssl.PKCS7_PADDING)
}
