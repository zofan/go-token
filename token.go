package token

import (
	"crypto/cipher"
	"encoding/ascii85"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/zofan/go-bits"
	"time"
)

const (
	encodeBase64  = `1`
	encodeHex     = `2`
	encodeAscii85 = `3`

	encodedSize = (8 * 2) + (2 * 4)
)

var (
	ErrInvalidToken = errors.New(`token: invalid rawToken`)
)

type Token struct {
	ID        uint64
	AccountID uint64
	Bits      bits.Bits32
	Service   int
}

func New() Token {
	now := time.Now()
	return Token{
		ID: uint64(now.UnixNano()),
	}
}

func Decode(rawToken string, gcm cipher.AEAD) (t *Token, err error) {
	var raw []byte

	raw, err = decodeRaw(rawToken)
	if err != nil {
		return
	}

	raw, err = decrypt(raw, gcm)
	if err != nil {
		return
	}

	t = &Token{
		ID:        binary.BigEndian.Uint64(raw[0:]),
		AccountID: binary.BigEndian.Uint64(raw[8:]),
		Bits:      bits.Bits32(binary.BigEndian.Uint32(raw[16:])),
		Service:   int(binary.BigEndian.Uint16(raw[20:])),
	}

	if t.Service == 0 || t.ID == 0 || t.AccountID == 0 {
		return nil, ErrInvalidToken
	}

	return
}

func decodeRaw(rawToken string) (raw []byte, err error) {
	if len(rawToken) <= 1 {
		return nil, ErrInvalidToken
	}

	switch rawToken[:1] {
	case encodeBase64:
		raw, err = base64.StdEncoding.DecodeString(rawToken[1:])
		if err != nil {
			return
		}
	case encodeHex:
		raw, err = hex.DecodeString(rawToken[1:])
		if err != nil {
			return
		}
	case encodeAscii85:
		buf := make([]byte, ascii85.MaxEncodedLen(len(rawToken)))
		n, _, err := ascii85.Decode(buf, []byte(rawToken[1:]), true)
		if err != nil {
			return nil, err
		}
		raw = buf[:n]
	default:
		return nil, ErrInvalidToken
	}

	if len(raw) < encodedSize {
		return nil, ErrInvalidToken
	}

	return
}

func (t *Token) Marshal() []byte {
	raw := make([]byte, encodedSize, encodedSize)

	binary.BigEndian.PutUint64(raw[0:], t.ID)
	binary.BigEndian.PutUint64(raw[8:], t.AccountID)
	binary.BigEndian.PutUint32(raw[16:], uint32(t.Bits))
	binary.BigEndian.PutUint16(raw[20:], uint16(t.Service))

	return raw
}

func (t *Token) Encode(gcm cipher.AEAD) ([]byte, error) {
	raw := t.Marshal()

	raw, err := encrypt(raw, gcm)
	if err != nil {
		return raw, err
	}

	return raw, nil
}

func (t *Token) EncodeBase64(gcm cipher.AEAD) string {
	raw, err := t.Encode(gcm)
	if err != nil {
		return ``
	}

	return encodeBase64 + base64.StdEncoding.EncodeToString(raw)
}

func (t *Token) EncodeHex(gcm cipher.AEAD) string {
	raw, err := t.Encode(gcm)
	if err != nil {
		return ``
	}

	return encodeHex + hex.EncodeToString(raw)
}

func (t *Token) EncodeAscii85(gcm cipher.AEAD) string {
	raw, err := t.Encode(gcm)
	if err != nil {
		return ``
	}

	buf := make([]byte, ascii85.MaxEncodedLen(len(raw)))
	n := ascii85.Encode(buf, raw)

	return encodeAscii85 + string(buf[:n])
}
