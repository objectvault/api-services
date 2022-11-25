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
	"github.com/objectvault/api-services/orm/query"
	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

func DBRegistryUserObjectsList(r rpf.GINProcessor, c *gin.Context) {
	// Get Required Parameters
	user_id := r.MustGet("user-id").(uint64)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Connect to User Shard
	db, err := dbm.Connect(user_id)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// List Registered User Orgs
	q := r.MustGet("query-conditions").(*query.QueryConditions)
	list, err := orm.UserObjectsQuery(db, user_id, q, true)

	// Failed Retrieving User?
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Save List
	r.Set("registry-user-objects", list)
}

func DBRegistryUserObjFindOrNil(r rpf.GINProcessor, c *gin.Context) {
	// Get User ID
	user := r.MustGet("user-id").(uint64)

	// Get Object ID
	object := r.MustGet("object-id").(uint64)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Connect to User Shard
	db, err := dbm.Connect(user)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Get Org User Entry
	e := &orm.UserObjectRegistry{}
	err = e.ByKey(db, user, object)

	// Failed Retrieving Entry?
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Does Link Exist?
	if !e.IsNew() { // NO: YES
		// Save Entry
		r.SetLocal("registry-user-object", e)
	}
}

func DBRegistryUserObjFind(r rpf.GINProcessor, c *gin.Context) {
	DBRegistryUserObjFindOrNil(r, c)

	if r.Aborted() || !r.HasLocal("registry-user-object") {
		r.Abort(5998, nil) // TODO: Error [User not Registered with Organization]
		return
	}
}

func DBRegistryUserObjFlush(r rpf.GINProcessor, c *gin.Context) {
	// Get User Registry Entry
	e := r.MustGet("registry-user-object").(*orm.UserObjectRegistry)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Connect to User Shard
	db, err := dbm.Connect(e.User())
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

// ORGANIZATION RELATED FUNCTIONS //

func DBRegistryUserOrgFindOrNil(r rpf.GINProcessor, c *gin.Context) {
	// Get Object ID
	r.SetLocal("object-id", r.MustGet("org-id").(uint64))
	DBRegistryUserObjFindOrNil(r, c)
}

func DBRegistryUserOrgFind(r rpf.GINProcessor, c *gin.Context) {
	DBRegistryUserOrgFindOrNil(r, c)

	if r.Aborted() || !r.HasLocal("registry-user-org") {
		r.Abort(5998, nil) // TODO: Error [Org NOT Registered with User]
		return
	}
}

func DBRegisterOrgWithUser(r rpf.GINProcessor, c *gin.Context) {
	// Get Organization Entry
	user_id := r.MustGet("user-id").(uint64)
	org_id := r.MustGet("org-id").(uint64)

	// Get Org Alias from Registry Object or Organization Object
	var orgAlias string
	if r.Has("registry-org") {
		org := r.MustGet("registry-org").(*orm.OrgRegistry)
		orgAlias = org.Alias()
	} else {
		// Get Organization Entry
		org := r.MustGet("org").(*orm.Organization)
		orgAlias = org.Alias()
	}

	// Create ORG Entry
	e := &orm.UserObjectRegistry{}
	e.SetKey(user_id, org_id)
	e.SetAlias(orgAlias)

	// Save Entry
	r.SetLocal("registry-user-object", e)

	// Flush Entry
	DBRegistryUserObjFlush(r, c)
}

func DBDeleteRegistryUserObject(r rpf.GINProcessor, c *gin.Context) {
	// Get Entry Coordinate
	uid := r.MustGet("registry-user-id").(uint64)
	oid := r.MustGet("registry-object-id").(uint64)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Connect to User Registry Shard
	db, e := dbm.Connect(uid)
	if e != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Delete Entry
	_, e = orm.UserObjectDelete(db, uid, oid)
	if e != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}
}
