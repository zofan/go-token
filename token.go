package token

import (
	"crypto/cipher"
	"encoding/ascii85"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/zofan/go-bits"
	"net"
	"time"
)

const (
	encodeBase64  = `10`
	encodeHex     = `20`
	encodeAscii85 = `30`

	encodedSize = (8 * 8) + (2 * 4)
)

var (
	ErrInvalidToken = errors.New(`token: invalid rawToken`)
)

type Token struct {
	ID        uint64
	AccountID uint64

	GroupBits  bits.Bits64
	AccessBits bits.Bits64
	FlagsBits  bits.Bits64

	Created   time.Time
	Activated time.Time
	Expired   time.Time

	Service int
	IP4     net.IP

	Payload []byte
}

func New() Token {
	now := time.Now()
	return Token{
		ID:      uint64(now.UnixNano()),
		Created: now,
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

		GroupBits:  bits.Bits64(binary.BigEndian.Uint64(raw[16:])),
		AccessBits: bits.Bits64(binary.BigEndian.Uint64(raw[24:])),
		FlagsBits:  bits.Bits64(binary.BigEndian.Uint64(raw[32:])),

		Created:   time.Unix(int64(binary.BigEndian.Uint64(raw[40:])), 0),
		Activated: time.Unix(int64(binary.BigEndian.Uint64(raw[48:])), 0),
		Expired:   time.Unix(int64(binary.BigEndian.Uint64(raw[56:])), 0),

		Service: int(binary.BigEndian.Uint32(raw[64:])),
		IP4:     net.IP(raw[68:]),

		Payload: raw[encodedSize:],
	}

	if t.Service == 0 || t.ID == 0 || t.AccountID == 0 {
		return nil, ErrInvalidToken
	}

	return
}

func decodeRaw(rawToken string) (raw []byte, err error) {
	if len(rawToken) <= 2 {
		return nil, ErrInvalidToken
	}

	switch rawToken[:2] {
	case encodeBase64:
		raw, err = base64.StdEncoding.DecodeString(rawToken[2:])
		if err != nil {
			return
		}
	case encodeHex:
		raw, err = hex.DecodeString(rawToken[2:])
		if err != nil {
			return
		}
	case encodeAscii85:
		buf := make([]byte, ascii85.MaxEncodedLen(len(rawToken)))
		n, _, err := ascii85.Decode(buf, []byte(rawToken[2:]), true)
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

	binary.BigEndian.PutUint64(raw[16:], uint64(t.GroupBits))
	binary.BigEndian.PutUint64(raw[24:], uint64(t.AccessBits))
	binary.BigEndian.PutUint64(raw[32:], uint64(t.FlagsBits))

	binary.BigEndian.PutUint64(raw[40:], uint64(t.Created.Unix()))
	binary.BigEndian.PutUint64(raw[48:], uint64(t.Activated.Unix()))
	binary.BigEndian.PutUint64(raw[56:], uint64(t.Expired.Unix()))

	binary.BigEndian.PutUint32(raw[64:], uint32(t.Service))

	for i, b := range t.IP4.To4() {
		raw[68+i] = b
	}

	return append(raw, t.Payload...)
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

func (t *Token) IsActive() bool {
	now := time.Now()
	return now.Sub(t.Activated) > 0 && now.Sub(t.Expired) < 0
}

func (t *Token) Epoch() uint16 {
	return uint16(t.Created.YearDay() % 365)
}
