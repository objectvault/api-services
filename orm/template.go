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

// Template Object Definition
type Template struct {
	dirty       bool       // Is Entry Dirty?
	stored      bool       // Is Entry Stored in Database
	id          *uint32    // LOCAL Template ID
	name        string     // Template Name
	version     uint16     // Template Version
	title       string     // Template Title
	description string     // Template Description
	model       string     // Template JSON Model
	created     *time.Time // Created TimeStamp
}

// TODO Implement Delete (Both From Within an Entry and Without a Structure)

// IsDirty Have the Object Properties Changed since last Serialization?
func (o *Template) IsDirty() bool {
	return o.dirty
}

func (o *Template) IsNew() bool {
	return !o.stored
}

func (o *Template) IsValid() bool {
	return o.hasKey() && o.name != "" && o.title != "" && o.version > 0 && o.model != ""
}

// ByID Finds Template By ID
func (o *Template) ByID(db *sql.DB, id uint32) error {
	// Reset Entry
	o.reset()

	// Execute Query
	var description sql.NullString
	var created sql.NullString
	e := sqlf.From("templates").
		Select("name").To(&o.name).
		Select("version").To(&o.version).
		Select("title").To(&o.title).
		Select("description").To(&description).
		Select("model").To(&o.model).
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

		if description.Valid {
			o.description = description.String
		}

		if created.Valid {
			o.created = mysql.MySQLTimeStampToGoTime(created.String)
		}

		// TODO Deal with Object
		o.stored = true
	}

	return nil
}

// Finds Latest Version ofTemplate By Name
func (o *Template) ByNameLatest(db *sql.DB, name string) error {
	// Validate Template
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("Missing Value for Required Parameter 'alias'")
	}
	name = strings.ToLower(name)

	// Execute Query
	var description sql.NullString
	var created sql.NullString
	e := sqlf.From("templates").
		Select("id").To(&o.id).
		Select("version").To(&o.version).
		Select("title").To(&o.title).
		Select("description").To(&description).
		Select("model").To(&o.model).
		Select("created").To(&created).
		Where("name = ?", name).
		Limit(1).
		OrderBy("version DESC").
		QueryRowAndClose(context.TODO(), db)

		// Error Executing Query?
	if e != nil && e != sql.ErrNoRows { // YES
		log.Printf("query error: %v\n", e)
		return e
	}

	// Did we retrieve an entry?
	if e == nil { // YES
		o.name = name

		if description.Valid {
			o.description = description.String
		}

		if created.Valid {
			o.created = mysql.MySQLTimeStampToGoTime(created.String)
		}

		o.stored = true
	}

	return nil

}

// Finds Template By Name and Version
func (o *Template) ByNameVersion(db *sql.DB, name string, version uint16) error {
	if version == 0 {
		return o.ByNameLatest(db, name)
	}

	// Validate Template
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("Missing Value for Required Parameter 'name'")
	}
	name = strings.ToLower(name)

	// Reset Entry
	o.reset()

	// Execute Query
	var description sql.NullString
	var created sql.NullString
	e := sqlf.From("templates").
		Select("title").To(&o.title).
		Select("description").To(&description).
		Select("model").To(&o.model).
		Select("created").To(&created).
		Where("name = ? && version = ?", name, version).
		QueryRowAndClose(context.TODO(), db)

	// Error Executing Query?
	if e != nil && e != sql.ErrNoRows { // YES
		log.Printf("query error: %v\n", e)
		return e
	}

	// Did we retrieve an entry?
	if e == nil { // YES
		o.name = name
		o.version = version

		if description.Valid {
			o.description = description.String
		}

		if created.Valid {
			o.created = mysql.MySQLTimeStampToGoTime(created.String)
		}

		o.stored = true
	}

	return nil
}

// ID Local ID
func (o *Template) ID() uint32 {
	if o.id == nil {
		return 0
	}

	return *o.id
}

func (o *Template) Template() string {
	return o.name
}

func (o *Template) Version() uint16 {
	return o.version
}

func (o *Template) Title() string {
	return o.title
}

func (o *Template) Description() string {
	return o.description
}

func (o *Template) Model() string {
	return o.model
}

func (o *Template) Created() *time.Time {
	return o.created
}

func (o *Template) SetID(id uint32) (uint32, error) {
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

	return 0, errors.New("Registered Template - ID is immutable")
}

func (o *Template) SetTemplate(name string, version uint16) error {
	if o.IsNew() {
		// Validate Template
		name = strings.TrimSpace(name)
		if name == "" {
			return errors.New("Missing Value for Required Parameter 'name'")
		}
		name = strings.ToLower(name)

		// Validate Version
		if version == 0 {
			return errors.New("Missing Value for Required Parameter 'version'")
		}

		// New State
		o.name = name
		o.version = version
		o.dirty = true
		return nil
	}

	return errors.New("Registered Template - Key is immutable")
}

func (o *Template) SetTitle(s string) (string, error) {
	// Current State
	current := o.title

	// Validate Store Alias
	if s != current {
		// New State
		o.title = s
		o.dirty = true
	}

	return current, nil
}

func (o *Template) SetDescription(s string) (string, error) {
	// Current State
	current := o.description

	// Validate Store Alias
	if s != current {
		// New State
		o.description = s
		o.dirty = true
	}

	return current, nil
}

func (o *Template) SetModel(s string) (string, error) {
	// Current State
	current := o.model

	// Validate Store Alias
	s = strings.TrimSpace(s)
	if s == "" {
		return "", errors.New("Missing Value for Required Parameter 'model'")
	}

	// New State
	o.model = s
	o.dirty = true

	return current, nil
}

func (o *Template) Flush(db sqlf.Executor, force bool) error {
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
		// Execute Insert
		s := sqlf.InsertInto("templates").
			Set("name", o.name).
			Set("title", o.title).
			Set("version", o.version).
			Set("model", o.model)

		if o.description != "" {
			s.Set("description", o.description)
		}

		_, e := s.Exec(context.TODO(), db)

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
		if o.hasKey() {
			return errors.New("Missing or Invalid Store Key")
		}

		// Create SQL Statement
		s := sqlf.Update("stores").
			Set("title", o.title).
			Set("model", o.model).
			Where("id = ?", o.id)

		// Is Template Description Set?
		if o.description != "" { // YES: Update that as well
			s.Set("description", o.description)
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

func (o *Template) hasKey() bool {
	return o.id != nil
}

func (o *Template) reset() {
	// Clean Entry
	o.id = nil
	o.name = ""
	o.version = 0
	o.title = ""
	o.description = ""
	o.model = ""
	o.created = nil

	// Mark State as Unregistered
	o.stored = false

	// Mark Entry as Clean
	o.dirty = false
}
