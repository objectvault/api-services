package request

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
	"github.com/objectvault/api-services/orm/request"

	"github.com/gin-gonic/gin"

	rpf "github.com/objectvault/goginrpf"
)

func DBRegistryRequestsList(r rpf.GINProcessor, c *gin.Context) {
	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Global Registry (Always in Group 0: Shard 0)
	db, err := dbm.ConnectTo(0, 0)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// List Requests
	q := r.MustGet("query-conditions").(*query.QueryConditions)
	requests, err := request.QueryRequests(db, q, true)

	// Failed Retrieving List?
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Save List
	r.Set("requests", requests)
}

func DBRegisterRequest(r rpf.GINProcessor, c *gin.Context) {
	// Get Request
	inv := r.MustGet("request").(*request.Request)

	// Request's Object ID
	object_id := inv.Object()

	// Get Shard Group and Shard ID from Creation User
	group := common.ShardGroupFromID(object_id)
	shard := common.ShardFromID(object_id)

	// Create Invitation Registry from Invitation
	entry, err := request.RequestToRegistry(inv, group, shard)
	if err != nil { // YES: Failed to Create Registry Entry
		r.Abort(5998, nil)
		return
	}

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Global Registry (Always in Group 0: Shard 0)
	db, err := dbm.ConnectTo(0, 0)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Save Entry
	err = entry.Flush(db, true)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	r.Set("registry-request", entry)
}

func DBGetRegistryRequestByGUID(r rpf.GINProcessor, c *gin.Context) {
	// Get Invitation UID
	guid := r.MustGet("request-guid").(string)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Registry
	db, err := dbm.ConnectTo(0, 0)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Get Entry by UID
	entry := &request.RequestRegistry{}
	err = entry.ByGUID(db, guid)

	// Failed Retrieving the Entry?
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Did we find the Entry?
	if !entry.IsValid() { // NO
		r.Abort(4000, nil)
		return
	}

	// Save Entry
	r.SetLocal("registry-request", entry)
}

func DBGetRegistryRequestByID(r rpf.GINProcessor, c *gin.Context) {
	// Get ID
	id := r.MustGet("request-id").(uint64)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Registry
	db, err := dbm.Connect(id)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Get Entry by ID
	entry := &request.RequestRegistry{}
	err = entry.ByID(db, id)

	// Failed Retrieving the Entry?
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Did we find the Entry?
	if !entry.IsValid() { // NO
		r.Abort(4000, nil)
		return
	}

	// Save Entry
	r.SetLocal("registry-request", entry)
}

func DBRegistryRequestUpdate(r rpf.GINProcessor, c *gin.Context) {
	// Get Registry Entry
	e := r.MustGet("registry-request").(*request.RequestRegistry)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Registry
	db, err := dbm.Connect(e.ID())
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Save Modifications
	err = e.Flush(db, false)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}
}
