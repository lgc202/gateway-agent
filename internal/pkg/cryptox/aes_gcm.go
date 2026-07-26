// Package cryptox 提供凭证加密所需的基础能力
package cryptox

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

const aes256KeySize = 32

// AESGCM 使用 AES-256-GCM 加密和认证敏感数据
type AESGCM struct {
	aead cipher.AEAD
}

// NewAESGCM 使用 32 字节主密钥创建 AES-256-GCM 加密器
func NewAESGCM(key []byte) (*AESGCM, error) {
	if len(key) != aes256KeySize {
		return nil, fmt.Errorf("AES-256 key must be %d bytes", aes256KeySize)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create AES cipher: %w", err)
	}

	aead, err := cipher.NewGCMWithRandomNonce(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	return &AESGCM{aead: aead}, nil
}

// Encrypt 加密明文，并由标准库生成随机 Nonce 后写入密文头部
func (c *AESGCM) Encrypt(plaintext []byte) []byte {
	return c.aead.Seal(nil, nil, plaintext, nil)
}

// Decrypt 从密文头部读取 Nonce，校验完整性后返回明文
func (c *AESGCM) Decrypt(ciphertext []byte) ([]byte, error) {
	plaintext, err := c.aead.Open(nil, nil, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt ciphertext: %w", err)
	}

	return plaintext, nil
}
