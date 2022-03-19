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
	"crypto/sha1"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/pjacferreira/sqlf"
)

// Invitation Object Definition
type Invitation struct {
	dirty         bool       // Is Entry Dirty?
	stored        bool       // Is Entry Stored in Database
	id            *uint32    // LOCAL Organization ID
	uid           string     // Unique ID (NO Shard or Local ID Information since Invitation Don't Require Session)
	creator       *uint64    // GLOBAL Creator User ID
	invitee_email string     // Invitee Email
	object        *uint64    // GLOBAL Object ID
	message       string     // Invitation Message
	key           *uint64    // GLOBAL Key ID
	key_pick      []byte     // Key Object Unlocker
	expiration    *time.Time // Expiration Time Stamp
	created       *time.Time // Creation Timestamp
	roles         S_Roles    // User Roles in Organization
}

// TODO Implement Delete (Both From Within an Entry and Without a Structure)
/* IMPORTANT NOTE:
 * MySQL stores timestamps in UTC, but serves them in local time, as set
 * on the MySQL Server HOST NODE.
 * This MEANS that the NODE on which the GO is Being Run has to be ON THE
 * SAME time zone setting as the MySQL HOST NODE
 */
func mySQLTimeStampToGoTime(t string) *time.Time {
	// Parse MySQL TimeStamp
	tm, e := time.Parse("2006-01-02 15:04:05", t)

	// NOTE: MySQL Server Should have Time Zone set to UTC
	// time.Parse() Returns time stamp relative to UTC if no Time Zone in String

	// Did parse genersate an error?
	if e != nil { // YES
		return nil
	}

	// Return Time Stam
	fmt.Println(tm)
	return &tm
}

func goTimeToMySQLTimeStamp(t *time.Time) string {
	utc := t.UTC()
	formatted := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
		utc.Year(), utc.Month(), utc.Day(),
		utc.Hour(), utc.Minute(), utc.Second())
	return formatted
}

// IsDirty Have the Object Properties Changed since last Serialization?
func (o *Invitation) IsDirty() bool {
	return o.dirty
}

func (o *Invitation) IsNew() bool {
	return !o.stored
}

func (o *Invitation) IsValid() bool {
	return o.creator != nil && o.invitee_email != "" && o.object != nil && o.expiration != nil
}

// ByID Finds Entry By ID
func (o *Invitation) ByID(db *sql.DB, id uint32) error {
	// Reset Entry
	o.reset()

	// Execute Query
	var roles sql.NullString
	var message sql.NullString
	var key sql.NullInt64
	var key_pick sql.NullString
	var expiration sql.NullString
	var created sql.NullString
	e := sqlf.From("invites").
		Select("id_creator").To(&o.creator).
		Select("invitee_email").To(&o.invitee_email).
		Select("id_object").To(&o.object).
		Select("roles").To(&roles).
		Select("message").To(&message).
		Select("id_key").To(&key).
		Select("key_pick").To(&key_pick).
		Select("expiration").To(&expiration).
		Select("created").To(&created).
		Where("id = ?", id).
		QueryRowAndClose(nil, db)

	// Error Executing Query?
	if e != nil && e != sql.ErrNoRows { // YES
		log.Printf("query error: %v\n", e)
		return e
	}

	// Did we retrieve an entry?
	if e == nil { // YES
		o.id = &id
		if roles.Valid {
			o.roles.RolesFromCSV(roles.String)
		}
		if message.Valid {
			o.message = message.String
		}
		if key.Valid {
			k := uint64(key.Int64)
			o.key = &k
		}
		if key_pick.Valid {
			s := key_pick.String
			o.key_pick = []byte(s)
		}
		if expiration.Valid {
			o.expiration = mySQLTimeStampToGoTime(expiration.String)
		}
		if created.Valid {
			o.created = mySQLTimeStampToGoTime(created.String)
		}
		// TODO Deal with Object, Creator/ed, Modifier/de
		o.stored = true
	}

	return nil
}

// ID Get User ID
func (o *Invitation) ID() uint32 {
	if o.id == nil {
		return 0
	}

	return *o.id
}

func (o *Invitation) UID() string {
	return o.uid
}

func (o *Invitation) Creator() uint64 {
	if o.creator == nil {
		return 0
	}

	return *o.creator
}

func (o *Invitation) InviteeEmail() string {
	return o.invitee_email
}

func (o *Invitation) Object() uint64 {
	if o.object == nil {
		return 0
	}

	return *o.object
}

func (o *Invitation) Message() string {
	return o.message
}

func (o *Invitation) Key() *uint64 {
	return o.key
}

func (o *Invitation) KeyPick() []byte {
	return o.key_pick
}

func (o *Invitation) Expiration() *time.Time {
	return o.expiration
}

func (o *Invitation) ExpirationUTC() string {
	utc := o.expiration.UTC()
	// RETURN ISO 8601 / RFC 3339 FORMAT in UTC
	return utc.Format(time.RFC3339)
}

func (o *Invitation) Created() *time.Time {
	return o.created
}

func (o *Invitation) SetCreator(id uint64) (uint64, error) {
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

func (o *Invitation) SetInviteeEmail(email string) (string, error) {
	if !o.IsNew() {
		return "", errors.New("Registered Invitation is immutable")
	}

	// Current State
	current := o.invitee_email

	// Validate Invitee Alias
	if email == "" {
		return current, errors.New("Missing Value for Organization Alias")
	}

	// New State
	o.invitee_email = strings.ToLower(email)

	o.dirty = true
	return current, nil
}

func (o *Invitation) SetObject(id uint64) (uint64, error) {
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

func (o *Invitation) SetMessage(m string) (string, error) {
	if !o.IsNew() {
		return "", errors.New("Registered Invitation is immutable")
	}

	// Current State
	current := o.message

	// Roles Changed?
	o.message = m
	o.dirty = true

	return current, nil
}

func (o *Invitation) SetKey(id uint64, pick []byte) error {
	if !o.IsNew() {
		return errors.New("Registered Invitation is immutable")
	}

	if pick == nil {
		return errors.New("Key Missing Lock")
	}

	o.key = &id
	o.key_pick = pick
	return nil
}

func (o *Invitation) SetExpiration(t time.Time) error {
	o.expiration = &t
	o.dirty = true

	return nil
}

func (o *Invitation) SetExpiresIn(days uint16) error {
	// Have DB Connection?
	if days == 0 { // NO: Abort
		return errors.New("Numer of days should be > 0")
	}

	now := time.Now()
	expires := now.AddDate(0, 0, int(days))
	o.expiration = &expires
	o.dirty = true
	return nil
}

// Implementation of I_Roles //
func (o *Invitation) Roles() []uint32 {
	return o.roles.Roles()
}

func (o *Invitation) IsRolesEmpty() bool {
	return o.roles.IsRolesEmpty()
}

func (o *Invitation) HasRole(role uint32) bool {
	return o.roles.HasRole(role)
}

func (o *Invitation) HasExactRole(role uint32) bool {
	return o.roles.HasExactRole(role)
}

func (o *Invitation) AddRole(role uint32) bool {
	modified := o.roles.AddRole(role)
	o.dirty = o.dirty || modified
	return modified
}

func (o *Invitation) AddRoles(roles []uint32) bool {
	modified := o.roles.AddRoles(roles)
	o.dirty = o.dirty || modified
	return modified
}

func (o *Invitation) RemoveRole(role uint32) bool {
	modified := o.roles.RemoveRole(role)
	o.dirty = o.dirty || modified
	return modified
}

func (o *Invitation) RemoveCategory(category uint16) bool {
	modified := o.roles.RemoveCategory(category)
	o.dirty = o.dirty || modified
	return modified
}

func (o *Invitation) RemoveExactRole(role uint32) bool {
	modified := o.roles.RemoveExactRole(role)
	o.dirty = o.dirty || modified
	return modified
}

func (o *Invitation) RemoveRoles(roles []uint32) bool {
	modified := o.roles.RemoveRoles(roles)
	o.dirty = o.dirty || modified
	return modified
}

func (o *Invitation) RemoveAllRoles() bool {
	modified := o.roles.RemoveAllRoles()
	o.dirty = o.dirty || modified
	return modified
}

func (o *Invitation) RolesFromCSV(csv string) bool {
	modified := o.roles.RolesFromCSV(csv)
	o.dirty = o.dirty || modified
	return modified
}

func (o *Invitation) RolesToCSV() string {
	return o.roles.RolesToCSV()
}

func (o *Invitation) Flush(db sqlf.Executor, force bool) error {
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

		// Create Unique ID
		o.uid = o.createUID()

		// Execute Insert
		s := sqlf.InsertInto("invites").
			Set("uid", o.uid).
			Set("id_creator", o.creator).
			Set("invitee_email", o.invitee_email).
			Set("id_object", o.object).
			Set("expiration", goTimeToMySQLTimeStamp(o.expiration))

		if o.key != nil {
			if o.key_pick == nil {
				return errors.New("Key Object Requires lock")
			}
			s.Set("id_key", o.key)
			s.Set("key_pick", o.key_pick)
		}

		if !o.roles.IsRolesEmpty() {
			s.Set("roles", o.RolesToCSV())
		}

		if o.message != "" {
			s.Set("message", o.message)
		}

		_, e = s.Exec(nil, db)

		// Error Occured?
		if e == nil { // NO: Get Last Insert ID
			var id uint32
			e = sqlf.Select("LAST_INSERT_ID()").
				To(&id).
				QueryRowAndClose(nil, db)

			// Error Occured?
			if e == nil { // NO: Set Object ID
				o.id = &id
			}
		}
	} else {
		return errors.New("Invitation Objects are Immutable")
	}

	if e == nil {
		o.stored = true
		o.dirty = false
	}
	return e
}

func (o *Invitation) createUID() string {
	key := fmt.Sprintf("%X:%X:%s:%X", *o.creator, *o.object, o.invitee_email, time.Now().Unix())
	hash := fmt.Sprintf("%x", sha1.Sum([]byte(key)))
	return hash
}

func (o *Invitation) reset() {
	// Clean Entry
	o.id = nil
	o.creator = nil
	o.invitee_email = ""
	o.object = nil
	o.message = ""
	o.expiration = nil
	o.created = nil
	o.key = nil
	o.key_pick = nil

	// Clear Roles
	o.roles.RemoveAllRoles()

	// Mark State as Unregistered
	o.stored = false

	// Mark Entry as Clean
	o.dirty = false
}
