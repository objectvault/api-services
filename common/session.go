// cSpell:ignore bson, gonic, paulo ferreira, userro
package common

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
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type StoreSession struct {
	store      uint64 // LOCAL Key ID
	key        []byte // Store Encryption Key
	expiration int64  // UTC Unix Expiration Time
	extend_by  uint16 // Default Session Expiration in Minutes
}

/* TODO: Add Single Use Store Session
 * this could be used in stores that require high security (i.e. per transaction validation)
 */

func ImportStoreSession(i string, valid_for uint16) (*StoreSession, error) {
	if valid_for == 0 {
		return nil, errors.New("Invalid Value for 'valid_for'")
	}

	// Create Object
	o := &StoreSession{
		extend_by: valid_for,
	}

	// Import Ok?
	e := o.Import(i)
	if e != nil { // NO: Abort
		return nil, e
	}

	// Calculate Expiration
	return o, nil
}

func NewStoreSession(store uint64, key []byte, valid_for uint16) (*StoreSession, error) {
	if key == nil {
		return nil, errors.New("Store 'key' missing value")
	}

	if valid_for == 0 {
		return nil, errors.New("Invalid Value for 'valid_for'")
	}

	o := &StoreSession{
		store:     store,
		key:       key,
		extend_by: valid_for,
	}

	// Calculate Expiration
	o.Extend(0)
	return o, nil
}

func (o *StoreSession) IsValid() bool {
	return o.store > 0 && len(o.key) > 0 && o.expiration > 0 && o.extend_by > 0
}

func (o *StoreSession) IsExpired() bool {
	if o.IsValid() {
		et := time.Now().Unix()
		return o.expiration < et
	}
	return true
}

func (o *StoreSession) Store() uint64 {
	return o.store
}

func (o *StoreSession) StoreHex() string {
	return fmt.Sprintf("%x", o.store)
}

func (o *StoreSession) Key() []byte {
	return o.key
}

func (o *StoreSession) KeyHex() string {
	h := hex.EncodeToString(o.key)
	return h
}

func (o *StoreSession) ExtendBy() uint16 {
	return o.extend_by
}

func (o *StoreSession) ExpireHex() string {
	return fmt.Sprintf("%x", o.expiration)
}

func (o *StoreSession) ExpireUnix() int64 {
	return o.expiration
}

func (o *StoreSession) ExpireTime() *time.Time {
	t := time.Unix(o.expiration, 0)
	return &t
}

func (o *StoreSession) Extend(d uint16) int64 {
	// Are we using a Different Extension Period?
	by := d
	if d == 0 { // NO: Use Default
		by = o.extend_by
	}

	// Calculate Expiration
	o.expiration = time.Now().Unix()
	o.expiration += int64(by * 60)
	return o.expiration
}

func (o *StoreSession) SetExtendBy(by uint16) (uint16, error) {
	current := o.extend_by

	if by == 0 {
		return 0, errors.New("Invalid Value for 'valid_for'")
	}

	return current, nil
}

func (o *StoreSession) Export() (string, error) {
	return fmt.Sprintf("/%s/%s/%s/", o.StoreHex(), o.KeyHex(), o.ExpireHex()), nil
}

func (o *StoreSession) Import(v string) error {

	// Test: Contains Possible Import Value
	v = strings.Trim(v, " ")
	if v[0] != '/' || v[len(v)-1] != '/' { // NO
		return errors.New("Not a valid import value in 'v'")
	}

	// Remove Leading and Trailing '/'
	s := v[1 : len(v)-1]

	// Do we have valid number of fields?
	parts := strings.Split(s, "/")
	if len(parts) != 3 {
		return errors.New("Not a valid import value in 'v'")
	}

	u, e := strconv.ParseUint(parts[0], 16, 64)
	if e != nil {
		return errors.New("Import Value 'v' contains an Invalid Store ID")
	}

	bs, e := hex.DecodeString(parts[1])
	if e != nil {
		return errors.New("Import Value 'v' contains an Invalid Key")
	}

	i, e := strconv.ParseInt(parts[2], 16, 64)
	if e != nil || i <= 0 {
		return errors.New("Import Value 'v' contains an Invalid Expiration Timestamp")
	}

	// Import StoreKey
	o.store = u
	o.key = bs
	o.expiration = i
	return nil
}
