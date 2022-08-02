package action

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
	"github.com/objectvault/api-services/orm/action"

	"github.com/gin-gonic/gin"

	rpf "github.com/objectvault/goginrpf"
)

func DBRegisterAction(r rpf.GINProcessor, c *gin.Context) {
	// Get Request
	oa := r.MustGet("action").(*action.Action)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Global Registry (Always in Group 0: Shard 0)
	db, err := dbm.ConnectTo(0, 0)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Save Entry
	err = oa.Flush(db, true)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}
}

func DBMarkActionQueued(r rpf.GINProcessor, c *gin.Context) {
	// Get Request
	oa := r.MustGet("action").(*action.Action)

	// Mark Action as Queued
	oa.SetStateQueued()

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Global Registry (Always in Group 0: Shard 0)
	db, err := dbm.ConnectTo(0, 0)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Update Action
	err = oa.Flush(db, true)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}
}

func DBGetActionByGUID(r rpf.GINProcessor, c *gin.Context) {
	// Get Action GUID
	guid := r.MustGet("action-guid").(string)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Registry
	db, err := dbm.ConnectTo(0, 0)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Get Entry by UID
	entry := &action.Action{}
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
	r.SetLocal("action", entry)
}
