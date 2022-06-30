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
	"github.com/objectvault/api-services/orm"
	"github.com/objectvault/api-services/orm/query"
	"github.com/objectvault/api-services/requests/rpf/shared"

	"github.com/gin-gonic/gin"

	rpf "github.com/objectvault/goginrpf"
)

// INVITATION REGISTRY //

func ExportRegistryInvList(r rpf.GINProcessor, c *gin.Context) {
	// Get Invitations Registry Entries
	invs := r.Get("invitations").(query.TQueryResults)

	irs := &shared.ExportList{
		List: invs,
		ValueMapper: func(v interface{}) interface{} {
			return &RegistryInviteToJSON{
				Entry: v.(*orm.InvitationRegistry),
			}
		},
		FieldMapper: func(o_to_e string) string {
			switch o_to_e {
			case "id_invite":
				return "id"
			case "uid":
				return "uid"
			case "id_org":
				return "org"
			case "id_creator":
				return "creator"
			case "invitee_email":
				return "invitee"
			case "expiration":
				return "expiration"
			case "state":
				return "state"
			}

			// Invalid Field
			return ""
		},
	}

	r.SetResponseDataValue("invitations", irs)
}

func ExportRegistryInv(r rpf.GINProcessor, c *gin.Context) {
	// Get Invitations Registry Entries
	i := r.Get("registry-invitation").(*orm.InvitationRegistry)

	oi := &RegistryInviteToJSON{
		Entry: i,
	}

	r.SetResponseDataValue("invitation", oi)
}

func ExportNoSessionRegistryInv(r rpf.GINProcessor, c *gin.Context) {
	// Get Invitations Registry Entries
	i := r.Get("registry-invitation").(*orm.InvitationRegistry)
	u := r.Get("invitation-creator").(*orm.UserRegistry)

	oi := &NoSessionInviteToJSON{
		Invite:  i,
		Creator: u,
	}

	// Invitee Exists?
	if r.Has("invitation-invitee") { // YES
		oi.Invitee = r.Get("invitation-invitee").(*orm.UserRegistry)
	}

	r.SetResponseDataValue("invitation", oi)
}
