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
	"fmt"
	"log"
	"time"

	"github.com/pjacferreira/sqlf"

	"github.com/objectvault/api-services/common"
)

// Invitation Registry Object Definition
type InvitationRegistry struct {
	dirty         bool       // Is Entry Dirty?
	stored        bool       // Is Entry Stored in Database
	id            *uint64    // KEY: GLOBAL Invation ID
	uid           string     // Unique ID (NO Shard or Local ID Information since Invitation Don't Require Session)
	creator       *uint64    // Creator User ID
	object        *uint64    // Invitation to Object
	invitee_email string     // Invitee Email
	expiration    *time.Time // Expiration Time Stamp
	state         uint16     // Invite State (0-active, 1-accepted, 2-declined, 3-Revoked)
}

func InvitationToRegistry(i *Invitation, group uint16, shard uint32) (*InvitationRegistry, error) {
	// Validate Invitation (Can Only Create Registry after Invitation Stored)
	if !i.IsValid() || i.IsNew() {
		return nil, errors.New("Invalid Invitation Object")
	}

	// Create Invitation Registry Entry
	ir := &InvitationRegistry{
		uid:           i.UID(),
		invitee_email: i.InviteeEmail(),
		expiration:    i.Expiration(),
	}

	ir.setID(group, shard, i.ID())
	ir.setCreator(i.Creator())
	ir.setObject(i.Object())
	return ir, nil
}

func HasPendingInvitation(db *sql.DB, object uint64, invitee_email string) (bool, error) {
	if object == 0 || invitee_email == "" {
		return false, errors.New("Inviation Missing Required Values")
	}

	// Query Results Values
	var count uint64

	// Create SQL Statement
	s := sqlf.From("registry_invites").
		Select("COUNT(*)").To(&count).
		Where("id_object = ? and invitee_email = ?", object, invitee_email)

	// DEBUG: Print SQL
	fmt.Println(s.String())

	// Execute Count
	e := s.QueryRowAndClose(context.TODO(), db)

	// Error Occurred?
	if e != nil { // YES
		log.Printf("query error: %v\n", e)
		return false, e
	}

	return count > 0, nil
}

func CountInvitations(db *sql.DB, q TQueryConditions) (uint64, error) {
	// Query Results Values
	var count uint64

	// Create SQL Statement
	s := sqlf.From("registry_invites").
		Select("COUNT(*)").To(&count)

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

func QueryInvitations(db *sql.DB, q TQueryConditions, c bool) (TQueryResults, error) {
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
	var id, creator, object uint64
	var uid, invitee_email, expiration string
	var state uint16
	// Create SQL Statement
	s := sqlf.From("registry_invites").
		Select("id_invite").To(&id).
		Select("uid").To(&uid).
		Select("id_object").To(&object).
		Select("id_creator").To(&creator).
		Select("invitee_email").To(&invitee_email).
		Select("expiration").To(&expiration).
		Select("state").To(&state)

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
		list.AppendSort("uid", false)
		s.OrderBy("uid")
	}

	// DEBUG: Print SQL
	fmt.Println(s.String())

	// Execute Query
	e = s.QueryAndClose(nil, db, func(row *sql.Rows) {
		id_invite := id
		id_object := object
		id_creator := creator

		o := InvitationRegistry{
			id:            &id_invite,
			uid:           uid,
			object:        &id_object,
			creator:       &id_creator,
			expiration:    mySQLTimeStampToGoTime(expiration),
			invitee_email: invitee_email,
			state:         state,
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
		count, e := CountInvitations(db, q)
		if e != nil {
			return nil, e
		}

		list.SetMaxCount(count)
	}

	return list, nil
}

// IsDirty Have the Object Properties Changed since last Serialization?
func (o *InvitationRegistry) IsDirty() bool {
	return o.dirty
}

func (o *InvitationRegistry) IsNew() bool {
	return !o.stored
}

func (o *InvitationRegistry) IsValid() bool {
	return o.uid != "" && o.id != nil && o.creator != nil && o.object != nil && o.invitee_email != ""
}

func (o *InvitationRegistry) IsActive() bool {
	return o.state == 0
}

func (o *InvitationRegistry) IsExpired() bool {
	expires := o.expiration.Unix()
	now := time.Now().Unix()
	return now > expires
}

func (o *InvitationRegistry) HasPending(db *sql.DB) (bool, error) {
	if o.object == nil || o.invitee_email == "" {
		return false, errors.New("Inviation Missing Required Values")
	}

	return HasPendingInvitation(db, *o.object, o.invitee_email)
}

// ByUID Finds Entry By Unique ID String
func (o *InvitationRegistry) ByUID(db *sql.DB, uid string) error {
	// Reset Entry
	o.reset()

	// Execute Query
	var expiration sql.NullString
	e := sqlf.From("registry_invites").
		Select("id_invite").To(&o.id).
		Select("id_creator").To(&o.creator).
		Select("id_object").To(&o.object).
		Select("invitee_email").To(&o.invitee_email).
		Select("expiration").To(&expiration).
		Select("state").To(&o.state).
		Where("uid = ?", uid).
		QueryRowAndClose(context.TODO(), db)

	// Error Executing Query?
	if e != nil && e != sql.ErrNoRows { // YES
		log.Printf("query error: %v\n", e)
		return e
	}

	// Did we retrieve an entry?
	if e == nil { // YES
		o.uid = uid

		if expiration.Valid {
			o.expiration = mySQLTimeStampToGoTime(expiration.String)
		}

		o.stored = true
	}

	return nil
}

// ByID Finds Entry By ID
func (o *InvitationRegistry) ByID(db *sql.DB, id uint64) error {
	// Reset Entry
	o.reset()

	// Execute Query
	var expiration sql.NullString
	e := sqlf.From("registry_invites").
		Select("uid").To(&o.uid).
		Select("id_creator").To(&o.creator).
		Select("id_object").To(&o.object).
		Select("invitee_email").To(&o.invitee_email).
		Select("expiration").To(&expiration).
		Select("state").To(&o.state).
		Where("id_invite = ?", id).
		QueryRowAndClose(context.TODO(), db)

	// Error Executing Query?
	if e != nil && e != sql.ErrNoRows { // YES
		log.Printf("query error: %v\n", e)
		return e
	}

	// Did we retrieve an entry?
	if e == nil { // YES
		o.id = &id

		if expiration.Valid {
			o.expiration = mySQLTimeStampToGoTime(expiration.String)
		}

		o.stored = true
	}

	return nil
}

func (o *InvitationRegistry) ID() uint64 {
	if o.id == nil {
		return 0
	}

	return *o.id
}

func (o *InvitationRegistry) UID() string {
	return o.uid
}

func (o *InvitationRegistry) Creator() uint64 {
	if o.creator == nil {
		return 0
	}

	return *o.creator
}

func (o *InvitationRegistry) Object() uint64 {
	if o.object == nil {
		return 0
	}

	return *o.object
}

func (o *InvitationRegistry) InviteeEmail() string {
	return o.invitee_email
}

func (o *InvitationRegistry) Expiration() *time.Time {
	return o.expiration
}

func (o *InvitationRegistry) ExpirationUTC() string {
	utc := o.expiration.UTC()

	// RETURN ISO 8601 / RFC 3339 FORMAT in UTC
	return utc.Format(time.RFC3339)
}

func (o *InvitationRegistry) State() uint16 {
	return o.state
}

func (o *InvitationRegistry) SetAccepted() uint16 {
	current := o.state

	if current == 0 {
		o.state = 1
		o.dirty = true
	}

	return current
}

func (o *InvitationRegistry) SetDeclined() uint16 {
	current := o.state

	if current == 0 {
		o.state = 2
		o.dirty = true
	}

	return current
}

func (o *InvitationRegistry) SetRevoked() uint16 {
	current := o.state

	if current == 0 {
		o.state = 3
		o.dirty = true
	}

	return current
}

func (o *InvitationRegistry) setID(group uint16, shard uint32, lid uint32) uint64 {
	id := common.ShardGlobalID(group, common.OTYPE_INVITATION, shard, lid)

	current := uint64(0)
	if o.id != nil {
		current = *o.id
	}

	o.id = &id
	return current
}

func (o *InvitationRegistry) setCreator(id uint64) (uint64, error) {
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

func (o *InvitationRegistry) setObject(id uint64) (uint64, error) {
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

func (o *InvitationRegistry) setInviteeEmail(email string) (string, error) {
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
	o.invitee_email = email
	o.dirty = true
	return current, nil
}

func (o *InvitationRegistry) SetState(state uint16) error {
	// Current State
	current := state

	// State Changed?
	if current != state { // YES - Update
		o.state = state
		o.dirty = true
	}

	return nil
}

func (o *InvitationRegistry) SetExpiration(t time.Time) error {
	o.expiration = &t
	o.dirty = true

	return nil
}

func (o *InvitationRegistry) SetExpiresIn(days uint16) error {
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

func (o *InvitationRegistry) Flush(db sqlf.Executor, force bool) error {
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
		s := sqlf.InsertInto("registry_invites").
			Set("id_invite", o.id).
			Set("uid", o.uid).
			Set("id_creator", o.creator).
			Set("id_object", o.object).
			Set("invitee_email", o.invitee_email).
			Set("expiration", goTimeToMySQLTimeStamp(o.expiration)).
			Set("state", o.state)

		_, e = s.Exec(context.TODO(), db)
	} else { // NO: Update
		_, e = sqlf.Update("registry_invites").
			Set("state", o.state).
			Where("id_invite = ?", o.id).
			ExecAndClose(context.TODO(), db)
	}

	if e == nil {
		o.stored = true
		o.dirty = false
	}
	return e
}

func (o *InvitationRegistry) reset() {
	// Clean Entry
	o.uid = ""
	o.id = nil
	o.creator = nil
	o.invitee_email = ""
	o.state = 0

	// Mark State as Unregistered
	o.stored = false

	// Mark Entry as Clean
	o.dirty = false
}
