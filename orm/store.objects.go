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
	"time"

	"github.com/objectvault/api-services/orm/mysql"
	"github.com/pjacferreira/sqlf"
)

// STORE Object Definition
type StoreObject struct {
	dirty    bool       // Is Entry Dirty?
	stored   bool       // Is Entry Stored in Database
	store    uint32     // KEY: SHARD Local Store ID
	parent   uint32     // KEY: Parent Object ID (0 == ROOT)
	id       uint32     // KEY: SHARD UNIQUE Local Object ID
	title    string     // Object Title
	objtype  uint8      // Object Type
	object   []byte     // Encrypted Store Object
	creator  *uint64    // Global User ID of Creator
	created  *time.Time // Created TimeStamp
	modifier *uint64    // Global User ID of Last Modifier
	modified *time.Time // Modification TimeStamp
}

// KNOWN OBJECT TYPES
const OBJECT_TYPE_FOLDER = 0
const OBJECT_TYPE_JSON = 1

func ChildObjectFromParent(p *StoreObject) (*StoreObject, error) {
	if p.Type() != OBJECT_TYPE_FOLDER {
		return nil, errors.New("Parent is Not a Folder Object")
	}

	// Create Child Object
	o := &StoreObject{
		store:  p.store,
		parent: p.id,
	}

	return o, nil
}

func DeleteStoreObjectFolder(db *sql.DB, store uint32, folder uint32) error {

	// Delete Child Objects
	_, e := sqlf.DeleteFrom("objects").
		Where("id_store = ? and id_parent = ?", store, folder).
		ExecAndClose(context.TODO(), db)

		// Error Occurred?
	if e != nil { // YES
		log.Printf("query error: %v\n", e)
		return e
	}

	// Delete Folder Object
	_, e = sqlf.DeleteFrom("objects").
		Where("id_store = ? and id = ?", store, folder).
		ExecAndClose(context.TODO(), db)

	// Error Occurred?
	if e != nil { // YES
		log.Printf("query error: %v\n", e)
	}

	return e
}

func DeleteStoreObject(db *sql.DB, store uint32, oid uint32) error {
	// Delete Folder Object
	_, e := sqlf.DeleteFrom("objects").
		Where("id_store = ? and id = ?", store, oid).
		ExecAndClose(context.TODO(), db)

	// Error Occurred?
	if e != nil { // YES
		log.Printf("query error: %v\n", e)
	}

	return e
}

func CountStoreObject(db *sql.DB, store uint32, q TQueryConditions) (uint64, error) {
	// Query Results Values
	var count uint64

	// Create SQL Statement
	s := sqlf.From("objects").
		Select("COUNT(*)").To(&count).
		Where("id_store = ?", store)

		// Apply Query Conditions
	e := applyFilterConditions(s, q)

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

func CountStoreParentObject(db *sql.DB, store uint32, parent uint32, q TQueryConditions) (uint64, error) {
	// Query Results Values
	var count uint64

	// Create SQL Statement
	s := sqlf.From("objects").
		Select("COUNT(*)").To(&count).
		Where("id_store = ? and id_parent = ?", store, parent)

	// Apply Query Conditions
	e := applyFilterConditions(s, q)

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

// TODO Implement Delete (Both From Within an Entry and Without a Structure)
func QueryStoreObjects(db *sql.DB, store uint32, query string) ([]StoreObject, error) {
	var entries []StoreObject

	// Query Results Values
	var id uint32
	var parent uint32
	var title string
	var objtype uint8

	// Create SQL Statement
	s := sqlf.From("objects").
		Select("id_parent").To(&parent).
		Select("id").To(&id).
		Select("title").To(&title).
		Select("type").To(&objtype).
		Where("id_store = ?", store)

	// Execute Query
	e := s.QueryAndClose(context.TODO(), db, func(row *sql.Rows) {
		o := StoreObject{
			stored:  true,
			store:   store,
			parent:  parent,
			id:      id,
			title:   title,
			objtype: objtype,
		}

		if entries == nil {
			entries = make([]StoreObject, 1)
			entries[0] = o
		} else {
			entries = append(entries, o)
		}
	})

	// Error Occurred?
	if e != nil && e != sql.ErrNoRows { // YES
		log.Printf("query error: %v\n", e)
		return entries, e
	}

	return entries, nil
}

func QueryStoreParentObjects(db *sql.DB, store uint32, parent uint32, q TQueryConditions, c bool) (TQueryResults, error) {
	var list QueryResults = QueryResults{}
	list.SetMaxLimit(100) // Hard Code Maximum Limit

	// Set Query Page Limits
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
	var id uint32
	var title string
	var objtype uint8

	// Create SQL Statement
	s := sqlf.From("objects").
		Select("id").To(&id).
		Select("title").To(&title).
		Select("type").To(&objtype).
		Where("id_store = ? and id_parent = ?", store, parent)

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
		list.AppendSort("title", false)
		s.OrderBy("title")
	}

	// Apply Query Conditions
	e := applyFilterConditions(s, q)

	// Error Occurred?
	if e != nil { // YES
		log.Printf("query error: %v\n", e)
		return nil, e
	}

	// DEBUG: Print SQL
	fmt.Print(s.String())

	// Execute Query
	e = s.QueryAndClose(context.TODO(), db, func(row *sql.Rows) {
		o := StoreObject{
			stored:  true,
			store:   store,
			parent:  parent,
			id:      id,
			title:   title,
			objtype: objtype,
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
		count, e := CountStoreParentObject(db, store, parent, q)
		if e != nil {
			return nil, e
		}

		list.SetMaxCount(count)
	}

	return list, nil
}

// IsDirty Have the Object Properties Changed since last Serialization?
func (o *StoreObject) IsDirty() bool {
	return o.dirty
}

func (o *StoreObject) IsNew() bool {
	return !o.stored
}

// ByID Finds Entry By Object ID
func (o *StoreObject) ByID(db *sql.DB, id uint32) error {
	// Cleanup Entry
	o.reset()

	// Execute Query
	var object sql.NullString
	var created sql.NullString
	var modifier sql.NullInt64
	var modified sql.NullString
	e := sqlf.From("objects").
		Select("id_store").To(&o.store).
		Select("id_parent").To(&o.parent).
		Select("title").To(&o.title).
		Select("type").To(&o.objtype).
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
		o.id = id

		if object.Valid {
			s := object.String
			o.object = []byte(s)
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

		o.stored = true // Registered Entry
	}

	return nil
}

// ByKey Finds Entry By Store / Object ID
func (o *StoreObject) ByKey(db *sql.DB, store uint32, id uint32) error {
	// Cleanup Entry
	o.reset()

	// Execute Query
	var object sql.NullString
	var created sql.NullString
	var modifier sql.NullInt64
	var modified sql.NullString
	e := sqlf.From("objects").
		Select("id_parent").To(&o.parent).
		Select("title").To(&o.title).
		Select("type").To(&o.objtype).
		Select("object").To(&object).
		Select("creator").To(&o.creator).
		Select("created").To(&created).
		Select("modifier").To(&modifier).
		Select("modified").To(&modified).
		Where("id_store = ? and id = ?", store, id).
		QueryRowAndClose(context.TODO(), db)

		// Error Executing Query?
	if e != nil && e != sql.ErrNoRows { // YES
		log.Printf("query error: %v\n", e)
		return e
	}

	// Did we retrieve an entry?
	if e == nil { // YES
		o.store = store
		o.id = id

		if object.Valid {
			s := object.String
			o.object = []byte(s)
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

		o.stored = true // Registered Entry
	}

	return nil
}

func (o *StoreObject) Store() uint32 {
	return o.store
}

func (o *StoreObject) Parent() uint32 {
	return o.parent
}

func (o *StoreObject) ID() uint32 {
	return o.id
}

func (o *StoreObject) Title() string {
	return o.title
}

func (o *StoreObject) Type() uint8 {
	return o.objtype
}

func (o *StoreObject) Object() []byte {
	return o.object
}

func (o *StoreObject) Creator() uint64 {
	if o.creator == nil {
		return 0
	}
	return *o.creator
}

func (o *StoreObject) Created() *time.Time {
	return o.created
}

func (o *StoreObject) Modifier() *uint64 {
	if o.modifier == nil {
		return nil
	}
	return o.modifier
}

func (o *StoreObject) Modified() *time.Time {
	return o.modified
}

func (o *StoreObject) SetStore(s uint32) (uint32, error) {
	if o.IsNew() {
		// Current State
		current := o.store

		// New State
		o.store = s
		o.dirty = true
		return current, nil
	}

	return 0, errors.New("Registered Object - Parent Store is immutable")
}

func (o *StoreObject) SetParent(p uint32) (uint32, error) {
	// Current State
	current := o.parent

	// New State
	o.parent = p
	o.dirty = true
	return current, nil
}

func (o *StoreObject) SetTitle(t string) (string, error) {
	// Current State
	current := o.title

	// New State
	o.title = t
	o.dirty = true
	return current, nil
}

func (o *StoreObject) SetType(t uint8) (uint8, error) {
	if o.IsNew() {
		// Current State
		current := o.objtype

		// New State
		o.objtype = t
		o.dirty = true
		return current, nil
	}

	return 0, errors.New("Registered Object - Type is immutable")
}

func (o *StoreObject) SetObject(ob []byte) error {
	if ob == nil {
		o.object = nil
		return nil
	}

	if len(ob) > 65535 {
		return errors.New("Object too big")
	}

	o.object = ob
	return nil
}

func (o *StoreObject) SetCreator(id uint64) error {
	// Is Record New?
	if o.IsNew() { // YES
		o.creator = &id

		// Creation Time Stamp AUTO SET by MySQL
		return nil
	}

	return errors.New("Registered Object - Creator Cannot be Changed")
}

func (o *StoreObject) SetModifier(id uint64) error {
	// Set Modifier
	o.modifier = &id

	// Modification Time Stamp AUTO SET by MySQL
	return nil
}

func (o *StoreObject) Flush(db sqlf.Executor, force bool) error {
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
		msg := o.isValidForInsert()
		if msg != "" {
			return errors.New(msg)
		}

		// Execute Insert
		s := sqlf.
			InsertInto("objects").
			Set("id_store", o.store).
			Set("id_parent", o.parent).
			Set("id", o.id).
			Set("title", o.title).
			Set("type", o.objtype).
			Set("creator", o.creator)

		if o.object != nil {
			s.Set("object", o.object)
		}

		_, e = s.Exec(context.TODO(), db)

		// Error Occured?
		if e == nil { // NO: Get Last Insert ID
			var id uint32
			e = sqlf.Select("LAST_INSERT_ID()").
				To(&id).
				QueryRowAndClose(context.TODO(), db)

			// Error Occured?
			if e == nil { // NO: Set Object ID
				o.id = id
			}
		}
	} else { // NO: Update
		// TODO: Implement
		msg := o.isValidForUpdate()
		if msg != "" {
			return errors.New(msg)
		}

		// Create SQL Statement
		s := sqlf.Update("objects").
			Set("title", o.title).
			Set("modifier", o.modifier).
			Set("object", o.object).
			Where("id_store = ? and id = ?", o.store, o.id)

		// Execute Statement
		_, e = s.ExecAndClose(context.TODO(), db)
	}

	if e == nil {
		o.stored = true
		o.dirty = false
	}
	return e
}

func (o *StoreObject) isValidForInsert() string {
	if o.store == 0 {
		return "Object Missing Parent Store ID"
	}
	if o.title == "" {
		return "Object Missing Title"
	}
	if o.creator == nil {
		return "Object Missing Creation User ID"
	}
	if (o.objtype > 0) && (o.object == nil) { // Other Object Types Require a Value
		return "Object Missing Value"
	}
	return ""
}

func (o *StoreObject) isValidForUpdate() string {
	if o.store == 0 {
		return "Object Missing Parent Store ID"
	}
	if o.id == 0 {
		return "Object Missing ID"
	}
	if o.modifier == nil {
		return "Object Missing Modifier User ID"
	}
	return ""
}

func (o *StoreObject) reset() {
	// Clean Entry
	o.store = 0
	o.parent = 0
	o.id = 0
	o.title = ""
	o.objtype = 0
	o.object = nil
	o.creator = nil
	o.created = nil
	o.modifier = nil
	o.modified = nil

	// Mark State as Unregistered
	o.stored = false

	// Mark Entry as Clean
	o.dirty = false
}
