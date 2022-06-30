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
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/objectvault/api-services/orm/query"
	"github.com/pjacferreira/sqlf"
)

// User Object Definition
type OrgRegistry struct {
	States
	dirty  bool    // Is Entry Dirty?
	stored bool    // Is Entry Stored in Database
	id     *uint64 // KEY: GLOBAL Organization ID
	alias  string  // Organization Alias
	name   *string // Organization Name (Can be NIL)
	state  uint16  // GLOBAL Org State
}

// TODO Implement Delete (Both From Within an Entry and Without a Structure)
// TODO VERIFY if Organization Long Name 'name' should be part of registry entry

func CountRegisteredOrgs(db *sql.DB, q query.TQueryConditions) (uint64, error) {
	// Query Results Values
	var count uint64

	// Create SQL Statement
	s := sqlf.From("registry_orgs").
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

func QueryRegisteredOrgs(db *sql.DB, q query.TQueryConditions, c bool) (query.TQueryResults, error) {
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
	var alias string
	var state uint16
	var name sql.NullString

	// Create SQL Statement
	s := sqlf.From("registry_orgs").
		Select("id_org").To(&id).
		Select("orgname").To(&alias).
		Select("name").To(&name).
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
		list.AppendSort("id_org", false)
		s.OrderBy("id_org")
	}

	// DEBUG: Print SQL
	fmt.Print(s.String())

	// Execute Query
	e := s.QueryAndClose(context.TODO(), db, func(row *sql.Rows) {
		orgid := id

		o := OrgRegistry{
			id:    &orgid,
			alias: alias,
			state: state,
		}

		if name.Valid {
			o.name = &name.String
		}

		list.AppendValue(&o)
	})

	// Error Occurred?
	if e != nil && e != sql.ErrNoRows { // YES
		log.Printf("query error: %v\n", e)
		return nil, e
	}

	// Is Count of Entries Requested?
	if c { // YES: Count Entries under Same Conditions
		count, e := CountRegisteredOrgs(db, q)
		if e != nil {
			return nil, e
		}

		list.SetMaxCount(count)
	}

	return list, nil
}

// IsDirty Have the Object Properties Changed since last Serialization?
func (o *OrgRegistry) IsDirty() bool {
	return o.dirty
}

func (o *OrgRegistry) IsNew() bool {
	return !o.stored
}

func (o *OrgRegistry) IsValid() bool {
	return o.hasKey() && o.alias != ""
}

func (o *OrgRegistry) IsSystem() bool {
	return o.id != nil && *o.id == uint64(0x2000000000000)
}

func (o *OrgRegistry) Find(db *sql.DB, id interface{}) error {
	// is ID an integer?
	if _, ok := id.(uint64); ok { // YES: Find by ID
		return o.ByID(db, id.(uint64))
	} else if _, ok := id.(string); ok { // ELSE: Find by Alias
		return o.ByAlias(db, id.(string))
	}
	// ELSE: Missing or Invalid id
	return errors.New("'id' missing or of invalid type")
}

// ByID Finds User By ID
func (o *OrgRegistry) ByID(db *sql.DB, id uint64) error {
	// Cleanup Entry
	o.reset()

	// Execute Query
	var name sql.NullString
	e := sqlf.From("registry_orgs").
		Select("orgname").To(&o.alias).
		Select("name").To(&name).
		Select("state").To(&o.state).
		Where("id_org = ?", id).
		QueryRowAndClose(context.TODO(), db)

	// Error Executing Query?
	if e != nil && e != sql.ErrNoRows { // YES
		log.Printf("query error: %v\n", e)
		return e
	}

	// Did we retrieve an entry?
	if e == nil { // YES
		o.id = &id
		if name.Valid {
			o.name = &name.String
		}
		o.stored = true // Registered User
	}

	return nil
}

// ByOrgName Finds Org By Alias
func (o *OrgRegistry) ByAlias(db *sql.DB, alias string) error {
	// Validate Alias
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return errors.New("Missing Value for Required Parameter 'alias'")
	}

	// Cleanup Entry
	o.reset()

	// Is Incoming Parameter Valid?
	if alias == "" { // NO
		return errors.New("Missing Required Parameter 'alias'")
	}

	// Mark State as Unregistered
	o.stored = false

	// Mark Entry as Clean
	o.dirty = false

	// Execute Query
	var name sql.NullString
	e := sqlf.From("registry_orgs").
		Select("id_org").To(&o.id).
		Select("name").To(&name).
		Select("state").To(&o.state).
		Where("orgname = ?", alias).
		QueryRowAndClose(context.TODO(), db)

		// Error Executing Query?
	if e != nil && e != sql.ErrNoRows { // YES
		log.Printf("query error: %v\n", e)
		return e
	}

	// Did we retrieve an entry?
	if e == nil { // YES
		o.alias = alias
		if name.Valid {
			o.name = &name.String
		}
		o.stored = true // Registered Organiztation
	}

	return nil
}

func (o *OrgRegistry) ID() uint64 {
	if o.id == nil {
		return 0
	}

	return *o.id
}

func (o *OrgRegistry) Alias() string {
	return o.alias
}

func (o *OrgRegistry) Name() string {
	if o.name == nil {
		return ""
	}

	return *o.name
}

func (o *OrgRegistry) State() uint16 {
	return o.state
}

func (o *OrgRegistry) HasAnyStates(states uint16) bool {
	return HasAnyStates(o.state, states)
}

func (o *OrgRegistry) HasAllStates(states uint16) bool {
	return HasAllStates(o.state, states)
}

func (o *OrgRegistry) IsActive() bool {
	// User Account Active
	return !HasAllStates(o.state, STATE_INACTIVE)
}

func (o *OrgRegistry) IsBlocked() bool {
	// GLOBAL Org Access Blocked
	return HasAnyStates(o.state, STATE_INACTIVE|STATE_BLOCKED)
}

func (o *OrgRegistry) IsReadOnly() bool {
	return HasAllStates(o.state, STATE_READONLY)
}

func (o *OrgRegistry) SetID(id uint64) (uint64, error) {
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

func (o *OrgRegistry) SetAlias(alias string) (string, error) {
	// Current State
	current := o.alias

	// Validate Organization Alias
	if alias == "" {
		return current, errors.New("Missing Value for Organization Alias")
	}

	if alias != current {
		// New State
		o.alias = alias
		o.dirty = true
	}
	return current, nil
}

func (o *OrgRegistry) SetName(name string) (string, error) {
	// Current State
	var current string
	if o.name == nil {
		current = ""
	} else {
		current = *o.name
	}

	// New State (DO NOT TRIM HERE)
	if name == "" {
		o.name = nil
	} else {
		o.name = &name
	}
	o.dirty = true
	return current, nil
}

func (o *OrgRegistry) SetStates(states uint16) (uint16, error) {
	// Current State
	current := o.state

	// New State
	o.state = SetStates(o.state, states)
	if o.state != current {
		o.dirty = true
	}
	return current, nil
}

func (o *OrgRegistry) ClearStates(states uint16) (uint16, error) {
	// Current State
	current := o.state

	// New State
	o.state = ClearStates(o.state, states)
	if o.state != current {
		o.dirty = true
	}
	return current, nil
}

func (o *OrgRegistry) Flush(db sqlf.Executor, force bool) error {
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
		_, e = sqlf.InsertInto("registry_orgs").
			Set("id_org", o.id).
			Set("orgname", o.alias).
			Set("name", o.name).
			Set("state", o.state).
			ExecAndClose(context.TODO(), db)
	} else { // NO: Update
		if !o.hasKey() {
			return errors.New("Missing or Invalid Registry Key")
		}

		// Create SQL Statement
		s := sqlf.Update("registry_orgs").
			Set("name", o.name).
			Set("state", o.state).
			Where("id_org = ?", o.id)

		// Is Organization Alias Set?
		if o.alias != "" { // YES: Update that as well
			s.Set("orgname", o.alias)
		}

		// Execute Statement
		_, e = s.ExecAndClose(context.TODO(), db)
	}

	if e == nil {
		o.stored = true
		o.dirty = false
	}
	return e
}

func (o *OrgRegistry) hasKey() bool {
	return o.id != nil
}

func (o *OrgRegistry) reset() {
	// Clean Entry
	o.id = nil
	o.alias = ""
	o.name = nil
	o.state = 0

	// Mark State as Unregistered
	o.stored = false

	// Mark Entry as Clean
	o.dirty = false
}
