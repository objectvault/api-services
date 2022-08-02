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
	"github.com/objectvault/api-services/orm/query"

	"github.com/gin-gonic/gin"

	rpf "github.com/objectvault/goginrpf"
)

func DBRegistryInviteList(r rpf.GINProcessor, c *gin.Context) {
	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Global Registry (Always in Group 0: Shard 0)
	db, err := dbm.ConnectTo(0, 0)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// List Invitations
	q := r.MustGet("query-conditions").(*query.QueryConditions)
	invites, err := orm.QueryInvitations(db, q, true)

	// Failed Retrieving List?
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Save List
	r.Set("invitations", invites)
}

func DBRegisterInvitation(r rpf.GINProcessor, c *gin.Context) {
	// Get Invitation
	inv := r.MustGet("invitation").(*orm.Invitation)

	// Invitation into Object ID
	object_id := inv.Object()

	// Get Shard Group and Shard ID from Creation User
	group := common.ShardGroupFromID(object_id)
	shard := common.ShardFromID(object_id)

	// Create Invitation Registry from Invitation
	entry, err := orm.InvitationToRegistry(inv, group, shard)
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

	r.Set("registry-invitation", entry)
}

func DBGetRegistryInvitationByUID(r rpf.GINProcessor, c *gin.Context) {
	// Get Invitation UID
	uid := r.MustGet("request-invite-uid").(string)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Invitation Registry
	db, err := dbm.ConnectTo(0, 0)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Get Entry by UID
	entry := &orm.InvitationRegistry{}
	err = entry.ByUID(db, uid)

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
	r.SetLocal("registry-invitation", entry)
}

func DBGetRegistryInvitationByID(r rpf.GINProcessor, c *gin.Context) {
	// Get Invitation ID
	id := r.MustGet("invitation-id").(uint64)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Registry
	db, err := dbm.Connect(id)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Get Entry by UID
	entry := &orm.InvitationRegistry{}
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
	r.SetLocal("registry-invitation", entry)
}

func DBRegistryInvUpdate(r rpf.GINProcessor, c *gin.Context) {
	// Get Registry Entry
	e := r.MustGet("registry-invitation").(*orm.InvitationRegistry)

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

func DBInvitationAccepted(r rpf.GINProcessor, c *gin.Context) {
	// Get Registry Entry
	e := r.MustGet("registry-invitation").(*orm.InvitationRegistry)

	// Mark Registry as Accepted
	e.SetAccepted()

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Connect to Global Registry Shard
	db, err := dbm.ConnectTo(0, 0)
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

func DBInvitationDeclined(r rpf.GINProcessor, c *gin.Context) {
	// Get Registry Entry
	e := r.MustGet("registry-invitation").(*orm.InvitationRegistry)

	// Mark Registry as Declined
	e.SetDeclined()

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Connect to Global Registry Shard
	db, err := dbm.ConnectTo(0, 0)
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

func DBInvitationRevoked(r rpf.GINProcessor, c *gin.Context) {
	// Get Registry Entry
	e := r.MustGet("registry-invitation").(*orm.InvitationRegistry)

	// Mark Registry as Declined
	e.SetRevoked()

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Connect to Global Registry Shard
	db, err := dbm.ConnectTo(0, 0)
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
