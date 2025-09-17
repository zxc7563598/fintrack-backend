package middleware

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/zxc7563598/fintrack-backend/utils/response"

	"github.com/gin-gonic/gin"
)

var privateKey *rsa.PrivateKey

type DecryptRequest struct {
	Timestamp  int64  `json:"timestamp"`
	Sign       string `json:"sign"`
	EnData     string `json:"en_data"`
	EncPayload string `json:"enc_payload"`
}

// 初始化加载私钥
func InitRSAKey(path string) error {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	block, _ := pem.Decode(keyData)
	if block == nil {
		return errors.New("未能解析 PEM 文件")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return err
	}
	var ok bool
	privateKey, ok = key.(*rsa.PrivateKey)
	if !ok {
		return errors.New("不是合法的RSA私钥")
	}
	return nil
}

// 解密中间件
func DecryptMiddleware[T any]() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取请求数据
		var req DecryptRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, 300003)
			c.Abort()
			return
		}
		timestamp := req.Timestamp
		sign := req.Sign
		enData := req.EnData
		encPayload := req.EncPayload
		// 验证时间戳 ±1分钟
		if time.Now().Unix()-timestamp > 60 {
			response.Fail(c, 300004)
			c.Abort()
			return
		}
		// RSA 使用 RSA-OAEP-SHA1 解密 AESKEY + AESIV
		encBytes, _ := base64.StdEncoding.DecodeString(encPayload)
		aescBytes, err := rsa.DecryptOAEP(sha1.New(), rand.Reader, privateKey, encBytes, nil)
		if err != nil || len(aescBytes) < 48 {
			response.Fail(c, 300005)
			c.Abort()
			return
		}
		// 前 24 字节是 AESKEY_base64，后 24 字节是 AESIV_base64
		aesKeyBase64 := aescBytes[:24]
		aesIVBase64 := aescBytes[len(aescBytes)-24:]
		// AESKEY + AESIV 进行 Base64 解密
		aesKey, err := base64.StdEncoding.DecodeString(string(aesKeyBase64))
		if err != nil || len(aesKey) != 16 {
			response.Fail(c, 300006)
			c.Abort()
			return
		}
		aesIV, err := base64.StdEncoding.DecodeString(string(aesIVBase64))
		if err != nil || len(aesIV) != 16 {
			response.Fail(c, 300007)
			c.Abort()
			return
		}
		// 验证签名 MD5(AESKEY + AESIV + timestamp)
		appKey := append(aesKeyBase64, aesIVBase64...)
		firstMD5 := md5.Sum(appKey)
		firstMD5Hex := hex.EncodeToString(firstMD5[:])
		secondMD5 := md5.Sum([]byte(firstMD5Hex + strconv.FormatInt(timestamp, 10)))
		finalMD5Hex := hex.EncodeToString(secondMD5[:])
		if finalMD5Hex != sign {
			response.Fail(c, 300008)
			c.Abort()
			return
		}
		// AES-128-CBC 解密 en_data
		enBytes, err := base64.StdEncoding.DecodeString(enData)
		if err != nil {
			response.Fail(c, 300009)
			c.Abort()
			return
		}
		block, err := aes.NewCipher(aesKey)
		if err != nil {
			response.Fail(c, 300010)
			c.Abort()
			return
		}
		mode := cipher.NewCBCDecrypter(block, aesIV)
		if len(enBytes)%aes.BlockSize != 0 {
			response.Fail(c, 300011)
			c.Abort()
			return
		}
		decrypted := make([]byte, len(enBytes))
		mode.CryptBlocks(decrypted, enBytes)
		pad := int(decrypted[len(decrypted)-1])
		if pad > aes.BlockSize || pad <= 0 {
			response.Fail(c, 300012)
			c.Abort()
			return
		}
		// 获取请求数据进行传递
		decrypted = decrypted[:len(decrypted)-pad]
		var request T
		if err := json.Unmarshal(decrypted, &request); err != nil {
			response.Fail(c, 300013)
			c.Abort()
			return
		}
		c.Set("payload", request)
		c.Next()
	}
}
