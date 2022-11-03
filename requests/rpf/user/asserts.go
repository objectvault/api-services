// cSpell:ignore goginrpf, gonic, paulo ferreira
package user

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

	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

func AssertNotSystemUserRegistry(r rpf.GINProcessor, c *gin.Context) {
	registry := r.MustGet("registry-user").(*orm.UserRegistry)

	// Is System User?
	if registry.ID() == common.SYSTEM_ADMINISTRATOR { // YES: Abort
		r.Abort(4101, nil)
		return
	}
}

func AssertUserUnblocked(r rpf.GINProcessor, c *gin.Context) {
	// Get Request User
	user := r.MustGet("registry-user").(*orm.UserRegistry)

	// Is the User Blocked?
	if user.IsBlocked() { // YES: Can't Login
		r.Abort(4001, nil)
		return
	}
}

func AssertUserActive(r rpf.GINProcessor, c *gin.Context) {
	// Get Request User
	user := r.MustGet("registry-user").(*orm.UserRegistry)

	// Is the User Active?
	if !user.IsActive() { // YES: Can't Login : Password Failures?
		r.Abort(4001, nil)
		return
	}
}

func AssertUserNotReadOnly(r rpf.GINProcessor, c *gin.Context) {
	// Get Request User
	user := r.MustGet("registry-user").(*orm.UserRegistry)

	// Is the User Account in Read Only Mode?
	if user.IsReadOnly() { // YES: Can't Modify Anything
		r.Abort(4002, nil)
		return
	}
}

func AssertCredentials(r rpf.GINProcessor, c *gin.Context) {
	// Get Request User Password Hash
	user := r.MustGet("registry-user").(*orm.UserRegistry)

	// Test Passed?
	pass := false

	// Get Credentials to Test
	password := r.Get("password")
	// Do we have a Plain Text Password?
	if password != nil { // YES: Calculate HASH
		pass = user.TestPassword(password.(string))
	} else { // NO: Get Password Hash
		hash := r.MustGet("hash").(string)
		pass = user.TestHash(hash)
	}

	// Does the Password Match?
	if !pass { // NO
		r.Abort(3001, nil)
		return
	}
}
