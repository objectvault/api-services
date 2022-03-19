// cSpell:ignore goginrpf, gonic, paulo ferreira
package object

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
	rpf "github.com/objectvault/goginrpf"

	"github.com/objectvault/api-services/orm"
	"github.com/objectvault/api-services/requests/rpf/shared"
	"github.com/objectvault/api-services/requests/rpf/user"
)

func AddinNoExistingUserRegistration(g rpf.GINGroupProcessor, opts shared.TAddinCallbackOptions) rpf.GINGroupProcessor {
	g.Append(
		func(r rpf.GINProcessor, c *gin.Context) {
			i := r.MustGet("invitation").(*orm.Invitation)
			r.SetLocal("user-email", i.InviteeEmail())
		},
		user.DBRegistryUserFindByEmailOrNil,
		func(r rpf.GINProcessor, c *gin.Context) {
			u := r.Get("registry-user")
			if u != nil {
				// Create Processing Group
				group := &rpf.ProcessorGroup{}
				group.Parent = r

				group.Chain = rpf.ProcessChain{
					func(g rpf.GINProcessor, c *gin.Context) {
						u := g.MustGet("registry-user").(*orm.UserRegistry)
						i := g.MustGet("invitation").(*orm.Invitation)

						// Find User Object Reqgistry
						g.SetLocal("object-id", i.Object())
						g.SetLocal("user-id", u.ID())
					},
					DBRegistryObjectUserFindOrNil,
					func(g rpf.GINProcessor, c *gin.Context) {
						ou := g.Get("registry-object-user")
						if ou != nil { // ABORT: User already registered with Object
							r.Abort(4012, nil)
						}
					},
				}

				group.Run()
			}
		},
	)
	return g
}
