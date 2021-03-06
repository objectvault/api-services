// cSpell:ignore goginrpf, gonic, paulo ferreira
package entry

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
	"github.com/objectvault/api-services/common"
	"github.com/objectvault/api-services/orm"

	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

func DBStoreObjectList(r rpf.GINProcessor, c *gin.Context) {
	// Get Identifier's
	sid := r.MustGet("store-id").(uint64)
	pid := r.MustGet("store-parent-id").(uint32)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Store Shard
	db, err := dbm.Connect(sid)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// List Registered Orgs
	q := r.MustGet("query-conditions").(*orm.QueryConditions)
	objs, err := orm.QueryStoreParentObjects(db, common.LocalIDFromID(sid), pid, q, true)

	// Failed Retrieving User?
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Save List
	r.Set("store-objects", objs)
}

func DBGetStoreObjectByID(r rpf.GINProcessor, c *gin.Context) {
	// Get User Identifier (GLOBAL ID)
	sid := r.MustGet("store-id").(uint64)
	oid := r.MustGet("request-entry-id").(uint32)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Store Shard
	db, e := dbm.Connect(sid)
	if e != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Get User based on Type
	obj := &orm.StoreObject{}

	// Failed Retrieving User?
	e = obj.ByID(db, oid)
	if e != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Did we find the Object?
	if obj.IsNew() { // NO: Object Does not Exist
		r.Abort(4000, nil)
		return
	}

	// Save Store
	r.SetLocal("store-object", obj)
}

func DBInsertStoreObject(r rpf.GINProcessor, c *gin.Context) {
	// Get Organization ID
	obj := r.MustGet("store-object").(*orm.StoreObject)
	sid := r.MustGet("store-id").(uint64)

	// User ID of Creator
	uid := r.MustGet("user-id").(uint64)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Store Shard
	db, e := dbm.Connect(sid)
	if e != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Set Object Creator
	obj.SetCreator(uid)

	// Save Object
	e = obj.Flush(db, true)
	if e != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}
}

func DBUpdateStoreObject(r rpf.GINProcessor, c *gin.Context) {
	// Get Organization ID
	obj := r.MustGet("store-object").(*orm.StoreObject)
	sid := r.MustGet("store-id").(uint64)

	// User ID of Creator
	uid := r.MustGet("user-id").(uint64)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Store Shard
	db, e := dbm.Connect(sid)
	if e != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Set Object Creator
	obj.SetModifier(uid)

	// Save Object
	e = obj.Flush(db, true)
	if e != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}
}

func DBDeleteStoreObject(r rpf.GINProcessor, c *gin.Context) {
	// Get Organization ID
	sid := r.MustGet("store-id").(uint64)
	obj := r.MustGet("store-object").(*orm.StoreObject)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Store Shard
	db, e := dbm.Connect(sid)
	if e != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	switch obj.Type() {
	case orm.OBJECT_TYPE_FOLDER:
		orm.DeleteStoreObjectFolder(db, common.LocalIDFromID(sid), obj.ID())
	default:
		orm.DeleteStoreObject(db, common.LocalIDFromID(sid), obj.ID())
	}
}
