package request

/*
 * This file is part of the ObjectVault Project.
 * Copyright (C) 2020-2022 Paulo Ferreira <vault at sourcenotes.org>
 *
 * This work is published under the GNU AGPLv3.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

// cSpell:ignore pjacferreira, reqtype, sqlf
import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/pjacferreira/sqlf"

	"github.com/objectvault/api-services/common"
	"github.com/objectvault/api-services/orm/mysql"
	"github.com/objectvault/api-services/orm/query"
)

// Request Registry Object Definition
type RequestRegistry struct {
	dirty      bool       // Is Entry Dirty?
	stored     bool       // Is Entry Stored in Database
	id         *uint64    // (REQUIRED) KEY: GLOBAL Request ID
	guid       string     // (REQUIRED) Globally Unique ID (NO Shard or Local ID Information since Request Might not Require Session)
	reqtype    string     // (REQUIRED) Request Type
	object     *uint64    // (OPTIONAL) Request Reference Object (used if request linked to an object)
	expiration *time.Time // (OPTIONAL) Expiration Time Stamp
	state      uint16     // (REQUIRED) Request State (0-active, 10-processing, 90-error, 98-expired, 99-completed)
	creator    *uint64    // (OPTIONAL) Creator User ID
	created    *time.Time // (REQUIRED) Request Creation Date
}

func RequestToRegistry(r *Request, group uint16, shard uint32) (*RequestRegistry, error) {
	// Validate Request (Can Only Create Registry after Request Stored)
	if !r.IsValid() || r.IsNew() {
		return nil, errors.New("Invalid Invitation Object")
	}

	// Create Invitation Registry Entry
	rr := &RequestRegistry{
		guid:       r.GUID(),
		reqtype:    r.RequestType(),
		expiration: r.Expiration(),
	}

	rid := common.ShardGlobalID(group, common.OTYPE_REQUEST, shard, r.ID())
	rr.id = &rid

	oid := r.Object()
	if oid != 0 {
		r.object = &oid
	}

	cid := r.Creator()
	if cid != 0 {
		rr.creator = &cid
	}

	return rr, nil
}

func CountRequests(db *sql.DB, q query.TQueryConditions) (uint64, error) {
	// Query Results Values
	var count uint64

	// Create SQL Statement
	s := sqlf.From("registry_requests").
		Select("COUNT(*)").
		To(&count)

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

func CountRequestsByType(db *sql.DB, t string, active bool, oid *uint64) (uint64, error) {
	// Query Results Values
	var count uint64

	// Create Base SQL Statement
	s := sqlf.From("registry_requests").
		Select("COUNT(*)").
		Where("type = ?", t)

	// Have we an Reference Object?
	if oid != nil { // YES: Add Limit
		s.Where("object = ?", oid)
	}

	// Only Active Requests?
	if active {
		s.Where("state < ?", 90)
	}

	// DEBUG: Print SQL
	fmt.Println(s.String())

	// Execute Count
	e := s.
		To(&count).
		QueryRowAndClose(context.TODO(), db)

	// Error Occurred?
	if e != nil { // YES
		log.Printf("query error: %v\n", e)
		return 0, e
	}

	return count, nil
}

func ExpiryRequests(db *sql.DB, q query.TQueryConditions) (uint64, error) {
	// TODO Implement
	return 0, nil
}

func ExpiryRequestsByType(db *sql.DB, t string, oid *uint64) (uint64, error) {
	// TODO Implement
	return 0, nil
}

func GetLastActiveRegistryByType(db *sql.DB, t string, oid *uint64) (*RequestRegistry, error) {
	o := &RequestRegistry{}

	e := o.ByLastActive(db, t, oid, nil)
	if e != nil {
		return nil, e
	}

	if !o.stored {
		return nil, nil
	}

	return o, nil
}

func QueryRequests(db *sql.DB, q query.TQueryConditions, c bool) (query.TQueryResults, error) {
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
	var guid, reqtype string
	var state uint16
	var object, creator sql.NullInt64
	var expiration, created sql.NullString

	// Create SQL Statement
	s := sqlf.From("registry_requests").
		Select("id_request").To(&id).
		Select("guid").To(&guid).
		Select("type").To(&reqtype).
		Select("object").To(&object).
		Select("expiration").To(&expiration).
		Select("id_creator").To(&creator).
		Select("state").To(&state).
		Select("creator").To(&creator).
		Select("created").To(&created)

	// Is OFFSET Set?
	if list.Offset() > 0 { // YES: Use it
		s.Offset(list.Offset())
	}

	// Is LIMIT Set?
	if list.Limit() > 0 { // YES: Use it
		s.Limit(list.Limit())
	}

	// Apply Query Conditions
	e := query.ApplyFilterConditions(s, q)

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
	} else { // DEFAULT: Sort by Type
		list.AppendSort("type", false)
		s.OrderBy("type")
	}

	// DEBUG: Print SQL
	fmt.Println(s.String())

	// Execute Query
	e = s.QueryAndClose(context.TODO(), db, func(row *sql.Rows) {
		id_request := id

		o := RequestRegistry{
			dirty:   false,
			stored:  true,
			id:      &id_request,
			guid:    guid,
			reqtype: reqtype,
			state:   state,
		}

		if object.Valid {
			v := uint64(object.Int64)
			o.object = &v
		}

		if expiration.Valid {
			o.expiration = mysql.MySQLTimeStampToGoTime(expiration.String)
		}

		if creator.Valid {
			v := uint64(creator.Int64)
			o.creator = &v
		}

		if created.Valid {
			o.created = mysql.MySQLTimeStampToGoTime(created.String)
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
		count, e := CountRequests(db, q)
		if e != nil {
			return nil, e
		}

		list.SetMaxCount(count)
	}

	return list, nil
}

// IsDirty Have the Object Properties Changed since last Serialization?
func (o *RequestRegistry) IsDirty() bool {
	return o.dirty
}

func (o *RequestRegistry) IsNew() bool {
	return !o.stored
}

func (o *RequestRegistry) IsValid() bool {
	return o.id != nil && o.guid != "" && o.reqtype != ""
}

func (o *RequestRegistry) IsActive() bool {
	return o.state == 0
}

func (o *RequestRegistry) IsExpired() bool {
	if o.expiration != nil {
		expires := o.expiration.Unix()
		now := time.Now().Unix()
		return now > expires
	}
	return false
}

// ByGUID Finds Entry By Unique ID String
func (o *RequestRegistry) ByGUID(db *sql.DB, guid string) error {
	// Reset Entry
	o.reset()

	// NULLABLE Values
	var object sql.NullInt64
	var expiration sql.NullString
	var creator sql.NullInt64
	var created sql.NullString

	// Execute Query
	e := sqlf.From("registry_requests").
		Select("id").To(o.id).
		Select("type").To(&o.reqtype).
		Select("object").To(&object).
		Select("expiration").To(&expiration).
		Select("creator").To(&creator).
		Select("created").To(&created).
		Where("guid = ?", guid).
		QueryRowAndClose(context.TODO(), db)

	// Error Executing Query?
	if e != nil && e != sql.ErrNoRows { // YES
		log.Printf("query error: %v\n", e)
		return e
	}

	// Did we retrieve an entry?
	if e == nil { // YES
		o.guid = guid
		if object.Valid {
			v := uint64(object.Int64)
			o.object = &v
		}
		if expiration.Valid {
			o.expiration = mysql.MySQLTimeStampToGoTime(expiration.String)
		}
		if creator.Valid {
			v := uint64(creator.Int64)
			o.creator = &v
		}
		if created.Valid {
			o.created = mysql.MySQLTimeStampToGoTime(created.String)
		}
		o.stored = true
	}

	return nil
}

// ByID Finds Entry By ID
func (o *RequestRegistry) ByID(db *sql.DB, id uint64) error {
	// Reset Entry
	o.reset()

	// NULLABLE Values
	var object sql.NullInt64
	var expiration sql.NullString
	var creator sql.NullInt64
	var created sql.NullString

	// Execute Query
	e := sqlf.From("registry_requests").
		Select("guid").To(&o.guid).
		Select("type").To(&o.reqtype).
		Select("object").To(&object).
		Select("expiration").To(&expiration).
		Select("creator").To(&creator).
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
		if object.Valid {
			v := uint64(object.Int64)
			o.object = &v
		}
		if expiration.Valid {
			o.expiration = mysql.MySQLTimeStampToGoTime(expiration.String)
		}
		if creator.Valid {
			v := uint64(creator.Int64)
			o.creator = &v
		}
		if created.Valid {
			o.created = mysql.MySQLTimeStampToGoTime(created.String)
		}
		o.stored = true
	}

	return nil
}

func (o *RequestRegistry) ByLastActive(db *sql.DB, t string, oid *uint64, cid *uint64) error {
	// Reset Entry
	o.reset()

	// NULLABLE Values
	var object sql.NullInt64
	var expiration sql.NullString
	var creator sql.NullInt64
	var created sql.NullString

	// Execute Query
	s := sqlf.From("registry_requests").
		Select("id").To(&o.id).
		Select("guid").To(&o.guid).
		Select("type").To(&o.reqtype).
		Select("object").To(&object).
		Select("expiration").To(&expiration).
		Select("creator").To(&creator).
		Select("created").To(&created).
		Where("active < ?", 90)

	if t != "" {
		s.Where("type = ?", t)
	}

	if oid != nil {
		s.Where("object = ?", oid)
	}

	if cid != nil {
		s.Where("creator = ?", cid)
	}

	e := s.
		OrderBy("type DESC").
		QueryRowAndClose(context.TODO(), db)

	// Error Executing Query?
	if e != nil && e != sql.ErrNoRows { // YES
		log.Printf("query error: %v\n", e)
		return e
	}

	// Did we retrieve an entry?
	if e == nil { // YES
		if object.Valid {
			v := uint64(object.Int64)
			o.object = &v
		}
		if expiration.Valid {
			o.expiration = mysql.MySQLTimeStampToGoTime(expiration.String)
		}
		if creator.Valid {
			v := uint64(creator.Int64)
			o.creator = &v
		}
		if created.Valid {
			o.created = mysql.MySQLTimeStampToGoTime(created.String)
		}
		o.stored = true
	}

	return nil
}

func (o *RequestRegistry) ID() uint64 {
	if o.id == nil {
		return 0
	}

	return *o.id
}

func (o *RequestRegistry) GUID() string {
	return o.guid
}

func (o *RequestRegistry) RequestType() string {
	return o.reqtype
}

func (o *RequestRegistry) Object() uint64 {
	if o.object == nil {
		return 0
	}

	return *o.object
}

func (o *RequestRegistry) Expiration() *time.Time {
	return o.expiration
}

func (o *RequestRegistry) ExpirationUTC() string {
	if o.expiration == nil {
		utc := o.expiration.UTC()

		// RETURN ISO 8601 / RFC 3339 FORMAT in UTC
		return utc.Format(time.RFC3339)
	}

	return ""
}

func (o *RequestRegistry) State() uint16 {
	return o.state
}

func (o *RequestRegistry) Creator() uint64 {
	if o.creator == nil {
		return 0
	}

	return *o.creator
}

func (o *RequestRegistry) Created() *time.Time {
	return o.created
}

func (o *RequestRegistry) SetState(state uint16) error {
	// Current State
	current := state

	// State Changed?
	if current != state { // YES - Update
		o.state = state
		o.dirty = true
	}

	return nil
}

func (o *RequestRegistry) Flush(db sqlf.Executor, force bool) error {
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
		s := sqlf.InsertInto("registry_requests").
			Set("id_request", o.id).
			Set("guid", o.guid).
			Set("reqtype", o.reqtype).
			Set("state", o.state)

		if o.object != nil {
			s.Set("object", o.object)
		}

		if o.expiration != nil {
			s.Set("expiration", mysql.GoTimeToMySQLTimeStamp(o.expiration))
		}

		if o.creator != nil {
			s.Set("creator", o.creator)
		}

		_, e = s.Exec(context.TODO(), db)
	} else { // NO: Update
		_, e = sqlf.Update("registry_requests").
			Set("state", o.state).
			Where("id_request = ?", o.id).
			ExecAndClose(context.TODO(), db)
	}

	if e == nil {
		o.stored = true
		o.dirty = false
	}
	return e
}

func (o *RequestRegistry) reset() {
	// Clean Entry
	o.id = nil
	o.guid = ""
	o.reqtype = ""
	o.object = nil
	o.expiration = nil
	o.state = 0
	o.creator = nil
	o.created = nil

	// Mark State as Unregistered
	o.stored = false

	// Mark Entry as Clean
	o.dirty = false
}
