package mycrypt

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/rombintu/goyametricsv2/internal/logger"
	"github.com/rombintu/goyametricsv2/lib/common"
	"go.uber.org/zap"
)

func ValidPrivateKey(filePath string) bool {
	if !common.FileIsExists(filePath) {
		logger.Log.Debug("file not exists", zap.String("file", filePath))
		return false
	}

	// Читаем содержимое файла
	keyBytes, err := os.ReadFile(filePath)
	if err != nil {
		logger.Log.Debug("error read file", zap.String("file", filePath))
		return false
	}

	// Декодируем PEM-блок
	block, _ := pem.Decode(keyBytes)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		logger.Log.Debug("error decode file", zap.String("file", filePath))
		return false
	}

	// Парсим приватный ключ
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		logger.Log.Debug("error parse private key", zap.String("file", filePath))
		return false
	}

	// Проверяем длину ключа
	if privateKey.N.BitLen() != 4096 {
		logger.Log.Debug("error private key len != 4096", zap.String("file", filePath))
		return false
	}

	// Проверка валидности ключа
	if err := privateKey.Validate(); err != nil {
		logger.Log.Debug("invalid RSA private key", zap.String("file", filePath))
		return false
	}

	return true
}

func GenRSAKeyPair(filename string) (*rsa.PrivateKey, error) {
	// Генерация приватного ключа
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	// Публичный ключ из приватного
	publicKey := &privateKey.PublicKey

	if err := SavePrivateKey(filename, privateKey); err != nil {
		return nil, err
	}
	if err := SavePublicKey(filename+".pub", publicKey); err != nil {
		return nil, err
	}

	return privateKey, nil
}

func SavePrivateKey(filename string, key *rsa.PrivateKey) error {
	// Сериализация приватного ключа в PKCS#1, ASN.1 DER формат
	keyBytes := x509.MarshalPKCS1PrivateKey(key)

	// Создание PEM блока
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	}

	// Сохранение в файл
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return pem.Encode(file, block)
}

func SavePublicKey(filename string, key *rsa.PublicKey) error {
	// Сериализация публичного ключа в PKIX, ASN.1 DER формат
	keyBytes, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return err
	}

	// Создание PEM блока
	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: keyBytes,
	}

	// Сохранение в файл
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return pem.Encode(file, block)
}

func LoadPrivateKey(filename string) (*rsa.PrivateKey, error) {
	privateKeyBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(privateKeyBytes)
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func LoadPublicKey(filename string) (*rsa.PublicKey, error) {
	publicKeyBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(publicKeyBytes)
	publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return publicKey, nil
}

func EncryptMiddleware(privateKey *rsa.PrivateKey) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Дешифрование данных из запроса
			if c.Request().Method == http.MethodPost || c.Request().Method == http.MethodPut {
				body, err := io.ReadAll(c.Request().Body)
				if err != nil {
					return err
				}
				decryptedBytes, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, body)
				if err != nil {
					return err
				}
				c.Request().Body = io.NopCloser(bytes.NewReader(decryptedBytes))
			}

			// Передача управления следующему обработчику
			if err := next(c); err != nil {
				c.Error(err)
			}

			return nil
		}
	}
}

func EncryptWithPublicKey(pubKey *rsa.PublicKey, buff *bytes.Buffer) error {
	// Чтение содержимого буфера
	plaintext := buff.Bytes()

	// Шифрование содержимого с помощью публичного ключа
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, plaintext)
	if err != nil {
		return err
	}

	// Очистка буфера и запись зашифрованного сообщения в буфер
	buff.Reset()
	_, err = buff.Write(ciphertext)
	if err != nil {
		return err
	}

	return nil
}
