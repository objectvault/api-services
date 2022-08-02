package action

/*
 * This file is part of the ObjectVault Project.
 * Copyright (C) 2020-2022 Paulo Ferreira <vault at sourcenotes.org>
 *
 * This work is published under the GNU AGPLv3.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

// cSpell:ignore atype, pjacferreira, sqlf
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
type Action struct {
	dirty      bool            // Is Entry Dirty?
	stored     bool            // Is Entry Stored in Database
	guid       string          // (REQUIRED) Globally Unique ID (Actions should be stored on Specific Shard Group)
	parent     string          // (OPTIONAL) Parent Action if (action decomposed)
	atype      string          // (REQUIRED) Request Type
	parameters maps.MapWrapper // (OPTIONAL) Request Parameters
	properties maps.MapWrapper // (OPTIONAL) Request Context Properties
	state      uint16          // (REQUIRED) Request State (0-active, 10-processing, 90-error, 98-expired, 99-completed)
	creator    *uint64         // (OPTIONAL) Creator User ID
	created    *time.Time      // (SET ON INSERT)  Creation Timestamp
}

func NewActionWithGUID(guid, t string, creator uint64) *Action {
	o := &Action{
		guid:    guid, // Set UUID
		atype:   t,
		creator: &creator,
	}

	return o
}

func NewAction(t string, creator uint64) *Action {
	o := &Action{
		atype:   t,
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
func (o *Action) IsDirty() bool {
	return o.dirty || o.parameters.IsModified() || o.properties.IsModified()
}

func (o *Action) IsNew() bool {
	return !o.stored
}

func (o *Action) IsValid() bool {
	if o.IsNew() {
		return o.atype != "" && o.creator != nil
	}
	return o.guid != "" && o.atype != "" && o.creator != nil
}

// ByGUID Finds Entry By GUID
func (o *Action) ByGUID(db *sql.DB, guid string) error {
	// Reset Entry
	o.reset()

	// Query Results Values
	var parent sql.NullString
	var params, props sql.NullString
	var creator sql.NullInt64
	var created sql.NullString

	// Execute Query
	e := sqlf.From("actions").
		Select("parent").To(&parent).
		Select("type").To(&o.atype).
		Select("params").To(&params).
		Select("props").To(&props).
		Select("state").To(&o.state).
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
		if parent.Valid {
			o.parent = parent.String
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

func (o *Action) GUID() string {
	if o.guid == "" {
		u, _ := uuid.NewV4()
		o.guid = u.String()
	}
	return o.guid

}

func (o *Action) Parent() string {
	return o.parent
}

func (o *Action) Type() string {
	return o.atype
}

func (o *Action) Parameters() *maps.MapWrapper {
	return &o.parameters
}

func (o *Action) Properties() *maps.MapWrapper {
	return &o.properties
}

func (o *Action) State() uint16 {
	return o.state
}

func (o *Action) Creator() uint64 {
	if o.creator == nil {
		return 0
	}

	return *o.creator
}

func (o *Action) Created() *time.Time {
	return o.created
}

func (o *Action) SetParent(guid string) (string, error) {
	if !o.IsNew() {
		return "", errors.New("Registered Invitation is immutable")
	}

	// Current State
	current := o.parent

	// New State
	o.parent = guid
	o.dirty = true
	return current, nil
}

func (o *Action) SetState(state uint16) error {
	// Current State
	current := state

	// State Changed?
	if current != state { // YES - Update
		o.state = state
		o.dirty = true
	}

	return nil
}

func (o *Action) SetStateQueued() error {
	return o.SetState(STATE_QUEUED)
}

func (o *Action) SetStateProcessed() error {
	return o.SetState(STATE_PROCESSED)
}

func (o *Action) SetCreator(id uint64) (uint64, error) {
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

func (o *Action) Flush(db sqlf.Executor, force bool) error {
	// Have DB Connection?
	if db == nil { // NO: Abort
		return errors.New("Missing Database Connection")
	}

	// Do we have a request type set?
	if !o.IsValid() { // NO: Abort
		return errors.New("Missing Value for Request Type")
	}

	// Has entry been modified?
	if !force && !o.IsDirty() { // NO: Abort
		return nil
	}

	// Is New Entry?
	var e error
	if o.IsNew() { // YES: Create
		// Execute Insert
		s := sqlf.InsertInto("actions").
			Set("guid", o.GUID()).
			Set("type", o.atype).
			Set("state", o.state)

		if o.parent != "" {
			s.Set("parent", o.parent)
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

		if o.creator != nil {
			s.Set("creator", o.creator)
		}

		/* NOTE:
				 * "created" and "modified" time stamps are set by database on insert / update
		     * so are never set in SQL Statement
		*/
		_, e = s.ExecAndClose(context.TODO(), db)
	} else {
		s := sqlf.Update("actions").
			Set("state", o.state).
			Where("guid = ?", o.guid)

		_, e = s.ExecAndClose(context.TODO(), db)
	}

	if e == nil {
		o.stored = true
		o.dirty = false
	}
	return e
}

func (o *Action) reset() {
	// Clean Entry
	o.guid = ""
	o.parent = ""
	o.atype = ""
	o.parameters.Reset()
	o.properties.Reset()
	o.creator = nil
	o.created = nil

	// Mark State as Unregistered
	o.stored = false

	// Mark Entry as Clean
	o.dirty = false
}
