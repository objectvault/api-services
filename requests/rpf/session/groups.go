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
	"github.com/objectvault/api-services/requests/rpf/user"

	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

func GroupGetSessionUser(parent rpf.GINProcessor, checkUserLock bool, noAdmin bool) *rpf.ProcessorGroup {
	// Create Processing Group
	group := &rpf.ProcessorGroup{}
	group.Parent = parent

	if noAdmin {
		group.Chain = rpf.ProcessChain{
			AssertUserSession,     // Is Active User Session
			AssertNotSystemAdmin,  // Not System Admin Session
			SessionUserToRegistry, // Get User Registry Entry
		}
	} else {
		group.Chain = rpf.ProcessChain{
			AssertUserSession,     // Is Active User Session
			SessionUserToRegistry, // Get User Registry Entry
		}
	}

	// Check if User Locked?
	if !checkUserLock { // YES
		// TODO: Never Check if Admin User is Locked
		group.Chain = append(group.Chain,
			user.AssertUserUnblocked, // Assert User is GLOBALLY Active
		)
	}

	return group
}

func GroupCloseSessionWithError(parent rpf.GINProcessor, ecode int) *rpf.ProcessorGroup {
	// Create Processing Group
	group := &rpf.ProcessorGroup{}
	group.Parent = parent

	group.Chain = rpf.ProcessChain{
		CloseUserSession, // Clear Session Information
		SaveSession,      // Update Session Cookie
		func(r rpf.GINProcessor, c *gin.Context) {
			r.Abort(ecode, nil)
			return
		},
	}

	return group
}
