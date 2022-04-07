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
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/pjacferreira/sqlf"
)

func toCypherBytes(hash []byte, plainbytes []byte) ([]byte, error) {
	// Encrypt User Validation Text using User Password Hash
	cypherbytes, e := gcmEncrypt(hash, plainbytes)
	if e != nil {
		return nil, e
	}

	return cypherbytes, nil
}

func toPlainBytes(hash []byte, cipherbytes []byte) ([]byte, error) {
	// Try to Decrypt Validation Text
	plainbytes, e := gcmDecrypt(hash, cipherbytes)
	if e != nil {
		return nil, e
	}

	return plainbytes, nil
}

/* TODO Evaluate Security Risk
 * Is it better to save the Users Password Salted in the Database
 * or have the transmitted password be salted?
 * Current Solution unsalted password in database, salted request password
 */

// User Profile
type User struct {
	dirty          bool       // Is Entry Dirty?
	updateRegistry bool       // Do we need to Update the Registry?
	stored         bool       // Is Entry Stored in Database
	id             *uint32    // LOCAL User ID
	username       string     // User Alias
	name           string     // User Long Name
	email          string     // User Email
	object         string     // JSON Object String
	ciphertext     []byte     // COPY: User Cipher Text to Validate Password
	expires        *time.Time // Date Time Expires Password
	lastpwdchange  *time.Time // Date Time of Last Password Change
	maxpwddays     *uint16
	creator        *uint64    // Global User ID of Creator
	created        *time.Time // Created TimeStamp
	modifier       *uint64    // Global User ID of Last Modifier
	modified       *time.Time // Last Modification TimeStamp
}

// TODO Implement Delete (Both From Within an Entry and Without a Structure)

// IsDirty Have the Object Properties Changed since last Serialization?
func (o *User) IsDirty() bool {
	return o.dirty
}

func (o *User) UpdateRegistry() bool {
	return o.dirty && o.updateRegistry
}
func (o *User) IsNew() bool {
	return !o.stored
}

func (o *User) IsValid() bool {
	return o.id != nil && o.username != "" && o.email != "" && len(o.ciphertext) > 0
}

func (o *User) Find(db *sql.DB, id interface{}) error {
	// is ID an integer?
	if _, ok := id.(uint64); ok { // YES: Find by ID
		oid := id.(uint64)
		return o.ByID(db, uint32(oid))
	} else if _, ok := id.(string); ok { // ELSE: Find by Alias or Email
		sid := id.(string)
		if strings.Index(sid, "@") < 0 {
			return o.ByUserName(db, sid)
		}

		return o.ByEmail(db, sid)
	}
	// ELSE: Missing or Invalid id
	return errors.New("'id' missing or of invalid type")
}

// ByID Finds User By ID
func (o *User) ByID(db *sql.DB, id uint32) error {
	// Reset Entry
	o.reset()

	// Execute Query
	var ciphertext sql.NullString
	var expires sql.NullString
	var lastpwdchange sql.NullString
	var maxpwddays sql.NullInt32
	var object sql.NullString
	var created sql.NullString
	var modifier sql.NullInt64
	var modified sql.NullString

	// TODO Process Database "object" field
	e := sqlf.From("users").
		Select("name").To(&o.name).
		Select("username").To(&o.username).
		Select("email").To(&o.email).
		Select("object").To(&object).
		Select("ciphertext").To(&ciphertext).
		Select("dt_expires").To(&expires).
		Select("dt_lastpwdchg").To(&lastpwdchange).
		Select("maxpwddays").To(&maxpwddays).
		Select("creator").To(&o.creator).
		Select("created").To(&created).
		Select("modifier").To(&modifier).
		Select("modified").To(&modified).
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

		if ciphertext.Valid {
			s := ciphertext.String
			o.ciphertext = []byte(s)
		}
		if expires.Valid {
			o.expires = mySQLTimeStampToGoTime(expires.String)
		}
		if lastpwdchange.Valid {
			o.lastpwdchange = mySQLTimeStampToGoTime(lastpwdchange.String)
		}
		if created.Valid {
			o.created = mySQLTimeStampToGoTime(created.String)
		}
		if modifier.Valid {
			m := uint64(modifier.Int64)
			o.modifier = &m
			if modified.Valid {
				o.modified = mySQLTimeStampToGoTime(created.String)
			}
		}

		// TODO Deal with Object
		o.stored = true
	}

	return nil
}

// ByUserName Finds User By User Name
func (o *User) ByUserName(db *sql.DB, username string) error {
	// Is Incoming Parameter Valid?
	if username == "" { // NO
		return errors.New("Missing Required Parameter 'username'")
	}

	// Reset Entry
	o.reset()

	// Aliases are always lower case
	username = strings.ToLower(username)

	// Execute Query
	var id uint32
	var ciphertext sql.NullString
	var expires sql.NullString
	var lastpwdchange sql.NullString
	var maxpwddays sql.NullInt32
	var object sql.NullString
	var created sql.NullString
	var modifier sql.NullInt64
	var modified sql.NullString

	// TODO Process Database "object" field
	e := sqlf.From("users").
		Select("id").To(&id).
		Select("name").To(&o.name).
		Select("email").To(&o.email).
		Select("object").To(&object).
		Select("ciphertext").To(&ciphertext).
		Select("dt_expires").To(&expires).
		Select("dt_lastpwdchg").To(&lastpwdchange).
		Select("maxpwddays").To(&maxpwddays).
		Select("creator").To(&o.creator).
		Select("created").To(&created).
		Select("modifier").To(&modifier).
		Select("modified").To(&modified).
		Where("username = ?", username).
		QueryRowAndClose(context.TODO(), db)

	// Error Executing Query?
	if e != nil && e != sql.ErrNoRows { // YES
		log.Printf("query error: %v\n", e)
		return e
	}

	// Did we retrieve an entry?
	if e == nil { // YES
		o.id = &id
		o.username = username

		if ciphertext.Valid {
			s := ciphertext.String
			o.ciphertext = []byte(s)
		}
		if expires.Valid {
			o.expires = mySQLTimeStampToGoTime(expires.String)
		}
		if lastpwdchange.Valid {
			o.lastpwdchange = mySQLTimeStampToGoTime(lastpwdchange.String)
		}
		if created.Valid {
			o.created = mySQLTimeStampToGoTime(created.String)
		}
		if modifier.Valid {
			m := uint64(modifier.Int64)
			o.modifier = &m
			if modified.Valid {
				o.modified = mySQLTimeStampToGoTime(created.String)
			}
		}

		// TODO Deal with Object
		o.stored = true
	}

	return nil
}

// ByEmail Finds User By Email
func (o *User) ByEmail(db *sql.DB, email string) error {
	// Is Incoming Parameter Valid?
	if email == "" { // NO
		return errors.New("Missing Required Parameter 'email'")
	}

	// Reset Entry
	o.reset()

	// Trim Incoming String
	email = strings.TrimSpace(email)

	// Emails are always lower case
	email = strings.ToLower(email)

	// Execute Query
	var id uint32
	var ciphertext sql.NullString
	var expires sql.NullString
	var lastpwdchange sql.NullString
	var maxpwddays sql.NullInt32
	var object sql.NullString
	var created sql.NullString
	var modifier sql.NullInt64
	var modified sql.NullString

	// TODO Process Database "object" field
	e := sqlf.From("users").
		Select("id").To(&id).
		Select("name").To(&o.name).
		Select("username").To(&o.username).
		Select("object").To(&object).
		Select("ciphertext").To(&ciphertext).
		Select("dt_expires").To(&expires).
		Select("dt_lastpwdchg").To(&lastpwdchange).
		Select("maxpwddays").To(&maxpwddays).
		Select("creator").To(&o.creator).
		Select("created").To(&created).
		Select("modifier").To(&modifier).
		Select("modified").To(&modified).
		Where("email = ?", email).
		QueryRowAndClose(context.TODO(), db)

	// Error Executing Query?
	if e != nil && e != sql.ErrNoRows { // YES
		log.Printf("query error: %v\n", e)
		return e
	}

	// Did we retrieve an entry?
	if e == nil { // YES
		o.id = &id
		o.email = email

		if ciphertext.Valid {
			s := ciphertext.String
			o.ciphertext = []byte(s)
		}
		if expires.Valid {
			o.expires = mySQLTimeStampToGoTime(expires.String)
		}
		if lastpwdchange.Valid {
			o.lastpwdchange = mySQLTimeStampToGoTime(lastpwdchange.String)
		}
		if created.Valid {
			o.created = mySQLTimeStampToGoTime(created.String)
		}
		if modifier.Valid {
			m := uint64(modifier.Int64)
			o.modifier = &m
			if modified.Valid {
				o.modified = mySQLTimeStampToGoTime(created.String)
			}
		}

		// TODO Deal with Object
		o.stored = true
	}

	return nil
}

// ID Get User ID
func (o *User) ID() uint32 {
	if o.id == nil {
		return 0
	}

	return *o.id
}

func (o *User) UserName() string {
	return o.username
}

func (o *User) Email() string {
	return o.email
}

func (o *User) Name() string {
	return o.name
}

func (o *User) Creator() uint64 {
	if o.creator == nil {
		return 0
	}
	return *o.creator
}

func (o *User) Created() *time.Time {
	return o.created
}

func (o *User) Modifier() *uint64 {
	if o.modifier == nil {
		return nil
	}
	return o.modifier
}

func (o *User) Modified() *time.Time {
	return o.modified
}

func (o *User) SetID(id uint32) (uint32, error) {
	if o.IsNew() {
		// Current State
		current := uint32(0)
		if o.id != nil {
			current = *o.id
		}

		// New State
		o.id = &id
		o.dirty = true
		return current, nil
	}

	return 0, errors.New("Registered User - ID is immutable")
}

func (o *User) SetUserName(username string) (string, error) {
	// Current State
	current := o.username

	// Validate Alias
	username = strings.TrimSpace(username)
	if username == "" {
		return current, errors.New("Missing Value for User Name")
	}

	// Aliases are always lower case
	username = strings.ToLower(username)

	// New State
	o.username = username
	o.dirty = true
	o.updateRegistry = true // Need to Update Registry
	return current, nil
}

func (o *User) SetEmail(email string) (string, error) {
	// Current State
	current := o.email

	// Validate Email
	email = strings.TrimSpace(email)
	if email == "" {
		return current, errors.New("Missing Value for Email")
	}

	// New State
	o.email = strings.ToLower(email)
	o.dirty = true
	o.updateRegistry = true // Need to Update Registry
	return current, nil
}

func (o *User) SetName(name string) (string, error) {
	// Current State
	current := o.name

	// Is Incoming Parameter Valid?
	name = strings.TrimSpace(name)
	if name == "" { // NO
		return current, errors.New("Missing Required Parameter 'name'")
	}

	// New State
	o.name = name
	o.dirty = true
	o.updateRegistry = true // Need to Update Registry
	return current, nil
}

func (o *User) SetCreator(id uint64) error {
	// Is Record New?
	if o.IsNew() { // YES
		o.creator = &id

		// Creation Time Stamp AUTO SET by MySQL
		return nil
	}

	return errors.New("Registered User - Creator Cannot be Changed")
}

func (o *User) SetModifier(id uint64) error {
	// Set Modifier
	o.modifier = &id

	// Modification Time Stamp AUTO SET by MySQL
	return nil
}

func (o *User) SetHash(hash string) error {
	// Is Password Hash Valid
	if hash == "" {
		return errors.New("Missing Password Hash")
	}

	// Are we Creating a User?
	if !o.IsNew() { // NO: Use UpdateHash(...)
		return errors.New("Password Hash has to be updated")
	}

	h, e := hex.DecodeString(hash)
	if e != nil {
		return e
	}

	// Get Semi Random Text Bytes to Encrypt
	pb, e := o.generatePlainText()
	if e != nil {
		return e
	}

	// Did We Generate Cypher Text?
	ct, e := toCypherBytes(h, pb)
	if e != nil { // NO: ABORT
		return e
	}

	// New State
	o.ciphertext = ct
	o.dirty = true
	o.updateRegistry = true
	return nil
}

func (o *User) UpdateHash(old, hash string) error {
	// Is Password Hash Valid
	if old == "" || hash == "" {
		return errors.New("Missing Password Hashes")
	}

	// Are we Creating a User?
	if o.IsNew() || len(o.ciphertext) == 0 { // NO: Use SetHash(...)
		return errors.New("Password Hash has to be set")
	}

	ho, e := hex.DecodeString(old)
	if e != nil {
		return e
	}

	hn, e := hex.DecodeString(hash)
	if e != nil {
		return e
	}

	// Decrypt Existing CipherText //
	// Decrypt Bytes
	pb, e := toPlainBytes(ho, o.ciphertext)
	if e != nil {
		return e
	}

	// Re-encrypt using Hash Hash
	cb, e := toCypherBytes(hn, pb)
	if e != nil { // NO: ABORT
		return e
	}

	// New State
	o.ciphertext = cb
	o.dirty = true
	o.updateRegistry = true
	return nil
}

func (o *User) TestPassword(password string) bool {
	// Convert USer Password to HASH
	hasher := sha256.Sum256([]byte(password))

	// TEST Hash Against User
	return o.testHash(hasher[:])
}

func (o *User) TestHash(hash string) bool {
	// Was Password Hash Given?
	if hash == "" { // NO: Never Match
		return false
	}

	// Is Hex String?
	h, e := hex.DecodeString(hash)
	if e != nil { // NO: Fail
		return false
	}

	// TEST Hash Against User
	return o.testHash(h)
}

func (o *User) Flush(db sqlf.Executor, force bool) error {
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

		// Is Creator Set
		if o.creator == nil {
			return errors.New("Creation User not Set")
		}

		// Is Password Hash Set
		if len(o.ciphertext) == 0 {
			return errors.New("Password Hash not Set")
		}

		// CREATE USER RECORD: Execute Insert
		s := sqlf.InsertInto("users").
			Set("name", o.name).
			Set("username", o.username).
			Set("email", o.email).
			Set("ciphertext", o.ciphertext).
			Set("creator", o.creator)

		_, e = s.Exec(context.TODO(), db)

		// Error Occured?
		if e != nil { // YES: Abort
			return e
		}

		// Get Last Insert ID (the Local User ID)
		var id uint32
		e = sqlf.Select("LAST_INSERT_ID()").
			To(&id).
			QueryRow(context.TODO(), db)

		// Error Occured?
		if e != nil { // YES: Abort
			return e
		}

		// Set User ID
		o.id = &id
	} else { // NO: Update
		if o.id == nil {
			return errors.New("User ID not Set")
		}

		if o.modifier == nil {
			return errors.New("Modification User not Set")
		}

		s := sqlf.Update("users").
			Set("name", o.name).
			Set("username", o.username).
			Set("email", o.email).
			Set("modifier", o.modifier).
			Where("id = ?", o.id)

		if len(o.ciphertext) > 0 {
			s.Set("ciphertext", o.ciphertext)
		}

		_, e = s.ExecAndClose(context.TODO(), db)
	}

	if e == nil {
		o.stored = true
		o.dirty = false
	}
	return e
}

func (o *User) generatePlainText() ([]byte, error) {
	if o.username == "" || o.email == "" {
		return nil, errors.New("Not Ready")
	}

	// Generate a Pseudo Random String to Hash
	rs := RandomAlphaNumericPunctuationString(128) // Random String to make things harder
	hashString := fmt.Sprintf("%s:%s:%s", o.username, rs, o.email)
	hash := sha256.Sum256([]byte(hashString))

	// Create Plain Text
	return hash[:], nil
}

func (o *User) testHash(hash []byte) bool {
	// Is Possible Password Hash?
	if hash == nil || len(hash) != 32 { // NO: Fail
		return false
	}

	// Does User have Password Set?
	if len(o.ciphertext) == 0 { // NO: Fail
		return false
	}

	// Does HASH Decode Cypher Text?
	_, e := toPlainBytes(hash, o.ciphertext)
	if e != nil { // NO: Invalid Password Hash
		return false
	}

	// OK
	return true
}

func (o *User) reset() {
	// Clean Entry
	o.id = nil
	o.username = ""
	o.name = ""
	o.email = ""
	o.object = ""
	o.ciphertext = nil
	o.expires = nil
	o.lastpwdchange = nil
	o.maxpwddays = nil
	o.creator = nil
	o.created = nil
	o.modifier = nil
	o.modified = nil

	// Mark State as Unregistered
	o.stored = false

	// Mark Entry as Clean
	o.dirty = false
	o.updateRegistry = false
}
