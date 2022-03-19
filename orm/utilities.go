// cSpell:ignore bson, paulo ferreira
package orm

/*
 * This file is part of the ObjectVault Project.
 * Copyright (C) 2020-2022 Paulo Ferreira <vault at sourcenotes.org>
 *
 * This work is published under the GNU AGPLv3.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

import (
	"crypto/aes"
	"crypto/cipher"
	cryptorand "crypto/rand"
	"errors"
	"io"
	"math"
	"math/rand"
)

// GLOBAL HELPERs //
const aBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const nBytes = "0123456789"
const pBytes = "\\|!\"@#$%&/()=?+*'`~^,;.:-_"
const anBytes = aBytes + nBytes
const apBytes = aBytes + pBytes
const anpBytes = aBytes + nBytes + pBytes

func randomByteString(set string, length int) string {
	b := make([]byte, length)
	bits := int(math.Ceil(math.Log2(float64(len(set))))) // Maximum Number of Bits for Set Index
	mask := int64(1) << (bits - 1)
	maxLetters := 63 / bits
	setLength := len(set)

	for i, cache, remain := length-1, rand.Int63(), maxLetters; i >= 0; {
		// Exhausted Random Bits?
		if remain == 0 { // YES: Reload
			cache, remain = rand.Int63(), maxLetters
		}

		// Next Set Index
		idx := int(cache & mask)

		// idx outside of set?
		if idx > setLength { // YES: Loop back to start of set
			idx = idx % setLength
		}

		// Set Next Random Character
		b[i] = set[idx]

		// Prepare for Next Cycle
		i--
		cache >>= bits
		remain--
	}

	return string(b)
}

func RandomAlphaString(n int) string {
	return randomByteString(aBytes, n)
}

func RandomAlphaNumericString(n int) string {
	return randomByteString(anBytes, n)
}

func RandomAlphaNumericPunctuationString(n int) string {
	return randomByteString(anpBytes, n)
}

func gcmDecrypt(key, cipherbytes []byte) ([]byte, error) {
	// Create and Initialize Block Cypher //
	block, e := aes.NewCipher(key)
	if e != nil {
		return nil, e
	}

	// Setup Galois Counter Mode
	aesGCM, e := cipher.NewGCM(block)
	if e != nil {
		return nil, e
	}

	// Extract NONCE
	nonceSize := aesGCM.NonceSize()

	// Extract the nonce from the encrypted data
	nonce := cipherbytes[:nonceSize]
	ciphertext := cipherbytes[nonceSize:]

	// Decrypt the data
	plainbytes, e := aesGCM.Open(nil, nonce, ciphertext, nil)
	if e != nil {
		return nil, e
	}

	return plainbytes, nil
}

func gcmEncrypt(key []byte, bytes []byte) ([]byte, error) {
	// NOTE: We use SHA256 HASH because it is 32 bytes long and can be user with AES-256
	if len(key) != 32 {
		return nil, errors.New("Encryption KEY not Strong Enough")
	}

	// Create Block Cipher //
	block, e := aes.NewCipher(key)
	if e != nil {
		return nil, e
	}

	// https://en.wikipedia.org/wiki/Galois/Counter_Mode
	aesGCM, e := cipher.NewGCM(block)
	if e != nil {
		return nil, e
	}

	nonce := make([]byte, aesGCM.NonceSize())
	_, e = io.ReadFull(cryptorand.Reader, nonce)
	if e != nil {
		return nil, e
	}

	cipherbytes := aesGCM.Seal(nonce, nonce, bytes, nil)
	return cipherbytes, nil
}
