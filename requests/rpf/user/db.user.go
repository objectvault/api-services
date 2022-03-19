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
	"fmt"

	"github.com/objectvault/api-services/common"
	"github.com/objectvault/api-services/orm"
	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

func DBGetUserByID(r rpf.GINProcessor, c *gin.Context) {
	// Get User Identifier (GLOBAL ID)
	id := r.Get("user-id").(uint64)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to User Registry
	db, err := dbm.Connect(id)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Get User based on Type
	user := &orm.User{}
	err = user.ByID(db, common.LocalIDFromID(id))

	// Failed Retrieving User?
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Did we find the User?
	if user.IsNew() { // NO: User does not exist
		r.Abort(4000, nil)
		return
	}

	// Save User
	r.Set("user", user)
}

func DBInsertUser(r rpf.GINProcessor, c *gin.Context) {
	// Get the User to Create
	u := r.MustGet("user").(*orm.User)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Random Shard
	shard := common.RandomShardID()

	// Get Connection to Random Shard
	db, err := dbm.ConnectTo(1, shard)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Save User
	err = u.Flush(db, true)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Calculate User Sharded ID
	id := common.ShardGlobalID(1, common.OTYPE_USER, shard, u.ID())
	h := fmt.Sprintf("%X", id)
	fmt.Printf("USER ID [%s]\n", h)

	// Save User Sharded ID
	r.SetLocal("user-id", id)
}
