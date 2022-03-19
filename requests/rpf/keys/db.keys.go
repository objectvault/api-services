// cSpell:ignore goginrpf, gonic, paulo ferreira
package keys

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
	"time"

	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
	"github.com/objectvault/api-services/common"
	"github.com/objectvault/api-services/orm"
)

func DBCreateKey(r rpf.GINProcessor, c *gin.Context) {
	// Get User Identifier
	user := r.MustGet("user-id").(uint64)

	// Get Bytes
	bytes := r.MustGet("key-bytes").([]byte)

	// Get Expiration
	expiration := r.MustGet("key-expiration").(*time.Time)

	// Create a Key Object
	key, k, err := orm.NewKey(user, bytes, *expiration)
	if err != nil { // ERROR: Unexpected
		r.Abort(5900, nil)
		return
	}

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Random Shard
	sid := common.RandomShardID()

	// Get Connection to Random Shard
	db, err := dbm.ConnectTo(1, sid)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	err = k.Flush(db, true)

	// Failed Retrieving User?
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Calculate Global Key ID
	kid := common.ShardGlobalID(1, common.OTYPE_KEY, sid, k.ID())
	h := fmt.Sprintf("%X", kid)
	fmt.Printf("KEY ID [%s]\n", h)

	// Save Key Information
	r.SetLocal("key-id", kid)   // Key Global ID
	r.SetLocal("key-key", key)  // Key Decryption Password
	r.SetLocal("key-object", k) // Key Object
}

func DBGetKeyByID(r rpf.GINProcessor, c *gin.Context) {
	// Get Identifier
	id := r.MustGet("key-id").(uint64)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Key's Shard
	db, err := dbm.Connect(id)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Get Entity Based on Shard ID
	key := &orm.Key{}
	err = key.ByID(db, common.LocalIDFromID(id))

	// Failed Retrieve?
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Did we find the Key?
	if !key.IsValid() { // NO: Key does not exist
		r.Abort(4998 /* TODO: Error [Key does not exist] */, nil)
		return
	}

	// Save Key Object
	r.SetLocal("key-object", key)
}
