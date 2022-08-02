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
	"github.com/objectvault/api-services/orm/request"

	"github.com/gin-gonic/gin"

	rpf "github.com/objectvault/goginrpf"
)

func DBGetRequestFromRegistry(r rpf.GINProcessor, c *gin.Context) {
	// Get Request Identifier
	reg := r.MustGet("registry-request").(*request.RequestRegistry)

	// Save Request Global ID
	r.SetLocal("request-id", reg.ID())

	// Get Request By ID
	DBGetRequestByID(r, c)
}

func DBGetRequestByID(r rpf.GINProcessor, c *gin.Context) {
	// Get Invitation Identifier
	id := r.MustGet("request-id").(uint64)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Shard
	db, err := dbm.Connect(id)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Get Request based on ID Type
	entry := &request.Request{}
	err = entry.ByID(db, common.LocalIDFromID(id))

	// Failed Retrieving Entry?
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Did we find the Entry?
	if !entry.IsValid() { // NO: Entry does not exist
		r.Abort(4998 /* TODO: Error [Invalid Request] */, nil)
		return
	}

	// Save Entry
	r.SetLocal("request", entry)
}

func DBInsertRequest(r rpf.GINProcessor, c *gin.Context) {
	// Get Invitation
	req := r.MustGet("request").(*request.Request)

	// Request's Object ID
	object_id := req.Object()

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Object's Shard
	db, err := dbm.Connect(object_id)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Save Request
	err = req.Flush(db, true)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}
}
