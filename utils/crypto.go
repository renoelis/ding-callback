package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	mrand "math/rand"
	"sort"
	"strings"
	"time"
)

// DingTalkCrypto 钉钉加密解密工具结构体
type DingTalkCrypto struct {
	Token          string
	EncodingAESKey string
	SuiteKey       string
	BKey           []byte
	Block          cipher.Block
}

// NewDingTalkCrypto 创建钉钉加密解密工具实例
func NewDingTalkCrypto(token, encodingAESKey, suiteKey string) *DingTalkCrypto {
	if len(encodingAESKey) != int(43) {
		panic("不合法的EncodingAESKey")
	}
	bkey, err := base64.StdEncoding.DecodeString(encodingAESKey + "=")
	if err != nil {
		panic(err.Error())
	}
	block, err := aes.NewCipher(bkey)
	if err != nil {
		panic(err.Error())
	}
	c := &DingTalkCrypto{
		Token:          token,
		EncodingAESKey: encodingAESKey,
		SuiteKey:       suiteKey,
		BKey:           bkey,
		Block:          block,
	}
	return c
}

// GetDecryptMsg 解密消息
func (c *DingTalkCrypto) GetDecryptMsg(signature, timestamp, nonce, secretMsg string) (string, error) {
	if !c.VerificationSignature(c.Token, timestamp, nonce, secretMsg, signature) {
		return "", errors.New("ERROR: 签名不匹配")
	}
	decode, err := base64.StdEncoding.DecodeString(secretMsg)
	if err != nil {
		return "", err
	}
	if len(decode) < aes.BlockSize {
		return "", errors.New("ERROR: 密文太短")
	}
	blockMode := cipher.NewCBCDecrypter(c.Block, c.BKey[:c.Block.BlockSize()])
	plantText := make([]byte, len(decode))
	blockMode.CryptBlocks(plantText, decode)
	plantText = pkCS7UnPadding(plantText)

	// 检查plantText长度是否足够
	if len(plantText) < 20 {
		return "", errors.New("ERROR: 解密后的数据长度不足")
	}

	size := binary.BigEndian.Uint32(plantText[16:20])

	// 检查size是否合理
	if size > uint32(len(plantText)-20) {
		return "", errors.New("ERROR: 解密后的数据长度与指定长度不匹配")
	}

	plantText = plantText[20:]
	corpID := plantText[size:]

	// 检查corpID是否合理
	if len(corpID) == 0 {
		return "", errors.New("ERROR: 解密后的corpID为空")
	}

	if string(corpID) != c.SuiteKey {
		return "", errors.New("ERROR: CorpID匹配不正确")
	}
	return string(plantText[:size]), nil
}

// GetEncryptMsg 加密消息
func (c *DingTalkCrypto) GetEncryptMsg(msg string) (map[string]string, error) {
	var timestamp = time.Now().Second()
	var nonce = randomString(12)
	str, sign, err := c.GetEncryptMsgDetail(msg, fmt.Sprint(timestamp), nonce)

	return map[string]string{"nonce": nonce, "timeStamp": fmt.Sprint(timestamp), "encrypt": str, "msg_signature": sign}, err
}

// GetEncryptMsgDetail 加密消息详情
func (c *DingTalkCrypto) GetEncryptMsgDetail(msg, timestamp, nonce string) (string, string, error) {
	size := make([]byte, 4)
	binary.BigEndian.PutUint32(size, uint32(len(msg)))
	msg = randomString(16) + string(size) + msg + c.SuiteKey
	plantText := pkCS7Padding([]byte(msg), c.Block.BlockSize())
	if len(plantText)%aes.BlockSize != 0 {
		return "", "", errors.New("ERROR: 消息体size不为16的倍数")
	}
	blockMode := cipher.NewCBCEncrypter(c.Block, c.BKey[:c.Block.BlockSize()])
	chipherText := make([]byte, len(plantText))
	blockMode.CryptBlocks(chipherText, plantText)
	outMsg := base64.StdEncoding.EncodeToString(chipherText)
	signature := c.CreateSignature(c.Token, timestamp, nonce, string(outMsg))
	return string(outMsg), signature, nil
}

// sha1Sign SHA1签名
func sha1Sign(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

// CreateSignature 创建签名
func (c *DingTalkCrypto) CreateSignature(token, timestamp, nonce, msg string) string {
	params := make([]string, 0)
	params = append(params, token)
	params = append(params, timestamp)
	params = append(params, nonce)
	params = append(params, msg)
	sort.Strings(params)
	return sha1Sign(strings.Join(params, ""))
}

// VerificationSignature 验证签名
func (c *DingTalkCrypto) VerificationSignature(token, timestamp, nonce, msg, sigture string) bool {
	return c.CreateSignature(token, timestamp, nonce, msg) == sigture
}

// pkCS7UnPadding 解密补位
func pkCS7UnPadding(plantText []byte) []byte {
	length := len(plantText)
	unpadding := int(plantText[length-1])
	return plantText[:(length - unpadding)]
}

// pkCS7Padding 加密补位
func pkCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// randomString 生成随机字符串
func randomString(n int, alphabets ...byte) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	var randby bool
	var rnd *mrand.Rand

	if num, err := rand.Read(bytes); num != n || err != nil {
		source := mrand.NewSource(time.Now().UnixNano())
		rnd = mrand.New(source)
		randby = true
	}

	for i, b := range bytes {
		if len(alphabets) == 0 {
			if randby {
				bytes[i] = alphanum[rnd.Intn(len(alphanum))]
			} else {
				bytes[i] = alphanum[b%byte(len(alphanum))]
			}
		} else {
			if randby {
				bytes[i] = alphabets[rnd.Intn(len(alphabets))]
			} else {
				bytes[i] = alphabets[b%byte(len(alphabets))]
			}
		}
	}
	return string(bytes)
}
