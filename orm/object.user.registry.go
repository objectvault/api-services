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

// cSpell:ignore cypherbytes, ciphertext, plainbytes, userid

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"

	"github.com/objectvault/api-services/orm/query"
	"github.com/pjacferreira/sqlf"
)

// Org User Registry Definition
type ObjectUserRegistry struct {
	States
	dirty      bool    // Is Entry Dirty?
	stored     bool    // Is Entry Stored in Database
	object     *uint64 // KEY: GLOBAL Object ID
	user       *uint64 // KEY: GLOBAL User ID
	username   string  // User Name (ALIAS)
	state      uint16  // User State in Object
	roles      S_Roles // User Roles in Object
	ciphertext []byte  // Encrypted Store Key
}

func ObjectUsersDeleteAll(db *sql.DB, object uint64) (uint64, error) {
	// Create SQL Statement
	s := sqlf.DeleteFrom("registry_object_users").
		Where("id_object= ?", object)

	// Execute
	r, e := s.ExecAndClose(context.TODO(), db)
	if e != nil { // YES
		log.Printf("query error: %v\n", e)
		return 0, e
	}

	// How many entries deleted?
	c, e := r.RowsAffected()
	if e != nil { // YES
		log.Printf("query error: %v\n", e)
		return 0, e
	}
	return uint64(c), nil
}

func ObjectUserDelete(db *sql.DB, object uint64, user uint64) (bool, error) {
	// Create SQL Statement
	s := sqlf.DeleteFrom("registry_object_users").
		Where("id_user = ? and id_object= ?", user, object)

	// Execute
	_, e := s.ExecAndClose(context.TODO(), db)
	if e != nil { // YES
		log.Printf("query error: %v\n", e)
		return false, e
	}

	/* IGNORE RESULT - Assume that if no error than entry deleted
	  // How many rows deleted?
		c, e := r.RowsAffected()
		if e != nil { // YES
			log.Printf("query error: %v\n", e)
			return false, e
		}
	*/
	return true, nil
}

func ObjectUsersCount(db *sql.DB, object uint64, q query.TQueryConditions) (uint64, error) {
	// Query Results Values
	var count uint64

	// Create SQL Statement
	s := sqlf.From("registry_object_users").
		Select("COUNT(*)").To(&count).
		Where("id_object = ?", object)

	// Apply Query Conditions
	e := query.ApplyFilterConditions(s, q)

	// DEBUG: Print SQL
	fmt.Println(s.String())

	// Error Occurred?
	if e != nil { // YES
		log.Printf("query error: %v\n", e)
		return 0, e
	}

	// Execute Count
	e = s.QueryRowAndClose(context.TODO(), db)

	// Error Occurred?
	if e != nil { // YES
		log.Printf("query error: %v\n", e)
		return 0, e
	}

	return count, nil
}

func ObjectUsersRoleManagersCount(db *sql.DB, object uint64) (uint64, error) {
	// Query Results Values
	var count uint64

	// Create SQL Statement
	s := sqlf.From("registry_object_users").
		Select("COUNT(*)").To(&count).
		Where("id_object = ?", object).
		Where("mgr_Roles = ?", 1)

	// DEBUG: Print SQL
	fmt.Println(s.String())

	// Execute Count
	e := s.QueryRowAndClose(context.TODO(), db)

	// Error Occurred?
	if e != nil { // YES
		log.Printf("query error: %v\n", e)
		return 0, e
	}

	return count, nil
}

func ObjectUsersInviteManagersCount(db *sql.DB, object uint64) (uint64, error) {
	// Query Results Values
	var count uint64

	// Create SQL Statement
	s := sqlf.From("registry_object_users").
		Select("COUNT(*)").To(&count).
		Where("id_object = ?", object).
		Where("mgr_invites = ?", 1)

	// DEBUG: Print SQL
	fmt.Println(s.String())

	// Execute Count
	e := s.QueryRowAndClose(context.TODO(), db)

	// Error Occurred?
	if e != nil { // YES
		log.Printf("query error: %v\n", e)
		return 0, e
	}

	return count, nil
}

func ObjectUsersQuery(db *sql.DB, object uint64, q query.TQueryConditions, c bool) (query.TQueryResults, error) {
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
	var username string
	var state uint16
	var csv sql.NullString

	// Create SQL Statement
	s := sqlf.From("registry_object_users").
		Select("id_user").To(&id).
		Select("username").To(&username).
		Select("state").To(&state).
		Select("roles").To(&csv).
		Where("id_object = ?", object)

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

	// Apply Query Conditions
	e := query.ApplyFilterConditions(s, q)

	// Error Occurred?
	if e != nil { // YES
		log.Printf("query error: %v\n", e)
		return nil, e
	}

	// DEBUG: Print SQL
	fmt.Println(s.String())

	// Execute Query
	e = s.QueryAndClose(context.TODO(), db, func(row *sql.Rows) {
		userid := id
		e := ObjectUserRegistry{
			object:   &object,
			user:     &userid,
			username: username,
			state:    state,
		}

		// Do we have Roles Set?
		if csv.Valid { // YES
			e.RolesFromCSV(csv.String)
		}

		list.AppendValue(&e)
	})

	// Error Occurred?
	if e != nil && e != sql.ErrNoRows { // YES
		// DEBUG: Print SQL
		fmt.Print(s.String())

		log.Printf("query error: %v\n", e)
		return nil, e
	}

	// Is Count of Entries Requested?
	if c { // YES: Count Entries under Same Conditions
		count, e := ObjectUsersCount(db, object, q)
		if e != nil {
			return nil, e
		}

		list.SetMaxCount(count)
	}

	return list, nil
}

// IsDirty Have the Object Properties Changed since last Serialization?
func (o *ObjectUserRegistry) IsDirty() bool {
	return o.dirty
}

func (o *ObjectUserRegistry) IsNew() bool {
	return !o.stored
}

func (o *ObjectUserRegistry) IsValid() bool {
	return o.hasKey() && o.username != ""
}

func (o *ObjectUserRegistry) IsSystemOrganization() bool {
	return o.object != nil && *o.object == uint64(0x2000000000000)
}

func (o *ObjectUserRegistry) IsAdminUser() bool {
	return o.user != nil && o.HasAllStates(STATE_SYSTEM)
}

// ByID Finds Entry By Org / User
func (o *ObjectUserRegistry) ByKey(db *sql.DB, object uint64, user uint64) error {
	// Cleanup Entry
	o.reset()

	// Mark Entry as Clean
	o.dirty = false

	// Execute Query
	var ciphertext, csv sql.NullString
	e := sqlf.From("registry_object_users").
		Select("username").To(&o.username).
		Select("state").To(&o.state).
		Select("roles").To(&csv).
		Select("ciphertext").To(&ciphertext).
		Where("id_object = ? and id_user = ?", object, user).
		QueryRowAndClose(context.TODO(), db)

	// Error Executing Query?
	if e != nil && e != sql.ErrNoRows { // YES
		log.Printf("query error: %v\n", e)
		return e
	}

	// Did we retrieve an entry?
	if e == nil { // YES
		o.object = &object
		o.user = &user

		// Do we have Roles Set?
		if csv.Valid { // YES
			o.RolesFromCSV(csv.String)
			o.dirty = false // Make Sure Entry Marked Clean
		}

		// Do we have a cipher?
		if ciphertext.Valid { // YES: Save it
			s := ciphertext.String
			o.ciphertext = []byte(s)
		}

		o.stored = true // Registered Entry
	}

	return nil
}

// ID Get Organization ID
func (o *ObjectUserRegistry) Object() uint64 {
	return *o.object
}

func (o *ObjectUserRegistry) User() uint64 {
	return *o.user
}

func (o *ObjectUserRegistry) UserName() string {
	return o.username
}

func (o *ObjectUserRegistry) SetKey(object uint64, user uint64) error {
	if o.IsNew() {
		// New State
		o.object = &object
		o.user = &user
		o.dirty = true
		return nil
	}

	return errors.New("Entry KEY is immutable")
}

func (o *ObjectUserRegistry) SetUserName(username string) (string, error) {
	// Current State
	current := o.username

	// New State
	o.username = username
	o.dirty = true
	return current, nil
}

// Implementation of I_Roles //
func (o *ObjectUserRegistry) Roles() []uint32 {
	return o.roles.Roles()
}

func (o *ObjectUserRegistry) IsRolesEmpty() bool {
	return o.roles.IsRolesEmpty()
}

func (o *ObjectUserRegistry) HasRole(role uint32) bool {
	return o.roles.HasRole(role)
}

func (o *ObjectUserRegistry) HasExactRole(role uint32) bool {
	return o.roles.HasExactRole(role)
}

func (o *ObjectUserRegistry) GetCategoryRole(category uint16) uint32 {
	return o.roles.GetCategoryRole(category)
}

func (o *ObjectUserRegistry) GetSubCategoryRole(subcategory uint16) uint32 {
	return o.roles.GetSubCategoryRole(subcategory)
}

func (o *ObjectUserRegistry) AddRole(role uint32) bool {
	modified := o.roles.AddRole(role)
	o.dirty = o.dirty || modified
	return modified
}

func (o *ObjectUserRegistry) AddRoles(roles []uint32) bool {
	modified := o.roles.AddRoles(roles)
	o.dirty = o.dirty || modified
	return modified
}

func (o *ObjectUserRegistry) RemoveRole(role uint32) bool {
	modified := o.roles.RemoveRole(role)
	o.dirty = o.dirty || modified
	return modified
}

func (o *ObjectUserRegistry) RemoveCategory(category uint16) bool {
	modified := o.roles.RemoveCategory(category)
	o.dirty = o.dirty || modified
	return modified
}

func (o *ObjectUserRegistry) RemoveExactRole(role uint32) bool {
	modified := o.roles.RemoveExactRole(role)
	o.dirty = o.dirty || modified
	return modified
}

func (o *ObjectUserRegistry) RemoveRoles(roles []uint32) bool {
	modified := o.roles.RemoveRoles(roles)
	o.dirty = o.dirty || modified
	return modified
}

func (o *ObjectUserRegistry) RemoveAllRoles() bool {
	modified := o.roles.RemoveAllRoles()
	o.dirty = o.dirty || modified
	return modified
}

func (o *ObjectUserRegistry) RolesFromCSV(csv string) bool {
	modified := o.roles.RolesFromCSV(csv)
	o.dirty = o.dirty || modified
	return modified
}

func (o *ObjectUserRegistry) RolesToCSV() string {
	return o.roles.RolesToCSV()
}

func (o *ObjectUserRegistry) IsRolesManager() bool {
	// ROLES Manager Requires User Read/List/Update
	r := o.roles.GetSubCategoryRole(SUBCATEGORY_USER)
	b := RoleMatchFunctions(FUNCTION_READ|FUNCTION_LIST, r)

	// Has Required User Permissions?
	if b { // YES: See if has Required Roles Permissions
		r = o.roles.GetSubCategoryRole(SUBCATEGORY_ROLES)
		b = b && RoleMatchFunctions(FUNCTION_READ|FUNCTION_UPDATE, r)
	}
	return b
}

func (o *ObjectUserRegistry) IsInvitationManager() bool {
	// ROLES Manager Requires User Read/List/Update
	r := o.roles.GetSubCategoryRole(SUBCATEGORY_USER)
	b := RoleMatchFunctions(FUNCTION_READ|FUNCTION_LIST, r)

	// Has Required User Permissions?
	if b { // YES: See if has Required Roles Permissions
		r = o.roles.GetSubCategoryRole(SUBCATEGORY_INVITE)
		b = b && RoleMatchFunctions(FUNCTION_READ|FUNCTION_CREATE|FUNCTION_DELETE, r)
	}
	return b
}

func (o *ObjectUserRegistry) State() uint16 {
	return o.state
}

func (o *ObjectUserRegistry) HasAnyStates(states uint16) bool {
	return HasAnyStates(o.state, states)
}

func (o *ObjectUserRegistry) HasAllStates(states uint16) bool {
	return HasAllStates(o.state, states)
}

func (o *ObjectUserRegistry) IsActive() bool {
	// User Account Active
	return !HasAnyStates(o.state, STATE_INACTIVE|STATE_BLOCKED|STATE_DELETE)
}

func (o *ObjectUserRegistry) IsDeleted() bool {
	// GLOBAL User marked for Deletion
	return HasAllStates(o.state, STATE_DELETE)
}

func (o *ObjectUserRegistry) IsBlocked() bool {
	// GLOBAL User Access Blocked
	return HasAnyStates(o.state, STATE_BLOCKED|STATE_DELETE)
}

func (o *ObjectUserRegistry) IsReadOnly() bool {
	return HasAllStates(o.state, STATE_READONLY)
}

func (o *ObjectUserRegistry) SetStates(states uint16) {
	// Current State
	current := o.state

	// New State
	o.state = SetStates(o.state, states)
	if o.state != current {
		o.dirty = true
	}
}

func (o *ObjectUserRegistry) ClearStates(states uint16) {
	// Current State
	current := o.state

	// New State
	o.state = ClearStates(o.state, states)
	if o.state != current {
		o.dirty = true
	}
}

// STORE KEY
func (o *ObjectUserRegistry) StoreKey(hexCypher string) ([]byte, error) {
	// Convert User Hash to byte Array
	key, e := hex.DecodeString(hexCypher)
	if e != nil {
		return nil, e
	}

	return o.StoreKeyBytes(key)
}

func (o *ObjectUserRegistry) StoreKeyBytes(key []byte) ([]byte, error) {
	// Decrypt Store Key
	plainbytes, e := toPlainBytes(key, o.ciphertext)
	if e != nil {
		return nil, e
	}

	return plainbytes, nil
}

func (o *ObjectUserRegistry) SetStoreKey(hexCypher string, key []byte) error {
	if o.IsNew() {
		// Convert User Hash to byte Array
		cypher, e := hex.DecodeString(hexCypher)
		if e != nil {
			return e
		}

		// Encrypt Store Key using User Password Hash
		cypherbytes, e := toCypherBytes(cypher, key)
		if e != nil {
			return e
		}

		o.ciphertext = cypherbytes
		o.dirty = true
		return nil
	}

	return errors.New("Store KEY is immutable")
}

func (o *ObjectUserRegistry) CreateStoreKey(hexCypher string) error {
	// Create a Random Store Key
	key, e := o.generateCipherText()
	if e != nil {
		return e
	}

	return o.SetStoreKey(hexCypher, key)
}

func (o *ObjectUserRegistry) Flush(db sqlf.Executor, force bool) error {
	// Have DB Connection?
	if db == nil { // NO: Abort
		return errors.New("Missing Database Connection")
	}

	// Has entry been modified?
	if !force && !o.IsDirty() { // NO: Abort
		return nil
	}

	// Is New Entry?
	var s *sqlf.Stmt
	if o.IsNew() { // YES: Create
		if !o.IsValid() {
			return errors.New("Invalid Registry Entry")
		}

		s = sqlf.InsertInto("registry_object_users").
			Set("id_object", o.object).
			Set("id_user", o.user).
			Set("username", o.username).
			Set("state", o.state)

		if o.ciphertext != nil {
			s.Set("ciphertext", o.ciphertext)
		}
	} else { // NO: Update
		if !o.hasKey() {
			return errors.New("Missing or Invalid Registry Key")
		}

		// Create SQL Statement
		s = sqlf.Update("registry_object_users").
			Set("state", o.state).
			Where("id_object = ? and id_user = ?", o.object, o.user)

		// Is User Name Set?
		if o.username != "" { // YES: Update that as well
			s.Set("username", o.username)
		}

		// Is Cipher Changed?
		if o.ciphertext != nil {
			s.Set("ciphertext", o.ciphertext)
		}
	}

	// Do we have Roles to Set?
	if !o.IsRolesEmpty() { // YES
		s.Set("roles", o.RolesToCSV())

		// Has Permissions of a Roles Manager?
		if o.IsRolesManager() { // YES
			s.Set("mgr_roles", 1)
		} else { // NO
			s.Set("mgr_roles", 0)
		}

		// Has Permissions of an Invitation Manager?
		if o.IsInvitationManager() { // YES
			s.Set("mgr_invites", 1)
		} else { // NO
			s.Set("mgr_invites", 0)
		}
	} else { // Roles Empty
		s.Set("roles", nil)
		s.Set("mgr_roles", 0)
		s.Set("mgr_invites", 0)
	}

	// Execute Statement
	_, e := s.ExecAndClose(context.TODO(), db)
	if e == nil {
		o.stored = true
		o.dirty = false
	}
	return e
}

func (o *ObjectUserRegistry) hasKey() bool {
	return o.object != nil && o.user != nil
}

func (o *ObjectUserRegistry) reset() {
	// Clean Entry
	o.object = nil
	o.user = nil
	o.username = ""
	o.state = 0
	o.ciphertext = nil
	o.RemoveAllRoles()

	// Mark State as Unregistered
	o.stored = false

	// Mark Entry as Clean
	o.dirty = false
}

func (o *ObjectUserRegistry) generateCipherText() ([]byte, error) {
	if o.user == nil || o.object == nil {
		return nil, errors.New("Not Ready")
	}

	// Generate a Pseudo Random String to Hash
	rs := RandomAlphaNumericPunctuationString(128) // Random String to make things harder
	hashString := fmt.Sprintf("%d:%s:%d", o.user, rs, o.object)
	hash := sha256.Sum256([]byte(hashString))

	// Create HASH of Plain Text
	return hash[:], nil
}
