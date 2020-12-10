package token

import (
	"fmt"
	"gotest.tools/assert"
	"testing"
)

func TestDecode(t *testing.T) {
	gcm, _ := NewGCM(`grtfndgclktyzbag`)

	tn := &Token{
		ID:      45645,
		Account: 18574,
		Service: 567,
	}

	_, err := tn.Encode(gcm)
	assert.NilError(t, err)

	encodedBase64 := tn.EncodeBase64(gcm)
	assert.NilError(t, err)
	fmt.Println(encodedBase64)

	encodedHex := tn.EncodeHex(gcm)
	assert.NilError(t, err)
	fmt.Println(encodedHex)

	encodedAscii85 := tn.EncodeAscii85(gcm)
	assert.NilError(t, err)
	fmt.Println(encodedAscii85)

	tnBase64, err := Decode(encodedBase64, gcm)
	assert.NilError(t, err)
	assert.DeepEqual(t, tnBase64.Marshal(), tn.Marshal())

	tnHex, err := Decode(encodedHex, gcm)
	assert.NilError(t, err)
	assert.DeepEqual(t, tnHex.Marshal(), tn.Marshal())

	tnAscii85, err := Decode(encodedAscii85, gcm)
	assert.NilError(t, err)
	assert.DeepEqual(t, tnAscii85.Marshal(), tn.Marshal())
}
