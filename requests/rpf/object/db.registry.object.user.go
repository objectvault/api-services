// cSpell:ignore goginrpf, gonic, paulo ferreira
package object

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

func DBRegistryObjectUsersList(r rpf.GINProcessor, c *gin.Context) {
	// Get Object Identifier
	obj := r.MustGet("object-id").(uint64)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Object Shard
	db, err := dbm.Connect(obj)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// List Registered Org Users
	q := r.MustGet("query-conditions").(*orm.QueryConditions)
	users, err := orm.QueryRegisteredObjectUsers(db, obj, q, true)

	// Failed Retrieving User?
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Save List
	r.Set("registry-object-users", users)
}

func DBRegistryObjectUserFindOrNil(r rpf.GINProcessor, c *gin.Context) {
	// Get Object Identifier
	obj := r.MustGet("object-id").(uint64)

	// Get User ID
	user := r.MustGet("user-id").(uint64)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Object Shard
	db, err := dbm.Connect(obj)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Get Object User Registry Entry
	e := &orm.ObjectUserRegistry{}
	err = e.ByKey(db, obj, user)

	// Failed Retrieving Entry?
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Does Org User Entry Exist?
	if !e.IsNew() { // NO: YES
		// Save Entry
		r.SetLocal("registry-object-user", e)
	}
}

func DBRegistryObjectUserFind(r rpf.GINProcessor, c *gin.Context) {
	DBRegistryObjectUserFindOrNil(r, c)

	if r.Aborted() || !r.HasLocal("registry-object-user") {
		r.Abort(5998, nil) // TODO: Error [User not Registered with Organization]
		return
	}
}

func DBRegistryObjectUserFlush(r rpf.GINProcessor, c *gin.Context) {
	// Get Registry Entry
	e := r.MustGet("registry-object-user").(*orm.ObjectUserRegistry)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Connect to Object Registry Shard
	db, err := dbm.Connect(e.Object())
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Save
	err = e.Flush(db, false)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}
}

func DBRegistryUpdateFromUser(r rpf.GINProcessor, c *gin.Context) {
	// Get Registry Entry
	e := r.MustGet("registry-object-user").(*orm.ObjectUserRegistry)

	// Get User
	u := r.MustGet("user").(*orm.User)

	// Do we need to Update the Registry?
	if u.UpdateRegistry() { // YES
		// Get Database Connection Manager
		dbm := c.MustGet("dbm").(*orm.DBSessionManager)

		// Connect to Registry Shard
		db, err := dbm.Connect(e.Object())
		if err != nil { // YES: Database Error
			r.Abort(5100, nil)
			return
		}

		// Update Registry Fields
		e.SetUserName(u.UserName())

		// Flush Registry
		err = e.Flush(db, false)
		if err != nil { // YES: Database Error
			r.Abort(5100, nil)
			return
		}
	}
}

// ORGANIZATION : ADAPTERS //

func DBRegistryOrgUsersList(r rpf.GINProcessor, c *gin.Context) {
	r.SetLocal("object-id", r.MustGet("org-id"))

	// Redirect to Object Registry
	DBRegistryObjectUsersList(r, c)
}

func DBRegistryOrgUserFind(r rpf.GINProcessor, c *gin.Context) {
	r.SetLocal("object-id", r.MustGet("org-id"))

	// Redirect to Object Registry
	DBRegistryObjectUserFindOrNil(r, c)

	if r.Aborted() || !r.HasLocal("registry-object-user") {
		r.Abort(5998, nil) // TODO: Error [User not Registered with Organization]
		return
	}
}

func DBRegisterUserWithOrg(r rpf.GINProcessor, c *gin.Context) {
	// Get Organization Global ID
	org_id := r.MustGet("org-id").(uint64)

	// Get User Informtaion
	user := r.MustGet("registry-user").(*orm.UserRegistry)

	// Create ORG Entry //
	e := &orm.ObjectUserRegistry{}
	e.SetKey(org_id, user.ID())
	e.SetUserName(user.UserName())

	// Are specific roles to be set?
	if r.Has("register-roles") { // YES
		roles := r.Get("register-roles").([]uint32)
		e.AddRoles(roles)
	}

	// Is User Org Admin?
	if r.Has("register-org-admin") { // YES
		e.SetState(orm.STATE_SYSTEM)
	}

	r.SetLocal("registry-object-user", e)
	DBRegistryObjectUserFlush(r, c)
}
