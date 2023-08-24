// cSpell:ignore hasher, userid
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

// cSpell:ignore ciphertext
import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/objectvault/api-services/orm/query"
	"github.com/pjacferreira/sqlf"
)

// User Object Definition
type UserRegistry struct {
	States
	dirty      bool    // Is Entry Dirty?
	stored     bool    // Is Entry Stored in Database
	id         *uint64 // KEY: GLOBAL User ID
	username   string  // User Alias
	email      string  // User Email
	name       string  // User Long Name
	state      uint16  // Global User State
	ciphertext []byte  // User Cipher Text to Validate Password
}

func UserRegistryFromUser(u *User) (*UserRegistry, error) {
	r := &UserRegistry{}
	r.UpdateRegistry(u)
	return r, nil
}

func RegisteredUsersCount(db *sql.DB, q query.TQueryConditions) (uint64, error) {
	// Query Results Values
	var count uint64

	// Create SQL Statement
	s := sqlf.From("registry_users").
		Select("COUNT(*)").To(&count)

	// Execute Count
	e := s.QueryRowAndClose(context.TODO(), db)

	// Error Occurred?
	if e != nil { // YES
		log.Printf("query error: %v\n", e)
		return 0, e
	}

	return count, nil
}

func RegisteredUsersQuery(db *sql.DB, q query.TQueryConditions, c bool) (query.TQueryResults, error) {
	var list query.QueryResults = query.QueryResults{}
	list.SetMaxLimit(100) // Hard Code Maximum Limit

	if q != nil {
		// Set Offset From Query Conditions
		if q.Offset() != nil {
			list.SetOffset(*q.Offset())
		}

		if q.Limit() != nil {
			list.SetLimit(*q.Limit())
		}
	}

	// Query Results Values
	var id uint64
	var state uint16
	var name, username, email string

	// Create SQL Statement
	s := sqlf.From("registry_users").
		Select("id_user").To(&id).
		Select("name").To(&name).
		Select("username").To(&username).
		Select("email").To(&email).
		Select("state").To(&state)

	// Is OFFSET Set?
	if list.Offset() > 0 { // YES: Use it
		s.Offset(list.Offset())
	}

	// Is LIMIT Set?
	if list.Limit() > 0 { // YES: Use it
		s.Limit(list.Limit())
	}

	// Do we have Sort Conditions
	if q != nil && q.Sort() != nil {
		var orderBy []string
		for _, i := range q.Sort() {
			if !i.Descending {
				orderBy = append(orderBy, i.Field)
			} else {
				orderBy = append(orderBy, i.Field+" DESC")
			}

			// Add Sort Conditions to Result
			list.AppendSort(i.Field, i.Descending)
		}

		// Apply Order By
		if len(orderBy) > 0 {
			s.OrderBy(orderBy...)
		}
	} else { // DEFAULT: Sort by ID
		list.AppendSort("id_user", false)
		s.OrderBy("id_user")
	}

	// Apply Extra Query Conditions
	e := query.ApplyFilterConditions(s, q)
	if e != nil { // Error Occurred
		// DEBUG: Print SQL
		log.Print(s.String())
		log.Printf("query error: %v\n", e)
		return nil, e
	}

	// DEBUG: Print SQL
	fmt.Print(s.String())

	e = s.QueryAndClose(context.TODO(), db, func(row *sql.Rows) {
		userid := id

		u := UserRegistry{
			id:       &userid,
			name:     name,
			username: username,
			email:    email,
			state:    state,
		}

		list.AppendValue(&u)
	})

	// Error Occurred?
	if e != nil && e != sql.ErrNoRows { // YES
		log.Printf("query error: %v\n", e)
		return nil, e
	}

	// Is Count of Entries Requested?
	if c { // YES: Count Entries under Same Conditions
		count, e := RegisteredUsersCount(db, q)
		if e != nil {
			return nil, e
		}

		list.SetMaxCount(count)
	}

	return list, nil
}

// IsDirty Have the Object Properties Changed since last Serialization?
func (o *UserRegistry) IsDirty() bool {
	return o.dirty
}

func (o *UserRegistry) IsNew() bool {
	return !o.stored
}

func (o *UserRegistry) IsValid() bool {
	return o.id != nil && o.username != "" && o.email != ""
}

func (o *UserRegistry) Find(db *sql.DB, id interface{}) error {
	// is ID an integer?
	if _, ok := id.(uint64); ok { // YES: Find by ID
		return o.ByID(db, id.(uint64))
	} else if _, ok := id.(string); ok { // ELSE: Find by Alias or Email
		sid := id.(string)
		if !strings.Contains(sid, "@") {
			return o.ByUserName(db, sid)
		}

		return o.ByEmail(db, sid)
	}
	// ELSE: Missing or Invalid id
	return errors.New("'id' missing or of invalid type")
}

// ByID Finds User By ID
func (o *UserRegistry) ByID(db *sql.DB, id uint64) error {
	// Reset Entry
	o.reset()

	// Execute Query
	var ciphertext sql.NullString
	e := sqlf.From("registry_users").
		Select("name").To(&o.name).
		Select("username").To(&o.username).
		Select("email").To(&o.email).
		Select("state").To(&o.state).
		Select("ciphertext").To(&ciphertext).
		Where("id_user = ?", id).
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
		o.stored = true
	}

	return nil
}

// ByUserName Finds User By User Name
func (o *UserRegistry) ByUserName(db *sql.DB, username string) error {
	// Validate user name
	username = strings.TrimSpace(username)
	if username == "" {
		return errors.New("Missing Required Parameter 'username'")
	}

	// Reset Entry
	o.reset()

	// Is Incoming Parameter Valid?
	if username == "" { // NO
		return errors.New("Missing Required Parameter 'username'")
	}

	// Aliases are always lower case
	username = strings.ToLower(username)

	// Execute Query
	var id uint64
	var ciphertext sql.NullString
	e := sqlf.From("registry_users").
		Select("id_user").To(&id).
		Select("name").To(&o.name).
		Select("email").To(&o.email).
		Select("state").To(&o.state).
		Select("ciphertext").To(&ciphertext).
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

		o.stored = true
	}

	return nil
}

// ByEmail Finds User By Email
func (o *UserRegistry) ByEmail(db *sql.DB, email string) error {
	// Reset Entry
	o.reset()

	// Is Incoming Parameter Valid?
	if email == "" { // NO
		return errors.New("Missing Required Parameter 'email'")
	}

	// Emails are always lower case
	email = strings.ToLower(email)

	// Execute Query
	var id uint64
	var ciphertext sql.NullString
	e := sqlf.From("registry_users").
		Select("id_user").To(&id).
		Select("name").To(&o.name).
		Select("username").To(&o.username).
		Select("state").To(&o.state).
		Select("ciphertext").To(&ciphertext).
		Where("email = ?", email).
		QueryRowAndClose(context.TODO(), db)

	// Did we retrieve an entry?
	if e == nil { // YES
		o.id = &id
		o.email = email

		if ciphertext.Valid {
			s := ciphertext.String
			o.ciphertext = []byte(s)
		}

		o.stored = true
	}

	return nil
}

// ID Get User ID
func (o *UserRegistry) ID() uint64 {
	return *o.id
}

func (o *UserRegistry) Name() string {
	return o.name
}

func (o *UserRegistry) UserName() string {
	return o.username
}

func (o *UserRegistry) Email() string {
	return o.email
}

func (o *UserRegistry) State() uint16 {
	return o.state
}

func (o *UserRegistry) HasAnyStates(states uint16) bool {
	return HasAnyStates(o.state, states)
}

func (o *UserRegistry) HasAllStates(states uint16) bool {
	return HasAllStates(o.state, states)
}

func (o *UserRegistry) IsActive() bool {
	// User Account Active
	return !HasAnyStates(o.state, STATE_INACTIVE|STATE_BLOCKED|STATE_DELETE)
}

func (o *UserRegistry) IsBlocked() bool {
	// GLOBAL User Access Blocked
	return HasAnyStates(o.state, STATE_BLOCKED|STATE_DELETE)
}

func (o *UserRegistry) IsDeleted() bool {
	// GLOBAL User marked for Deletion
	return HasAnyStates(o.state, STATE_DELETE)
}

func (o *UserRegistry) IsReadOnly() bool {
	return HasAllStates(o.state, STATE_READONLY)
}

func (o *UserRegistry) SetID(id uint64) (uint64, error) {
	if o.IsNew() {
		// Current State
		current := uint64(0)
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

func (o *UserRegistry) SetName(name string) (string, error) {
	// Current State
	current := o.name

	// Validate Alias
	name = strings.TrimSpace(name)
	if name == "" {
		return current, errors.New("Missing Value for Name")
	}

	// New State
	o.name = name
	o.dirty = true
	return current, nil
}

func (o *UserRegistry) SetUserName(username string) (string, error) {
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
	return current, nil
}

func (o *UserRegistry) SetEmail(email string) (string, error) {
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
	return current, nil
}

func (o *UserRegistry) SetStates(states uint16) {
	// Current State
	current := o.state

	// New State
	o.state = SetStates(o.state, states)
	if o.state != current {
		o.dirty = true
	}
}

func (o *UserRegistry) ClearStates(states uint16) {
	// Current State
	current := o.state

	// New State
	o.state = ClearStates(o.state, states)
	if o.state != current {
		o.dirty = true
	}
}

func (o *UserRegistry) UpdateRegistry(u *User) error {
	// Update Basic User Information
	o.username = u.UserName()
	o.email = u.Email()
	o.name = u.Name()
	o.ciphertext = u.ciphertext

	if !o.IsNew() {
		o.dirty = true
	}

	return nil
}

func (o *UserRegistry) UpdatePassword(u *User) error {
	// Update User Password Hash
	o.ciphertext = u.ciphertext

	if !o.IsNew() {
		o.dirty = true
	}

	return nil
}

func (o *UserRegistry) TestPassword(password string) bool {
	// Convert USer Password to HASH
	hasher := sha256.Sum256([]byte(password))

	// TEST Hash Against User
	return o.testHash(hasher[:])
}

func (o *UserRegistry) TestHash(hash string) bool {
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

func (o *UserRegistry) Flush(db sqlf.Executor, force bool) error {
	// Have DB Connection?
	if db == nil { // NO: Abort
		return errors.New("Missing Database Connection")
	}

	// Valid Entry?
	if !o.IsValid() { // NO: Abort
		return errors.New("Invalid Entry")
	}

	// Has entry been modified?
	if !force && !o.IsDirty() { // NO: Abort
		return nil
	}

	// Is New Entry?
	var e error
	if o.IsNew() { // YES: Create
		// Is Password  Set
		if len(o.ciphertext) == 0 {
			return errors.New("Password not Set")
		}

		s := sqlf.InsertInto("registry_users").
			Set("id_user", o.id).
			Set("name", o.name).
			Set("username", o.username).
			Set("email", o.email).
			Set("state", o.state).
			Set("ciphertext", o.ciphertext)

		_, e = s.ExecAndClose(context.TODO(), db)
	} else { // NO: Update
		// TODO: Create Special Update to Change User Password
		s := sqlf.Update("registry_users").
			Set("name", o.name).
			Set("username", o.username).
			Set("email", o.email).
			Set("state", o.state)

		if len(o.ciphertext) > 0 {
			s.Set("ciphertext", o.ciphertext)
		}

		_, e = s.
			Where("id_user = ?", o.id).
			ExecAndClose(context.TODO(), db)
	}

	if e == nil {
		o.stored = true
		o.dirty = false
	}
	return e
}

func (o *UserRegistry) testHash(hash []byte) bool {
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
	return e == nil
}

func (o *UserRegistry) reset() {
	// Clean Entry
	o.id = nil
	o.username = ""
	o.email = ""
	o.name = ""
	o.state = 0
	o.ciphertext = nil

	// Mark State as Unregistered
	o.stored = false

	// Mark Entry as Clean
	o.dirty = false
}
