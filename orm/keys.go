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
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/pjacferreira/sqlf"
)

// Keys Object
type Key struct {
	dirty      bool       // Is Entry Dirty?
	stored     bool       // Is Entry Stored in Database
	id         *uint32    // LOCAL Key ID
	ciphertext []byte     // Encrypted KEY
	expiration *time.Time // Expiration Time Stamp
	creator    *uint64    // Global User ID of Creator
	created    *time.Time // Creation Timestamp
}

func NewKey(creator uint64, bytes []byte, exp time.Time) ([]byte, *Key, error) {
	// Create a Key Object
	k := &Key{}

	// Create Cypher Text
	key, e := k.EncryptKey(creator, bytes)
	if e != nil { // ERROR
		return nil, nil, e
	}

	// Set Expiration Time Stamp
	e = k.SetExpiration(exp)
	if e != nil { // ERROR
		return nil, nil, e
	}

	return key, k, nil
}

func (o *Key) IsDirty() bool {
	return o.dirty
}

func (o *Key) IsNew() bool {
	return !o.stored
}

func (o *Key) IsValid() bool {
	return o.creator != nil && o.ciphertext != nil
}

// ByID Finds Store By ID
func (o *Key) ByID(db *sql.DB, id uint32) error {
	// Reset Entry
	o.reset()

	// Execute Query
	var expires sql.NullString
	var creator sql.NullInt64
	var created sql.NullString
	e := sqlf.From("ciphers").
		Select("ciphertext").To(&o.ciphertext).
		Select("expiration").To(&expires).
		Select("id_creator").To(&creator).
		Select("created").To(&created).
		Where("id = ?", id).
		QueryRowAndClose(context.TODO(), db)

	// Error Executing Query?
	if e != nil && e != sql.ErrNoRows { // YES
		log.Printf("query error: %v\n", e)
		return e
	}

	// Did we retrieve an entry?
	if e == nil { // YES
		o.id = &id
		if expires.Valid {
			o.expiration = mySQLTimeStampToGoTime(expires.String)
		}
		if creator.Valid {
			cid := uint64(creator.Int64)
			o.creator = &cid
		}
		if created.Valid {
			o.created = mySQLTimeStampToGoTime(created.String)
		}
		o.stored = true
	}

	return nil
}

// ID Local ID
func (o *Key) ID() uint32 {
	if o.id == nil {
		return 0
	}

	return *o.id
}

func (o *Key) DecryptKey(key []byte) ([]byte, error) {
	// Do we have a Password Set?
	if len(o.ciphertext) == 0 { // NO
		return nil, nil
	}

	// Convert String to Byte Array
	return gcmDecrypt(key, o.ciphertext)
}

func (o *Key) Expiration() *time.Time {
	return o.expiration
}

func (o *Key) ExpirationUTC() string {
	utc := o.expiration.UTC()
	// RETURN ISO 8601 / RFC 3339 FORMAT in UTC
	return utc.Format(time.RFC3339)
}

func (o *Key) Creator() uint64 {
	if o.creator == nil {
		return 0
	}

	return *o.creator
}

func (o *Key) Created() *time.Time {
	return o.created
}

func (o *Key) SetExpiration(t time.Time) error {
	o.expiration = &t
	o.dirty = true

	return nil
}

func (o *Key) SetExpiresIn(days uint16) error {
	// Have DB Connection?
	if days == 0 { // NO: Abort
		return errors.New("Numer of days should be > 0")
	}

	now := time.Now()
	expires := now.AddDate(0, 0, int(days))
	o.expiration = &expires
	o.dirty = true
	return nil
}

func (o *Key) setCreator(id uint64) (uint64, error) {
	if !o.IsNew() {
		return 0, errors.New("Registered Keys is immutable")
	}

	// Current State
	current := uint64(0)
	if o.creator != nil {
		current = *o.creator
	}

	// New State
	o.creator = &id
	o.dirty = true
	return current, nil
}

func (o *Key) EncryptKey(creator uint64, bytes []byte) ([]byte, error) {
	_, e := o.setCreator(creator)
	if e != nil {
		return nil, e
	}

	key, e := o.generateEncryptionKey()
	if e != nil {
		return nil, e
	}

	// NOTE: We use SHA256 HASH because it is 32 bytes long and can be user with AES-256
	cypherbytes, e := gcmEncrypt(key, bytes)
	if e != nil {
		return nil, e
	}

	o.ciphertext = cypherbytes

	// Return Encryption Key
	return key, nil
}

func (o *Key) Flush(db sqlf.Executor, force bool) error {
	// Have DB Connection?
	if db == nil { // NO: Abort
		return errors.New("Missing Database Connection")
	}

	// Has entry been modified?
	if !force && !o.IsDirty() { // NO: Abort
		return nil
	}

	// Is New Entry?
	var e error
	if o.IsNew() { // YES: Create
		if !o.IsValid() {
			return errors.New("Invalid Key Object")
		}

		// Execute Insert
		s := sqlf.InsertInto("ciphers").
			Set("ciphertext", o.ciphertext).
			Set("expiration", goTimeToMySQLTimeStamp(o.expiration)).
			Set("id_creator", o.creator)

		// DEBUG: Print SQL
		fmt.Println(s.String())

		_, e = s.Exec(context.TODO(), db)

		// Error Occured?
		if e == nil { // NO: Get Last Insert ID
			var id uint32
			e = sqlf.Select("LAST_INSERT_ID()").
				To(&id).
				QueryRowAndClose(context.TODO(), db)

			// Error Occured?
			if e == nil { // NO: Set Object ID
				o.id = &id
			}
		}
	} else {
		return errors.New("Key Objects are Immutable")
	}

	if e == nil {
		o.stored = true
		o.dirty = false
	}
	return e
}

func (o *Key) reset() {
	// Clean Entry
	o.id = nil
	o.ciphertext = nil
	o.expiration = nil
	o.creator = nil
	o.created = nil

	// Mark State as Unregistered
	o.stored = false

	// Mark Entry as Clean
	o.dirty = false
}

func (o *Key) generateEncryptionKey() ([]byte, error) {
	if o.creator == nil {
		return nil, errors.New("Not Ready")
	}

	// Generate a Pseudo Random String to Hash
	rs := RandomAlphaNumericPunctuationString(128) // Random String to make things harder
	hashString := fmt.Sprintf("%d:%s:%d", o.creator, rs, time.Now().UTC().UnixNano())
	hash := sha256.Sum256([]byte(hashString))

	// Create HASH of Plain Text
	return hash[:], nil
}
