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

func DBGetStoreByID(r rpf.GINProcessor, c *gin.Context) {
	// Store Entry Already Exists?
	if r.Has("store") { // YES: Do Nothing
		return
	}

	// Get Identifier
	id := r.MustGet("store-id").(uint64)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Organization's Shard
	db, err := dbm.Connect(id)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Get Entity Based on Shard ID
	entry := &orm.Store{}
	err = entry.ByID(db, common.LocalIDFromID(id))

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

func DBDeleteStoreByID(r rpf.GINProcessor, c *gin.Context) {
	/*
		// Get Identifier
		id := r.MustGet("store-id").(uint64)

		// Get Database Connection Manager
		dbm := c.MustGet("dbm").(*orm.DBSessionManager)

		// Get Connection to Organization's Shard
		db, err := dbm.Connect(id)
		if err != nil { // YES: Database Error
			r.Abort(5100, nil)
			return
		}
	*/
	// TODO: IMPLEMENT
	r.Abort(5999, nil)
}

func DBInsertStore(r rpf.GINProcessor, c *gin.Context) {
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
	r.SetLocal("store-id", store_id)
}

func DBUpdateStore(r rpf.GINProcessor, c *gin.Context) {
	// Store ID
	store_id := r.MustGet("store-id").(uint64)

	// Get Store
	store := r.MustGet("store").(*orm.Store)

	// User ID of Modifier
	user_id := r.MustGet("user-id").(uint64)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Shard
	db, err := dbm.Connect(store_id)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Set Modifier
	store.SetModifier(user_id)

	// Save Store
	err = store.Flush(db, false)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}
}
