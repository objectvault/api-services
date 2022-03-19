// cSpell:ignore goginrpf, gonic, paulo ferreira
package session

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
	orm "github.com/objectvault/api-services/orm"
	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

func IFCloseSessionOnBlockedUser(parent rpf.GINProcessor, u *orm.UserRegistry) *rpf.ProcessorIF {
	rif := rpf.NestedIF(parent,
		func(r rpf.ProcessorIF, c *gin.Context) {
			// Get Request User
			if !u.IsActive() || u.IsBlocked() { // NO: Verify if User Exists
				r.ContinueTrue()
			}
		},
		func(r rpf.ProcessorIF, c *gin.Context) {
			// Close Session
			group := GroupCloseSessionWithError(parent, 5998 /* TODO: ERROR [Invalid User Session] */)
			group.Run()
		},
		nil)

	return rif
}
