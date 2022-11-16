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

	"github.com/objectvault/api-services/common"
	"github.com/objectvault/api-services/orm/mysql"
	"github.com/objectvault/api-services/orm/query"

	"github.com/pjacferreira/sqlf"
)

// User <--> Container Registry Definition
type UserObjectRegistry struct {
	dirty    bool    // Is Entry Dirty?
	stored   bool    // Is Entry Stored in Database
	user     *uint64 // KEY: GLOBAL User ID
	object   *uint64 // KEY: GLOBAL Container ID
	alias    string  // Container Alias
	favorite bool    // FLAG: Is Container Favorite
}

func DeleteRegisteredUserObject(db *sql.DB, user uint64, object uint64) (bool, error) {
	// Create SQL Statement
	s := sqlf.DeleteFrom("registry_user_objects").
		Where("id_user = ? and id_object= ?", user, object)

	// Execute Count
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
			return 0, e
		}
	*/
	return true, nil
}

func CountRegisteredUserObjectsByType(db *sql.DB, user uint64, ltype uint16, q query.TQueryConditions) (uint64, error) {
	// Query Results Values
	var count uint64

	// Create SQL Statement
	s := sqlf.From("registry_user_objects").
		Select("COUNT(*)").To(&count).
		Where("id_user = ? and type = ?", user, ltype)

	// Apply Extra Query Conditions
	e := query.ApplyFilterConditions(s, q)
	if e != nil { // Error Occurred
		log.Printf("query error: %v\n", e)
		return 0, e
	}

	// Execute Count
	e = s.QueryRowAndClose(context.TODO(), db)
	if e != nil { // YES
		log.Printf("query error: %v\n", e)
		return 0, e
	}

	return count, nil
}

func CountRegisteredUserObjects(db *sql.DB, user uint64, q query.TQueryConditions) (uint64, error) {
	// Query Results Values
	var count uint64

	// Create SQL Statement
	s := sqlf.From("registry_user_objects").
		Select("COUNT(*)").To(&count).
		Where("id_user = ?", user)

	// Apply Extra Query Conditions
	e := query.ApplyFilterConditions(s, q)
	if e != nil { // Error Occurred
		log.Printf("query error: %v\n", e)
		return 0, e
	}

	// Execute Count
	e = s.QueryRowAndClose(context.TODO(), db)
	if e != nil { // YES
		log.Printf("query error: %v\n", e)
		return 0, e
	}

	return count, nil
}

// TODO Implement Delete (Both From Within an Entry and Without a Structure)
func QueryRegisteredUserObjectsByType(db *sql.DB, user uint64, ltype uint16, q query.TQueryConditions, c bool) (query.TQueryResults, error) {
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
	var favorite uint16

	// Create SQL Statement
	s := sqlf.From("registry_user_objects").
		Select("id_object").To(&id).
		Select("alias").To(&alias).
		Select("favorite").To(&favorite)

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
	} else { // DEFAULT: Sort by Title
		list.AppendSort("type", false)
		list.AppendSort("alias", false)
		s.OrderBy("type, alias")
	}

	// Apply User Filter
	s.Where("id_user = ? and type = ?", user, ltype)

	// Apply Extra Query Conditions
	e := query.ApplyFilterConditions(s, q)
	if e != nil { // Error Occurred
		// DEBUG: Print SQL
		log.Print(s.String())
		log.Printf("query error: %v\n", e)
		return nil, e
	}

	// Execute Query
	e = s.QueryAndClose(context.TODO(), db, func(row *sql.Rows) {
		object_id := id

		o := UserObjectRegistry{
			user:     &user,
			object:   &object_id,
			alias:    alias,
			favorite: favorite == 1,
		}

		list.AppendValue(&o)
	})

	// Error Occurred?
	if e != nil && e != sql.ErrNoRows { // YES
		// DEBUG: Print SQL
		fmt.Print(s.String())

		log.Printf("query error: %v\n", e)
		return nil, e
	}

	return list, nil
}

func QueryRegisteredUserObjects(db *sql.DB, user uint64, q query.TQueryConditions, c bool) (query.TQueryResults, error) {
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
	var favorite uint16

	// Create SQL Statement
	s := sqlf.From("registry_user_objects").
		Select("id_object").To(&id).
		Select("alias").To(&alias).
		Select("favorite").To(&favorite)

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
	} else { // DEFAULT: Sort by Title
		list.AppendSort("type", false)
		list.AppendSort("alias", false)
		s.OrderBy("type, alias")
	}

	// Apply User Filter
	s.Where("id_user = ?", user)

	// Apply Extra Query Conditions
	e := query.ApplyFilterConditions(s, q)
	if e != nil { // Error Occurred
		// DEBUG: Print SQL
		log.Print(s.String())
		log.Printf("query error: %v\n", e)
		return nil, e
	}

	// Execute Query
	e = s.QueryAndClose(context.TODO(), db, func(row *sql.Rows) {
		object_id := id

		o := UserObjectRegistry{
			user:     &user,
			object:   &object_id,
			alias:    alias,
			favorite: favorite == 1,
		}

		list.AppendValue(&o)
	})

	// Error Occurred?
	if e != nil && e != sql.ErrNoRows { // YES
		// DEBUG: Print SQL
		log.Printf("query error: %v\n", e)
		return nil, e
	}

	return list, nil
}

// IsDirty Have the Object Properties Changed since last Serialization?
func (o *UserObjectRegistry) IsDirty() bool {
	return o.dirty
}

func (o *UserObjectRegistry) IsNew() bool {
	return !o.stored
}

func (o *UserObjectRegistry) IsValid() bool {
	return o.user != nil && o.object != nil && o.alias != ""
}

func (o *UserObjectRegistry) IsSystemOrganization() bool {
	return o.object != nil && *o.object == uint64(0x2000000000000)
}

// ByID Finds Entry By User / Org
func (o *UserObjectRegistry) ByKey(db *sql.DB, user uint64, object uint64) error {
	// Mark State as Unregistered
	o.stored = false

	// Mark Entry as Clean
	o.dirty = false

	// Execute Query
	var fav uint8
	e := sqlf.From("registry_user_objects").
		Select("alias").To(&o.alias).
		Select("favorite").To(&fav).
		Where("id_user = ? and id_object = ?", user, object).
		QueryRowAndClose(context.TODO(), db)

	// Test Results
	switch {
	case e == sql.ErrNoRows: // Entry not Found
		o.user = nil
		o.object = nil
		o.alias = ""
	case e == nil: // Entry found
		o.user = &user
		o.object = &object
		o.favorite = mysql.MySQLtoBool(fav)
		o.stored = true
	default: // DB Error
		o.user = nil
		o.object = nil
		o.alias = ""
		log.Printf("query error: %v\n", e)
		return e
	}

	return nil
}

// ID Get Organization ID
func (o *UserObjectRegistry) User() uint64 {
	return *o.user
}

func (o *UserObjectRegistry) Type() uint16 {
	if o.object != nil {
		return common.ObjectTypeFromID(*o.object)
	}
	return common.OTYPE_NOTSET
}

func (o *UserObjectRegistry) Object() uint64 {
	return *o.object
}

func (o *UserObjectRegistry) Alias() string {
	return o.alias
}

func (o *UserObjectRegistry) Favorite() bool {
	return o.favorite
}

func (o *UserObjectRegistry) SetKey(user uint64, object uint64) error {
	if o.IsNew() {
		// New State
		o.user = &user
		o.object = &object
		o.dirty = true
		return nil
	}

	return errors.New("Entry KEY is immutable")
}

func (o *UserObjectRegistry) SetAlias(alias string) (string, error) {
	// Current State
	current := o.alias

	// Modify Alias
	o.alias = alias
	o.dirty = true
	return current, nil
}

func (o *UserObjectRegistry) SetFavorite(f bool) (bool, error) {
	// Current State
	current := o.favorite

	// New Organization Name
	o.favorite = f
	o.dirty = true
	return current, nil
}

func (o *UserObjectRegistry) Flush(db sqlf.Executor, force bool) error {
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
		_, e = sqlf.InsertInto("registry_user_objects").
			Set("id_user", o.user).
			Set("type", common.ObjectTypeFromID(*o.object)).
			Set("id_object", o.object).
			Set("alias", o.alias).
			Set("favorite", mysql.BoolToMySQL(o.favorite)).
			ExecAndClose(context.TODO(), db)
	} else { // NO: Update
		if o.user == nil {
			return errors.New("User ID not Set")
		}

		if o.object == nil {
			return errors.New("Object ID not Set")
		}

		if o.alias == "" {
			return errors.New("Link Alias not Set")
		}

		_, e = sqlf.Update("registry_user_objects").
			Set("alias", o.alias).
			Set("favorite", mysql.BoolToMySQL(o.favorite)).
			Where("id_user = ? and id_object = ?", o.user, o.object).
			ExecAndClose(context.TODO(), db)
	}

	if e == nil {
		o.stored = true
		o.dirty = false
	}
	return e
}
