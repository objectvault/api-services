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
	"log"
	"time"

	"github.com/gofrs/uuid"

	"github.com/objectvault/api-services/maps"
	"github.com/objectvault/api-services/orm/mysql"
	"github.com/pjacferreira/sqlf"
)

// Invitation Object Definition
type Request struct {
	dirty      bool            // Is Entry Dirty?
	stored     bool            // Is Entry Stored in Database
	id         *uint32         // (REQUIRED) LOCAL Request ID
	guid       string          // (SET ON CREATE) Globally Unique ID (NO Shard or Local ID Information since Request Might not Require Session)
	reqtype    string          // (REQUIRED) Request Type
	object     *uint64         // (OPTIONAL) Request Reference Object (used if request linked to an object)
	parameters maps.MapWrapper // (OPTIONAL) Request Parameters
	properties maps.MapWrapper // (OPTIONAL) Request Context Properties
	expiration *time.Time      // (OPTIONAL) Expiration Time Stamp
	creator    *uint64         // (OPTIONAL) Creator User ID
	created    *time.Time      // (SET ON INSERT)  Creation Timestamp
	modifier   *uint64         // (OPTIONAL) Creator User ID
	modified   *time.Time      // (SET ON SAVE) Creation Timestamp
}

func NewRequest(t string, creator uint64) *Request {
	o := &Request{
		reqtype: t,
		creator: &creator,
	}

	// Set UUID
	o.GUID()
	return o
}

// TODO Implement Delete (Both From Within an Entry and Without a Structure)
/* IMPORTANT NOTE:
 * MySQL stores timestamps in UTC, but serves them in local time, as set
 * on the MySQL Server HOST NODE.
 * This MEANS that the NODE on which the GO is Being Run has to be ON THE
 * SAME time zone setting as the MySQL HOST NODE
 */

// IsDirty Have the Object Properties Changed since last Serialization?
func (o *Request) IsDirty() bool {
	return o.dirty || o.parameters.IsModified() || o.properties.IsModified()
}

func (o *Request) IsNew() bool {
	return !o.stored
}

func (o *Request) IsValid() bool {
	return o.creator != nil && o.reqtype != ""
}

// ByID Finds Entry By ID
func (o *Request) ByID(db *sql.DB, id uint32) error {
	// Reset Entry
	o.reset()

	// NULLABLE Values
	var object sql.NullInt64
	var params, props sql.NullString
	var expiration sql.NullString
	var creator, modifier sql.NullInt64
	var created, modified sql.NullString

	// Execute Query
	e := sqlf.From("requests").
		Select("guid").To(&o.guid).
		Select("type").To(&o.reqtype).
		Select("object").To(&object).
		Select("params").To(&params).
		Select("props").To(&props).
		Select("expiration").To(&expiration).
		Select("creator").To(&creator).
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
		if object.Valid {
			v := uint64(object.Int64)
			o.object = &v
		}
		if params.Valid {
			e = o.parameters.Import(params.String)
			if e != nil {
				return e
			}
		}
		if props.Valid {
			e = o.properties.Import(props.String)
			if e != nil {
				return e
			}
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
		if modifier.Valid {
			v := uint64(modifier.Int64)
			o.modifier = &v
		}
		if created.Valid {
			o.created = mysql.MySQLTimeStampToGoTime(created.String)
		}
		o.stored = true
	}

	return nil
}

// ID Get User ID
func (o *Request) ID() uint32 {
	if o.id == nil {
		return 0
	}

	return *o.id
}

func (o *Request) GUID() string {
	if o.guid == "" {
		u, _ := uuid.NewV4()
		o.guid = u.String()
	}
	return o.guid
}

func (o *Request) RequestType() string {
	return o.reqtype
}

func (o *Request) Object() uint64 {
	if o.object == nil {
		return 0
	}

	return *o.object
}

func (o *Request) Parameters() *maps.MapWrapper {
	return &o.parameters
}

func (o *Request) Properties() *maps.MapWrapper {
	return &o.properties
}

func (o *Request) Expiration() *time.Time {
	return o.expiration
}

func (o *Request) ExpirationUTC() string {
	if o.expiration == nil {
		utc := o.expiration.UTC()

		// RETURN ISO 8601 / RFC 3339 FORMAT in UTC
		return utc.Format(time.RFC3339)
	}

	return ""
}

func (o *Request) Creator() uint64 {
	if o.creator == nil {
		return 0
	}

	return *o.creator
}

func (o *Request) Created() *time.Time {
	return o.created
}

func (o *Request) Modifier() uint64 {
	if o.modifier == nil {
		return 0
	}

	return *o.modifier
}

func (o *Request) Modified() *time.Time {
	return o.modified
}

func (o *Request) SetObject(id uint64) (uint64, error) {
	if !o.IsNew() {
		return 0, errors.New("Registered Invitation is immutable")
	}

	// Current State
	current := uint64(0)
	if o.object != nil {
		current = *o.object
	}

	// New State
	o.object = &id
	o.dirty = true
	return current, nil
}

func (o *Request) SetExpiration(t time.Time) error {
	o.expiration = &t
	o.dirty = true

	return nil
}

func (o *Request) SetExpiresIn(days uint16) error {
	// Have DB Connection?
	if days == 0 { // NO: Abort
		return errors.New("Number of days should be > 0")
	}

	now := time.Now()
	expires := now.AddDate(0, 0, int(days))
	o.expiration = &expires
	o.dirty = true
	return nil
}

func (o *Request) SetCreator(id uint64) (uint64, error) {
	if !o.IsNew() {
		return 0, errors.New("Registered Invitation is immutable")
	}

	// Current State
	current := uint64(0)
	if o.creator != nil {
		current = *o.creator
	}

	// New State
	o.creator = &id
	o.dirty = true
	return current, nil
}

func (o *Request) SetModifier(id uint64) (uint64, error) {
	// NOTE: Setting "modifier" for new request has no meaningful effect

	// Current State
	current := uint64(0)
	if o.modifier != nil {
		current = *o.modifier
	}

	// New State
	o.modifier = &id
	o.dirty = true
	return current, nil
}

func (o *Request) Flush(db sqlf.Executor, force bool) error {
	// Have DB Connection?
	if db == nil { // NO: Abort
		return errors.New("Missing Database Connection")
	}

	// Has entry been modified?
	if !force && !o.IsDirty() { // NO: Abort
		return nil
	}

	// Do we have a request type set?
	if o.reqtype == "" { // NO: Abort
		return errors.New("Missing Value for Request Type")
	}

	// Is New Entry?
	var e error
	if o.IsNew() { // YES: Create
		// Execute Insert
		s := sqlf.InsertInto("requests").
			Set("guid", o.GUID()).
			Set("type", o.reqtype)

		if o.object != nil {
			s.Set("object", o.object)
		}

		if !o.parameters.IsEmpty() {
			v := o.parameters.Export()
			if v != "" {
				s.Set("params", v)
			}
		}

		if !o.properties.IsEmpty() {
			v := o.properties.Export()
			if v != "" {
				s.Set("props", v)
			}
		}

		if o.expiration != nil {
			s.Set("expiration", mysql.GoTimeToMySQLTimeStamp(o.expiration))
		}

		if o.creator != nil {
			s.Set("creator", o.creator)
		}

		/* NOTE:
				 * "created" and "modified" time stamps are set by database on insert / update
		     * so are never set in SQL Statement
		*/
		_, e = s.Exec(context.TODO(), db)

		// Error Occurred?
		if e == nil { // NO: Get Last Insert ID
			var id uint32
			e = sqlf.Select("LAST_INSERT_ID()").
				To(&id).
				QueryRowAndClose(context.TODO(), db)

				// Error Occurred?
			if e == nil { // NO: Set Object ID
				o.id = &id
			}
		}
	} else {
		if o.modifier == nil {
			return errors.New("Modifier user not Set")
		}

		return errors.New("Request Objects are Immutable")
	}

	if e == nil {
		o.stored = true
		o.dirty = false
	}
	return e
}

func (o *Request) reset() {
	// Clean Entry
	o.id = nil
	o.guid = ""
	o.reqtype = ""
	o.object = nil
	o.object = nil
	o.parameters.Reset()
	o.properties.Reset()
	o.expiration = nil
	o.creator = nil
	o.created = nil
	o.modifier = nil
	o.modified = nil

	// Mark State as Unregistered
	o.stored = false

	// Mark Entry as Clean
	o.dirty = false
}
