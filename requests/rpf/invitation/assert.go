// cSpell:ignore goginrpf, gonic, orgs, paulo, ferreira
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
	"github.com/gin-gonic/gin"

	"github.com/objectvault/api-services/common"
	"github.com/objectvault/api-services/orm"

	rpf "github.com/objectvault/goginrpf"
)

func AssertInvitationActive(r rpf.GINProcessor, c *gin.Context) {
	// Get Invitation
	invitation := r.MustGet("registry-invitation").(*orm.InvitationRegistry)

	// Is Invitation Still Active?
	if !invitation.IsActive() { // NO
		r.Abort(4390, nil)
		return
	}

	// Has Invitation Expired?
	if invitation.IsExpired() { // YES
		r.Abort(4391, nil)
		return
	}
}

func AssertInvitationNotExpired(r rpf.GINProcessor, c *gin.Context) {
	// Get Invitation
	invitation := r.MustGet("registry-invitation").(*orm.InvitationRegistry)

	// Has Invitation Expired?
	if invitation.IsExpired() { // YES
		r.Abort(4391, nil)
		return
	}
}

func AssertNoPendingInvitation(r rpf.GINProcessor, c *gin.Context) {
	// Get Invitation
	i := r.MustGet("invitation").(*orm.Invitation)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Global Registry (Always in Group 0: Shard 0)
	db, err := dbm.ConnectTo(0, 0)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Are there any pending Invitations
	dup, err := orm.HasPendingInvitation(db, i.Object(), i.InviteeEmail())
	if err != nil { // UNKNOWN: Database Error
		r.Abort(5100, nil)
		return
	}

	if dup { // YES: Pending Invitation Found
		r.Abort(5998, nil) // TODO: Error [Pending Invitation]
		return
	}
}

func AssertSameObjectInvitation(r rpf.GINProcessor, c *gin.Context) {
	// Get Invitation
	invitation := r.MustGet("registry-invitation").(*orm.InvitationRegistry)

	// Object
	object := r.MustGet("object-id").(uint64)

	// Is Invitation in Request Object?
	if invitation.Object() != object { // NO
		r.Abort(4390, nil)
		return
	}
}

func AssertOrgInvitation(r rpf.GINProcessor, c *gin.Context) {
	// Get Invitation
	i := r.MustGet("registry-invitation").(*orm.InvitationRegistry)

	// Is Invitation into Organization?
	if !common.IsObjectOfType(i.Object(), common.OTYPE_ORG) { // NO
		r.Abort(4390, nil)
		return
	}
}

func AssertStoreInvitation(r rpf.GINProcessor, c *gin.Context) {
	// Get Invitation
	i := r.MustGet("registry-invitation").(*orm.InvitationRegistry)

	// Is Invitation into Store?
	if !common.IsObjectOfType(i.Object(), common.OTYPE_STORE) { // NO
		r.Abort(4390, nil)
		return
	}
}
