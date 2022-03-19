// cSpell:ignore goginrpf, gonic, paulo ferreira
package invitation

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

	"github.com/gin-gonic/gin"

	rpf "github.com/objectvault/goginrpf"
)

func DBGetInvitationFromRegistry(r rpf.GINProcessor, c *gin.Context) {
	// Get Invitation Identifier
	reg := r.MustGet("registry-invitation").(*orm.InvitationRegistry)

	// Save Invitation Global ID
	r.SetLocal("invitation-id", reg.ID())

	// Get Invitation By ID
	DBGetInvitationByID(r, c)
}

func DBGetInvitationByID(r rpf.GINProcessor, c *gin.Context) {
	// Get Invitation Identifier
	id := r.MustGet("invitation-id").(uint64)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Shard
	db, err := dbm.Connect(id)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Get Org based on ID Type
	entry := &orm.Invitation{}
	err = entry.ByID(db, common.LocalIDFromID(id))

	// Failed Retrieving User?
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Did we find the Entry?
	if !entry.IsValid() { // NO: Entry does not exist
		r.Abort(4998 /* TODO: Error [Invalid Invitation] */, nil)
		return
	}

	// Save Entry in Registry
	r.SetLocal("invitation", entry)
}

func DBInsertInvitation(r rpf.GINProcessor, c *gin.Context) {
	// Get Invitation
	inv := r.MustGet("invitation").(*orm.Invitation)

	// Invite into Object ID
	object_id := inv.Object()

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Object's Shard
	db, err := dbm.Connect(object_id)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Save Invitation
	err = inv.Flush(db, true)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}
}
