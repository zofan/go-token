package token

import (
	"gotest.tools/assert"
	"net"
	"testing"
	"time"
)

func TestDecode(t *testing.T) {
	gcm, _ := NewGCM(`grtfndgclktyzbag`)

	tn := &Token{
		ID:        45645,
		AccountID: 18574,
		Created:   time.Now(),
		Activated: time.Now().Add(time.Second * 1),
		Expired:   time.Now().Add(time.Hour * 24 * 365),
		Service:   567,
		Payload:   []byte(`hello!`),
		IP4:       net.ParseIP(`157.52.36.89`),
	}

	_, err := tn.Encode(gcm)
	assert.NilError(t, err)

	encodedBase64 := tn.EncodeBase64(gcm)
	assert.NilError(t, err)

	encodedHex := tn.EncodeHex(gcm)
	assert.NilError(t, err)

	encodedAscii85 := tn.EncodeAscii85(gcm)
	assert.NilError(t, err)

	tnBase64, err := Decode(encodedBase64, gcm)
	assert.NilError(t, err)
	assert.DeepEqual(t, tnBase64.Marshal(), tn.Marshal())
	assert.Equal(t, tnBase64.IP4.String(), tn.IP4.String())

	tnHex, err := Decode(encodedHex, gcm)
	assert.NilError(t, err)
	assert.DeepEqual(t, tnHex.Marshal(), tn.Marshal())

	tnAscii85, err := Decode(encodedAscii85, gcm)
	assert.NilError(t, err)
	assert.DeepEqual(t, tnAscii85.Marshal(), tn.Marshal())
}
