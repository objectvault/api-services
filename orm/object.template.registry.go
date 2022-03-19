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

	"github.com/pjacferreira/sqlf"
)

// Object Template Registry Definition
type ObjectTemplateRegistry struct {
	dirty    bool    // Is Entry Dirty?
	stored   bool    // Is Entry Stored in Database
	object   *uint64 // KEY: GLOBAL Object ID
	template string  // Template Name
	title    string  // Template Title
}

func ExistsRegisteredObjectTemplate(db *sql.DB, object uint64, template string) (bool, error) {
	// Query Results Values
	var count uint64

	// Create SQL Statement
	s := sqlf.From("registry_object_templates").
		Select("COUNT(*)").To(&count).
		Where("id_object = ? and template = ?", object, template)

	// DEBUG: Print SQL
	fmt.Println(s.String())

	// Execute Count
	e := s.QueryRowAndClose(context.TODO(), db)

	// Error Executing Query?
	if e != nil && e != sql.ErrNoRows { // YES
		log.Printf("query error: %v\n", e)
		return false, e
	}

	// Error Occurred?
	if e != nil { // YES
		log.Printf("query error: %v\n", e)
		return false, e
	}

	return count > 0, nil
}

func DeleteRegisteredObjectTemplate(db *sql.DB, object uint64, template string) error {
	// Create SQL Statement
	s := sqlf.DeleteFrom("registry_object_templates").
		Where("id_object = ? and template = ?", object, template)

	// DEBUG: Print SQL
	fmt.Println(s.String())

	// Execute Count
	e := s.QueryRowAndClose(context.TODO(), db)

	// Error Executing Query?
	if e != nil && e != sql.ErrNoRows { // YES
		log.Printf("query error: %v\n", e)
		return e
	}

	// Error Occurred?
	if e != nil { // YES
		log.Printf("query error: %v\n", e)
		return e
	}

	return nil
}

func CountRegisteredObjectTemplates(db *sql.DB, object uint64, q TQueryConditions) (uint64, error) {
	// Query Results Values
	var count uint64

	// Create SQL Statement
	s := sqlf.From("registry_object_templates").
		Select("COUNT(*)").To(&count).
		Where("id_object = ?", object)

	// Apply Query Conditions
	e := applyFilterConditions(s, q)

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

func QueryRegisteredObjectTemplates(db *sql.DB, object uint64, q TQueryConditions, c bool) (TQueryResults, error) {
	var list QueryResults = QueryResults{}
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
	var name, title string

	// Create SQL Statement
	s := sqlf.From("registry_object_templates").
		Select("id_object").To(&id).
		Select("title").To(&title).
		Select("template").To(&name)

	// Is OFFSET Set?
	if list.Offset() > 0 { // YES: Use it
		s.Offset(list.Offset())
	}

	// Is LIMIT Set?
	if list.Limit() > 0 { // YES: Use it
		s.Limit(list.Limit())
	}

	// Apply Query Conditions
	e := applyFilterConditions(s, q)

	// Error Occurred?
	if e != nil && e != sql.ErrNoRows { // YES
		log.Printf("query error: %v\n", e)
		return nil, e
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
		list.AppendSort("template", false)
		s.OrderBy("template")
	}

	// Apply Organization Filter
	s.Where("id_object = ?", object)

	// DEBUG: Print SQL
	fmt.Println(s.String())

	// Execute Query
	e = s.QueryAndClose(context.TODO(), db, func(row *sql.Rows) {
		id_object := id
		e := ObjectTemplateRegistry{
			object:   &id_object,
			template: name,
			title:    title,
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
		count, e := CountRegisteredObjectTemplates(db, object, q)
		if e != nil {
			return nil, e
		}

		list.SetMaxCount(count)
	}

	return list, nil
}

// IsDirty Have the Object Properties Changed since last Serialization?
func (o *ObjectTemplateRegistry) IsDirty() bool {
	return o.dirty
}

func (o *ObjectTemplateRegistry) IsNew() bool {
	return !o.stored
}

func (o *ObjectTemplateRegistry) IsValid() bool {
	return o.hasKey()
}

func (o *ObjectTemplateRegistry) IsSystemOrganization() bool {
	return o.object != nil && *o.object == uint64(0x2000000000000)
}

// ID Get Organization ID
func (o *ObjectTemplateRegistry) Object() uint64 {
	return *o.object
}

func (o *ObjectTemplateRegistry) Template() string {
	return o.template
}

func (o *ObjectTemplateRegistry) Title() string {
	return o.title
}

func (o *ObjectTemplateRegistry) SetKey(object uint64, template string) error {
	if o.IsNew() {
		// New State
		o.object = &object
		o.template = template
		o.dirty = true
		return nil
	}

	return errors.New("Entry KEY is immutable")
}

func (o *ObjectTemplateRegistry) SetTitle(s string) (string, error) {
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

func (o *ObjectTemplateRegistry) Delete(db *sql.DB, force bool) error {
	if !o.hasKey() {
		return errors.New("Invalid Registry Entry")
	}

	e := DeleteRegisteredObjectTemplate(db, *o.object, o.template)
	if e != nil {
		return e
	}

	// Mark State as Unregistered
	o.stored = false

	// Mark Entry as Dirty
	o.dirty = false
	return nil
}

func (o *ObjectTemplateRegistry) Flush(db sqlf.Executor, force bool) error {
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

		s = sqlf.InsertInto("registry_object_templates").
			Set("id_object", o.object).
			Set("username", o.template)
	} else { // NO: Update
		return errors.New("Table does not allow update")
	}

	// Execute Statement
	_, e := s.ExecAndClose(context.TODO(), db)
	if e == nil {
		o.stored = true
		o.dirty = false
	}
	return e
}

func (o *ObjectTemplateRegistry) hasKey() bool {
	return o.object != nil && o.template != ""
}

func (o *ObjectTemplateRegistry) reset() {
	// Clean Entry
	o.object = nil
	o.template = ""

	// Mark State as Unregistered
	o.stored = false

	// Mark Entry as Clean
	o.dirty = false
}
