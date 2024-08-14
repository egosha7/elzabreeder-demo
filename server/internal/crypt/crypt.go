// Данный пакет не реализован в этом проекте

package crypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
)

// encrypt используется для шифрования данных
func encrypt(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Создаем генератор случайных чисел
	iv := make([]byte, aes.BlockSize)
	_, err = rand.Read(iv)
	if err != nil {
		return nil, err
	}

	// Реализуем шифрование данных
	mode := cipher.NewCBCEncrypter(block, iv)
	paddedData := pkcs7Padding(data, aes.BlockSize)
	encryptedData := make([]byte, len(paddedData))
	mode.CryptBlocks(encryptedData, paddedData)

	// Возвращаем зашифрованные данные с вектором инициализации
	return append(iv, encryptedData...), nil
}

// decrypt используется для дешифрования данных
func decrypt(encryptedData []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Извлекаем вектор инициализации из зашифрованных данных
	iv := encryptedData[:aes.BlockSize]
	encryptedData = encryptedData[aes.BlockSize:]

	// Реализуем дешифрование данных
	mode := cipher.NewCBCDecrypter(block, iv)
	decryptedData := make([]byte, len(encryptedData))
	mode.CryptBlocks(decryptedData, encryptedData)

	// Убираем дополнительные байты padding'а
	return pkcs7Unpadding(decryptedData), nil
}

// pkcs7Padding используется для дополнения данных до размера блока
func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// pkcs7Unpadding используется для удаления padding'а из данных
func pkcs7Unpadding(data []byte) []byte {
	padding := int(data[len(data)-1])
	return data[:len(data)-padding]
}
