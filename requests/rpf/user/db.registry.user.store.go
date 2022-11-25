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
	"github.com/objectvault/api-services/common"
	"github.com/objectvault/api-services/orm"
	"github.com/objectvault/api-services/orm/query"
	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

func DBUserStoresList(r rpf.GINProcessor, c *gin.Context) {
	// Get Required Parameters
	user_id := r.MustGet("user-id").(uint64)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to User Registry
	db, err := dbm.Connect(user_id)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// List Registered User Orgs
	q := r.MustGet("query-conditions").(*query.QueryConditions)
	list, err := orm.UserObjectsByTypeQuery(db, user_id, common.OTYPE_STORE, q, true)

	// Failed Retrieving User?
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Save List
	r.Set("registry-user-objects", list)
}

func DBSingleShardUsersObjectDeleteAll(r rpf.GINProcessor, c *gin.Context) {
	// Get Required Parameters
	rid := r.MustGet("reference-id").(uint64)
	oid := r.MustGet("object-id").(uint64)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Reference IDs Group and Shard
	db, e := dbm.Connect(rid)
	if e != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Delete Entry
	_, e = orm.UserObjectsDeleteAll(db, oid)
	if e != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

}
