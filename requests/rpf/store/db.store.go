// cSpell:ignore goginrpf, gonic, paulo ferreira
package store

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
	"fmt"

	rpf "github.com/objectvault/goginrpf"

	"github.com/objectvault/api-services/common"
	"github.com/objectvault/api-services/orm"

	"github.com/gin-gonic/gin"
)

func DBStoreGetByID(r rpf.GINProcessor, c *gin.Context) {
	// Store Entry Already Exists?
	if r.Has("store") { // YES: Do Nothing
		return
	}

	// Get Identifier
	sid := r.MustGet("request-store").(uint64)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Organization's Shard
	db, err := dbm.Connect(sid)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Get Entity Based on Shard ID
	entry := &orm.Store{}
	err = entry.ByID(db, common.LocalIDFromID(sid))

	// Failed Retrieve?
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Did we find the Store?
	if !entry.IsValid() { // NO: Store does not exist
		r.Abort(4998 /* TODO: Error [Store does not exist] */, nil)
		return
	}

	// Save Store
	r.SetLocal("store", entry)
}

func DBStoreMarkDeletedByID(r rpf.GINProcessor, c *gin.Context) {
	// Get Identifier
	sid := r.MustGet("request-store").(uint64)
	uid := r.MustGet("user-id").(uint64)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Store's Shard
	db, e := dbm.Connect(sid)
	if e != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Mark Store Deleted
	b, e := orm.StoreMarkDeleted(db, uid, common.LocalIDFromID(sid))
	if e != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Does Store Exist?
	if !b { // NO: Abort
		r.Abort(4200, nil)
	}
}

func DBStoreDeleteByID(r rpf.GINProcessor, c *gin.Context) {
	// Get Identifier
	sid := r.MustGet("request-store").(uint64)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Store's Shard
	db, e := dbm.Connect(sid)
	if e != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Delete Store Entry
	_, e = orm.StoreDelete(db, common.LocalIDFromID(sid))
	if e != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}
}

func DBStoreInsert(r rpf.GINProcessor, c *gin.Context) {
	// Get Store
	store := r.MustGet("store").(*orm.Store)

	// User ID of Creator
	user_id := r.MustGet("user-id").(uint64)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Set Store Creator
	store.SetCreator(user_id)

	// Get Random Shard
	shard := common.RandomShardID()

	// Get Connection to Random Shard
	db, err := dbm.ConnectTo(1, shard)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Save Store
	err = store.Flush(db, true)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Calculate Global Store ID
	store_id := common.ShardGlobalID(1, common.OTYPE_STORE, shard, store.ID())
	h := fmt.Sprintf("%X", store_id)
	fmt.Printf("STORE ID [%s]\n", h)

	// Save Shard Information and Organization Entry
	r.SetLocal("request-store", store_id)
}

func DBStoreUpdate(r rpf.GINProcessor, c *gin.Context) {
	// Get Identifier
	sid := r.MustGet("request-store").(uint64)

	// Get Store
	store := r.MustGet("store").(*orm.Store)

	// User ID of Modifier
	uid := r.MustGet("user-id").(uint64)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Shard
	db, err := dbm.Connect(sid)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Set Modifier
	store.SetModifier(uid)

	// Save Store
	err = store.Flush(db, false)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}
}
