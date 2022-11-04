// cSpell:ignore ferreira, paulo
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

// cSpell:ignore storename

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/objectvault/api-services/orm/mysql"
	"github.com/pjacferreira/sqlf"
)

func GetShardStoreID(db sqlf.Executor, org uint64, alias string) (uint32, error) {
	// Query Results Values
	var id uint32

	// Create SQL Statement
	e := sqlf.From("stores").
		Select("id").To(&id).
		Where("id_org = ?", org).
		Where("storename = ?", alias).
		QueryRowAndClose(context.TODO(), db)

		// Error Executing Query?
	if e != nil && e != sql.ErrNoRows { // YES
		log.Printf("query error: %v\n", e)
		return 0, e
	}

	return id, nil
}

// Store Object Definition
type Store struct {
	dirty          bool       // Is Entry Dirty?
	updateRegistry bool       // Do we need to Update the Registry?
	stored         bool       // Is Entry Stored in Database
	id             *uint32    // LOCAL Store ID
	org            *uint64    // Global Organization ID store Belongs To
	alias          string     // Store Alias
	name           *string    // Store Name (Can be NULL)
	creator        *uint64    // Global User ID of Creator
	created        *time.Time // Created TimeStamp
	modifier       *uint64    // Global User ID of Last Modifier
	modified       *time.Time // Modification TimeStamp
}

// TODO Implement Delete (Both From Within an Entry and Without a Structure)

// IsDirty Have the Object Properties Changed since last Serialization?
func (o *Store) IsDirty() bool {
	return o.dirty
}

func (o *Store) UpdateRegistry() bool {
	return o.dirty && o.updateRegistry
}

func (o *Store) IsNew() bool {
	return !o.stored
}

func (o *Store) IsValid() bool {
	return o.hasKey() && o.org != nil && o.alias != ""
}

func (o *Store) Find(db *sql.DB, id interface{}) error {
	// is ID an integer?
	if _, ok := id.(uint64); ok { // YES: Find by ID
		oid := id.(uint64)
		return o.ByID(db, uint32(oid))
	} else if _, ok := id.(string); ok { // ELSE: Find by Alias
		return o.ByAlias(db, id.(string))
	}
	// ELSE: Missing or Invalid id
	return errors.New("'id' missing or of invalid type")
}

// ByID Finds Store By ID
func (o *Store) ByID(db *sql.DB, id uint32) error {
	// Reset Entry
	o.reset()

	// Execute Query
	// TODO Process Database "object" field
	var name sql.NullString
	var object sql.NullString
	var created sql.NullString
	var modifier sql.NullInt64
	var modified sql.NullString
	e := sqlf.From("stores").
		Select("id_org").To(&o.org).
		Select("storename").To(&o.alias).
		Select("name").To(&name).
		Select("object").To(&object).
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
		if name.Valid {
			o.name = &name.String
		}
		if created.Valid {
			o.created = mysql.MySQLTimeStampToGoTime(created.String)
		}
		if modifier.Valid {
			m := uint64(modifier.Int64)
			o.modifier = &m
			if modified.Valid {
				o.modified = mysql.MySQLTimeStampToGoTime(created.String)
			}
		}
		// TODO Deal with Object
		o.stored = true
	}

	return nil
}

// ByOrgName Finds Store By Name
func (o *Store) ByAlias(db *sql.DB, alias string) error {
	// Validate Alias
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return errors.New("Missing Value for Required Parameter 'alias'")
	}

	// Reset Entry
	o.reset()

	// Is Incoming Parameter Valid?
	if alias == "" { // NO
		return errors.New("Missing Required Parameter 'storename'")
	}

	// Execute Query
	// TODO Process Database "object" field
	var name sql.NullString
	var object sql.NullString
	var creator sql.NullInt32
	e := sqlf.From("stores").
		Select("id").To(o.id).
		Select("name").To(&name).
		Select("creator").To(&creator).
		Select("object").To(&object).
		Where("storename = ?", alias).
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
		// TODO Deal with Object, Creator/ed, Modifier/de
		o.stored = true
	}

	return nil
}

// ID Local ID
func (o *Store) ID() uint32 {
	if o.id == nil {
		return 0
	}

	return *o.id
}

// ID Store Parent Organization
func (o *Store) Organization() uint64 {
	if o.org == nil {
		return 0
	}

	return *o.org
}

func (o *Store) Alias() string {
	return o.alias
}

func (o *Store) Name() string {
	if o.name == nil {
		return ""
	}
	return *o.name
}

func (o *Store) Creator() uint64 {
	if o.creator == nil {
		return 0
	}
	return *o.creator
}

func (o *Store) Created() *time.Time {
	return o.created
}

func (o *Store) Modifier() *uint64 {
	if o.modifier == nil {
		return nil
	}
	return o.modifier
}

func (o *Store) Modified() *time.Time {
	return o.modified
}

func (o *Store) SetID(id uint32) (uint32, error) {
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

	return 0, errors.New("Registered Store - ID is immutable")
}

func (o *Store) SetOrganization(org uint64) (uint64, error) {
	if o.IsNew() {
		// Current State
		current := uint64(0)
		if o.org != nil {
			current = *o.org
		}

		// New State
		o.org = &org
		o.dirty = true
		return current, nil
	}

	return 0, errors.New("Registered Store - Organization is immutable")
}

func (o *Store) SetAlias(alias string) (string, error) {
	// Current State
	current := o.alias

	// Validate Store Alias
	if alias == "" {
		return current, errors.New("Missing Value for Store Alias")
	}

	if alias != current {
		// New State
		o.alias = alias
		o.dirty = true
		o.updateRegistry = true // Need to Update Registry
	}

	return current, nil
}

func (o *Store) SetName(name string) (string, error) {
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

	// Name Changed?
	if current != name { // YES - Entry Modified and Registry Needs to be Updated
		o.dirty = true
		o.updateRegistry = true // Need to Update Registry
	}

	return current, nil
}

func (o *Store) SetCreator(id uint64) error {
	// Is Record New?
	if o.IsNew() { // YES
		o.creator = &id

		// TODO: Set Creation Time Stamp
		return nil
	}

	return errors.New("Registered Store - Creator Cannot be Changed")
}

func (o *Store) SetModifier(id uint64) error {
	// Set Modifier
	o.modifier = &id

	// TODO: Set Modification Time Stamp
	return nil
}

func (o *Store) Flush(db sqlf.Executor, force bool) error {
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
		if o.creator == nil {
			return errors.New("Creation User not Set")
		}

		// Execute Insert
		_, e = sqlf.InsertInto("stores").
			Set("id_org", o.org).
			Set("storename", o.alias).
			Set("name", o.name).
			Set("creator", o.creator).
			ExecAndClose(context.TODO(), db)

		// Error Occurred?
		if e == nil { // NO: Get New Store's ID
			// Error Occurred?
			id, e := GetShardStoreID(db, *o.org, o.alias)
			if e == nil { // NO: Set Object ID
				o.id = &id
			}
		}
	} else { // NO: Update
		if o.hasKey() {
			return errors.New("Missing or Invalid Store Key")
		}

		if o.modifier == nil {
			return errors.New("Modification User not Set")
		}

		// Create SQL Statement
		s := sqlf.Update("stores").
			Set("name", o.name).
			Set("modifier", o.modifier).
			Where("id = ?", o.id)

		// Is Store Alias Set?
		if o.alias != "" { // YES: Update that as well
			s.Set("storename", o.alias)
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

func (o *Store) hasKey() bool {
	return o.id != nil
}

func (o *Store) reset() {
	// Clean Entry
	o.id = nil
	o.alias = ""
	o.name = nil
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
