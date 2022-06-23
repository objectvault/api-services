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

// Organization Object Definition
type Organization struct {
	dirty          bool       // Is Entry Dirty?
	updateRegistry bool       // Do we need to Update the Registry?
	stored         bool       // Is Entry Stored in Database
	id             *uint32    // LOCAL Organization ID
	alias          string     // Organization Alias
	name           *string    // Organization Name (Can be NULL)
	creator        *uint64    // Global User ID of Creator
	created        *time.Time // Created TimeStamp
	modifier       *uint64    // Global User ID of Last Modifier
	modified       *time.Time // Modification TimeStamp
}

// TODO Implement Delete (Both From Within an Entry and Without a Structure)

// IsDirty Have the Object Properties Changed since last Serialization?
func (o *Organization) IsDirty() bool {
	return o.dirty
}

func (o *Organization) UpdateRegistry() bool {
	return o.dirty && o.updateRegistry
}

func (o *Organization) IsNew() bool {
	return !o.stored
}

func (o *Organization) IsValid() bool {
	return o.id != nil && o.alias != ""
}

func (o *Organization) IsSystem() bool {
	return o.id != nil && *o.id == uint32(0)
}

func (o *Organization) Find(db *sql.DB, id interface{}) error {
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

// ByID Finds User By ID
func (o *Organization) ByID(db *sql.DB, id uint32) error {
	// Reset Entry
	o.reset()

	// Execute Query
	// TODO Process Database "object" field
	var name sql.NullString
	var object sql.NullString
	var created sql.NullString
	var modifier sql.NullInt64
	var modified sql.NullString
	e := sqlf.From("orgs").
		Select("orgname").To(&o.alias).
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

// ByOrgName Finds Organization By Name
func (o *Organization) ByAlias(db *sql.DB, alias string) error {
	// Validate Alias
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return errors.New("Missing Value for Required Parameter 'alias'")
	}

	// Reset Entry
	o.reset()

	// Execute Query
	// TODO Process Database "object" field
	var name sql.NullString
	var object sql.NullString
	var creator sql.NullInt32
	e := sqlf.From("orgs").
		Select("id").To(o.id).
		Select("name").To(&name).
		Select("creator").To(&creator).
		Select("object").To(&object).
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
		// TODO Deal with Object, Creator/ed, Modifier/de
		o.stored = true
	}

	return nil
}

// ID Get User ID
func (o *Organization) ID() uint32 {
	if o.id == nil {
		return 0
	}

	return *o.id
}

func (o *Organization) Alias() string {
	return o.alias
}

func (o *Organization) Name() string {
	if o.name == nil {
		return ""
	}
	return *o.name
}

func (o *Organization) Creator() uint64 {
	if o.creator == nil {
		return 0
	}
	return *o.creator
}

func (o *Organization) Created() *time.Time {
	return o.created
}

func (o *Organization) Modifier() *uint64 {
	if o.modifier == nil {
		return nil
	}
	return o.modifier
}

func (o *Organization) Modified() *time.Time {
	return o.modified
}

func (o *Organization) SetID(id uint32) (uint32, error) {
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

	return 0, errors.New("Registered Organization - ID is immutable")
}

func (o *Organization) SetAlias(alias string) (string, error) {
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
		o.updateRegistry = true // Need to Update Registry
	}

	return current, nil
}

func (o *Organization) SetName(name string) (string, error) {
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
	}

	return current, nil
}

func (o *Organization) SetCreator(id uint64) error {
	// Is Record New?
	if o.IsNew() { // YES
		o.creator = &id

		// Creation Time Stamp AUTO SET by MySQL
		return nil
	}

	return errors.New("Registered Organization - Creator Cannot be Changed")
}

func (o *Organization) SetModifier(id uint64) error {
	// Set Modifier
	o.modifier = &id

	// Modification Time Stamp AUTO SET by MySQL
	return nil
}

func (o *Organization) Flush(db sqlf.Executor, force bool) error {
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
		_, e = sqlf.InsertInto("orgs").
			Set("orgname", o.alias).
			Set("name", o.name).
			Set("creator", o.creator).
			Exec(context.TODO(), db)

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
	} else { // NO: Update
		if o.id == nil {
			return errors.New("Organization ID not Set")
		}

		if o.modifier == nil {
			return errors.New("Modification User not Set")
		}

		_, e = sqlf.Update("orgs").
			Set("orgname", o.alias).
			Set("name", o.name).
			Set("modifier", o.modifier).
			Where("id = ?", o.id).
			ExecAndClose(context.TODO(), db)
	}

	if e == nil {
		o.stored = true
		o.dirty = false
	}
	return e
}

func (o *Organization) reset() {
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
