package mycrypt

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"os"
	"time"

	"github.com/rombintu/goyametricsv2/internal/logger"
	"github.com/rombintu/goyametricsv2/lib/common"
)

func ValidPrivateKey(filePath string) bool {
	if !common.FileIsExists(filePath) {
		logger.Log.Error("file not exists")
		return false
	}

	// Читаем содержимое файла
	keyBytes, err := os.ReadFile(filePath)
	if err != nil {
		logger.Log.Error("error read file")
		return false
	}

	// Декодируем PEM-блок
	block, _ := pem.Decode(keyBytes)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		logger.Log.Error("error decode file")
		return false
	}

	// Парсим приватный ключ
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		logger.Log.Error("error parse private key")
		return false
	}

	// Проверяем длину ключа
	if privateKey.N.BitLen() != 4096 {
		logger.Log.Error("error private key len != 4096")
		return false
	}

	// Проверка валидности ключа
	if err := privateKey.Validate(); err != nil {
		logger.Log.Error("invalid RSA private key")
		return false
	}

	return true
}

func GenPrivKeyAndCertPEM(filePath string) error {
	cert := &x509.Certificate{
		// указываем уникальный номер сертификата
		SerialNumber: big.NewInt(1658),
		// заполняем базовую информацию о владельце сертификата
		Subject: pkix.Name{
			Organization: []string{"Yandex.Praktikum"},
			Country:      []string{"RU"},
		},

		IPAddresses: []net.IP{net.IPv4(0, 0, 0, 0), net.IPv6loopback},
		// сертификат верен, начиная со времени создания
		NotBefore: time.Now(),
		// время жизни сертификата — 10 лет
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		// устанавливаем использование ключа для цифровой подписи,
		// а также клиентской и серверной авторизации
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		logger.Log.Error("error gen private key")
		return err
	}

	// создаём сертификат x.509
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		log.Fatal(err)
	}

	// кодируем сертификат и ключ в формате PEM, который
	// используется для хранения и обмена криптографическими ключами
	var certPEM bytes.Buffer
	pem.Encode(&certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	var privateKeyPEM bytes.Buffer
	pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err := common.ReWriteFile(filePath, privateKeyPEM.Bytes()); err != nil {
		logger.Log.Error("error rewrite private key file")
		return err
	}
	if err := common.ReWriteFile(filePath+".cert", certPEM.Bytes()); err != nil {
		logger.Log.Error("error rewrite certificate file")
		return err
	}
	return nil
}
