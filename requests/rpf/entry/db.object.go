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

// cSpell:ignore objs
import (
	"github.com/objectvault/api-services/common"
	"github.com/objectvault/api-services/orm"
	"github.com/objectvault/api-services/orm/query"

	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

func DBStoreObjectsList(r rpf.GINProcessor, c *gin.Context) {
	// Get Request Parameters
	sid := r.MustGet("request-store").(uint64)
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
	q := r.MustGet("query-conditions").(*query.QueryConditions)
	objs, err := orm.QueryStoreParentObjects(db, common.LocalIDFromID(sid), pid, q, true)

	// Failed Retrieving User?
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Save List
	r.Set("store-objects", objs)
}

func DBStoreObjectGetByID(r rpf.GINProcessor, c *gin.Context) {
	// Get Request Parameters
	sid := r.MustGet("request-store").(uint64)
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

func DBStoreObjectInsert(r rpf.GINProcessor, c *gin.Context) {
	// Get Request Parameters
	sid := r.MustGet("request-store").(uint64)
	obj := r.MustGet("store-object").(*orm.StoreObject)

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

func DBStoreObjectUpdate(r rpf.GINProcessor, c *gin.Context) {
	// Get Request Parameters
	sid := r.MustGet("request-store").(uint64)
	obj := r.MustGet("store-object").(*orm.StoreObject)

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

func DBStoreObjectDelete(r rpf.GINProcessor, c *gin.Context) {
	// Get Request Parameters
	sid := r.MustGet("request-store").(uint64)
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
		orm.StoreObjectDeleteFolder(db, common.LocalIDFromID(sid), obj.ID())
	default:
		orm.StoreObjectDelete(db, common.LocalIDFromID(sid), obj.ID())
	}
}

func DBStoreObjectsDeleteAll(r rpf.GINProcessor, c *gin.Context) {
	// Get Request Parameters
	sid := r.MustGet("request-store").(uint64)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Store Shard
	db, e := dbm.Connect(sid)
	if e != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	_, e = orm.StoreObjectsDeleteAll(db, common.LocalIDFromID(sid))
	if e != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}
}
