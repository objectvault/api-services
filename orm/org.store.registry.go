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

// Org Store Registry Definition
type OrgStoreRegistry struct {
	States
	dirty       bool    // Is Entry Dirty?
	stored      bool    // Is Entry Stored in Database
	org         *uint64 // KEY: GLOBAL Organization ID
	store       *uint64 // KEY: GLOBAL User ID
	store_alias string  // Store Alias
	state       uint16  // Store State in Organization
}

// TODO Implement Delete (Both From Within an Entry and Without a Structure)

func CountRegisteredOrgStores(db *sql.DB, org uint64, q query.TQueryConditions) (uint64, error) {
	// Query Results Values
	var count uint64

	// Create SQL Statement
	s := sqlf.From("registry_org_stores").
		Select("COUNT(*)").To(&count).
		Where("id_org = ?", org)

	// Execute Count
	e := s.QueryRowAndClose(context.TODO(), db)

	// Error Occurred?
	if e != nil { // YES
		log.Printf("query error: %v\n", e)
		return 0, e
	}

	return count, nil
}

func QueryRegisteredStores(db *sql.DB, org uint64, q query.TQueryConditions, c bool) (query.TQueryResults, error) {
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

	// Create SQL Statement
	s := sqlf.From("registry_org_stores").
		Select("id_store").To(&id).
		Select("storename").To(&alias).
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
		list.AppendSort("id_store", false)
		s.OrderBy("id_store")
	}

	// Apply Organization Filter
	s.Where("id_org = ?", org)

	// DEBUG: Print SQL
	fmt.Print(s.String())

	// Execute Query
	e := s.QueryAndClose(context.TODO(), db, func(row *sql.Rows) {
		storeid := id
		o := OrgStoreRegistry{
			org:         &org,
			store:       &storeid,
			store_alias: alias,
			state:       state,
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
		count, e := CountRegisteredOrgStores(db, org, q)
		if e != nil {
			return nil, e
		}

		list.SetMaxCount(count)
	}

	return list, nil
}

// General Find Store in Org
func (o *OrgStoreRegistry) Find(db *sql.DB, org uint64, id interface{}) error {
	// is ID an integer?
	if _, ok := id.(uint64); ok { // YES: Find by ID
		return o.ByID(db, org, id.(uint64))
	} else if _, ok := id.(string); ok { // ELSE: Find by Alias
		return o.ByAlias(db, org, id.(string))
	}
	// ELSE: Missing or Invalid id
	return errors.New("'id' missing or of invalid type")
}

// ByID Find Store in Organization by ID
func (o *OrgStoreRegistry) ByID(db *sql.DB, org uint64, store uint64) error {
	// Cleanup Entry
	o.reset()

	// Execute Query
	e := sqlf.From("registry_org_stores").
		Select("storename").To(&o.store_alias).
		Select("state").To(&o.state).
		Where("id_org = ? and id_store = ?", org, store).
		QueryRowAndClose(context.TODO(), db)

	// Error Executing Query?
	if e != nil && e != sql.ErrNoRows { // YES
		log.Printf("query error: %v\n", e)
		return e
	}

	// Did we retrieve an entry?
	if e == nil { // YES
		o.org = &org
		o.store = &store
		o.dirty = false // Clear Dirty Flag
		o.stored = true // Registered Store
	}

	return nil
}

// ByAlias Find Store in Organization by Alias
func (o *OrgStoreRegistry) ByAlias(db *sql.DB, org uint64, store string) error {
	store = strings.TrimSpace(store)
	if store == "" {
		return errors.New("Missing Value for Required Parameter 'store'")
	}

	// Cleanup Entry
	o.reset()

	// Execute Query
	e := sqlf.From("registry_org_stores").
		Select("id_store").To(&o.store).
		Select("state").To(&o.state).
		Where("id_org = ? and storename = ?", org, store).
		QueryRowAndClose(context.TODO(), db)

	// Error Executing Query?
	if e != nil && e != sql.ErrNoRows { // YES
		log.Printf("query error: %v\n", e)
		return e
	}

	// Did we retrieve an entry?
	if e == nil { // YES
		o.org = &org
		o.store_alias = store
		o.dirty = false // Clear Dirty Flag
		o.stored = true // Registered Store
	}

	return nil
}

// IsDirty Have the Object Properties Changed since last Serialization?
func (o *OrgStoreRegistry) IsDirty() bool {
	return o.dirty
}

func (o *OrgStoreRegistry) IsNew() bool {
	return !o.stored
}

func (o *OrgStoreRegistry) IsValid() bool {
	return o.hasKey() && o.store_alias != ""
}

// ID Get Organization ID
func (o *OrgStoreRegistry) Organization() uint64 {
	return *o.org
}

func (o *OrgStoreRegistry) Store() uint64 {
	return *o.store
}

func (o *OrgStoreRegistry) StoreAlias() string {
	return o.store_alias
}

func (o *OrgStoreRegistry) State() uint16 {
	return o.state
}

func (o *OrgStoreRegistry) HasAnyStates(states uint16) bool {
	return HasAnyStates(o.state, states)
}

func (o *OrgStoreRegistry) HasAllStates(states uint16) bool {
	return HasAllStates(o.state, states)
}

func (o *OrgStoreRegistry) IsBlocked() bool {
	// GLOBAL Org Access Blocked
	return HasAnyStates(o.state, STATE_INACTIVE|STATE_BLOCKED)
}

func (o *OrgStoreRegistry) IsReadOnly() bool {
	return HasAnyStates(o.state, STATE_READONLY)
}

func (o *OrgStoreRegistry) SetKey(org uint64, user uint64) error {
	if o.IsNew() {
		// New State
		o.org = &org
		o.store = &user
		o.dirty = true
		return nil
	}

	return errors.New("Entry KEY is immutable")
}

func (o *OrgStoreRegistry) SetStoreAlias(alias string) (string, error) {
	// Current State
	current := o.store_alias

	// Validate Organization Alias
	if alias == "" {
		return current, errors.New("Missing Value for Organization Alias")
	}

	if alias != current {
		// New State
		o.store_alias = alias
		o.dirty = true
	}
	return current, nil
}

func (o *OrgStoreRegistry) SetStates(states uint16) (uint16, error) {
	// Current State
	current := o.state

	// New State
	o.state = SetStates(o.state, states)
	if o.state != current {
		o.dirty = true
	}
	return current, nil
}

func (o *OrgStoreRegistry) ClearStates(states uint16) (uint16, error) {
	// Current State
	current := o.state

	// New State
	o.state = ClearStates(o.state, states)
	if o.state != current {
		o.dirty = true
	}
	return current, nil
}

func (o *OrgStoreRegistry) Flush(db sqlf.Executor, force bool) error {
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
		_, e = sqlf.InsertInto("registry_org_stores").
			Set("id_org", o.org).
			Set("id_store", o.store).
			Set("storename", o.store_alias).
			Set("state", o.state).
			ExecAndClose(context.TODO(), db)
	} else { // NO: Update
		if !o.hasKey() {
			return errors.New("Missing or Invalid Registry Key")
		}

		// Create SQL Statement
		s := sqlf.Update("registry_org_stores").
			Set("state", o.state).
			Where("id_org = ? and id_store = ?", o.org, o.store)

		// Is Store Alias Set?
		if o.store_alias != "" { // YES: Update that as well
			s.Set("storename", o.store_alias)
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

func (o *OrgStoreRegistry) hasKey() bool {
	return o.org != nil && o.store != nil
}

func (o *OrgStoreRegistry) reset() {
	// Clean Entry
	o.org = nil
	o.store = nil
	o.store_alias = ""
	o.state = 0

	// Mark State as Unregistered
	o.stored = false

	// Mark Entry as Clean
	o.dirty = false
}
