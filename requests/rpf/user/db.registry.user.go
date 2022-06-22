// cSpell:ignore goginrpf, gonic, paulo ferreira
package user

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
	"github.com/objectvault/api-services/orm"
	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

func DBRegistryUserList(r rpf.GINProcessor, c *gin.Context) {
	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to User Registry
	db, err := dbm.ConnectTo(0, 0)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// TODO Add Query Options to Requrest
	// Get User Identifier (GLOBAL ID)
	// query := r.Get("query").(string)

	// List Registered User
	q := r.MustGet("query-conditions").(*orm.QueryConditions)
	users, err := orm.QueryRegisteredUsers(db, q, true)

	// Failed Retrieving User?
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Save List
	r.Set("users", users)
}

func DBRegistryUserFind(r rpf.GINProcessor, c *gin.Context) {
	// Get User Identifier
	id := r.MustGet("user")

	// Is ID and Integer?
	_, is_iid := id.(uint64)
	if is_iid {
		r.Set("user-id", r.MustGet("user"))
		DBRegistryUserFindByID(r, c)
		return
	}

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to User Registry
	db, err := dbm.ConnectTo(0, 0)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Get User based on Type
	entry := &orm.UserRegistry{}
	err = entry.Find(db, id)

	// Failed Retrieving User?
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Did we find the User?
	if !entry.IsValid() { // NO: User does not exist
		r.Abort(4000, nil)
		return
	}

	// Save User
	r.SetLocal("registry-user", entry)
}

func DBRegistryUserFindByID(r rpf.GINProcessor, c *gin.Context) {
	// Get User Identifier
	id := r.Get("user-id").(uint64)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to User Registry
	db, err := dbm.ConnectTo(0, 0)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Get User based on Type
	entry := &orm.UserRegistry{}
	err = entry.ByID(db, id)

	// Failed Retrieving User?
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Did we find the User?
	if !entry.IsValid() { // NO: User does not exist
		r.Abort(4000, nil)
		return
	}

	// Save User
	r.SetLocal("registry-user", entry)
}

func DBRegistryUserFindByEmailOrNil(r rpf.GINProcessor, c *gin.Context) {
	// Get User Identifier
	email := r.MustGet("user-email").(string)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to User Registry
	db, err := dbm.ConnectTo(0, 0)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Get User based on Email
	entry := &orm.UserRegistry{}
	err = entry.ByEmail(db, email)

	// Failed Retrieving User?
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Did we find the User?
	if entry.IsValid() { // YES
		r.SetLocal("registry-user", entry)
	}
}

func DBRegistryUserFindByEmail(r rpf.GINProcessor, c *gin.Context) {
	// Try to Find USer by Email
	DBRegistryUserFindByEmailOrNil(r, c)

	// Did we find the User?
	if r.Aborted() || !r.HasLocal("registry-user") { // NO: User does not exist
		r.Abort(4000, nil)
	}
}

func DBRegisterUser(r rpf.GINProcessor, c *gin.Context) {
	// Get User Information
	u := r.MustGet("user").(*orm.User)
	id := r.MustGet("user-id").(uint64)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Create Registry Entry
	o, err := orm.UserRegistryFromUser(u)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Save User Global ID
	o.SetID(id)

	// Connect to Global Registry Shard
	db, err := dbm.ConnectTo(0, 0)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Save Registry Entry
	err = o.Flush(db, true)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Save Entry
	r.SetLocal("registry-user", o)
}

func DBRegistryUserUpdate(r rpf.GINProcessor, c *gin.Context) {
	// Get User Registry Entry
	e := r.MustGet("user").(*orm.UserRegistry)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Connect to Global Registry Shard
	db, err := dbm.ConnectTo(0, 0)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Save Modifications
	err = e.Flush(db, true)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}
}

func DBRegistryUserUpdateFromUser(r rpf.GINProcessor, c *gin.Context) {
	// Get Organization Information
	user := r.MustGet("user").(*orm.User)

	// Get Organization Registry Entry
	e := r.MustGet("registry-user").(*orm.UserRegistry)

	// Do we need to Update the Organization Registry?
	if user.UpdateRegistry() { // YES
		e.SetUserName(user.UserName())
		e.SetName(user.Name())
	}

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Connect to Global Registry Shard
	db, err := dbm.ConnectTo(0, 0)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Save Organization
	err = e.Flush(db, true)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}
}
