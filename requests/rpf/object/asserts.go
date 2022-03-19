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
	"github.com/objectvault/api-services/orm"
	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

// USERS //

func AssertObjectUserUnblocked(r rpf.GINProcessor, c *gin.Context) {
	// Get Request Org's USer Entry
	entry := r.MustGet("registry-object-user").(*orm.ObjectUserRegistry)

	// Is the User Blocked?
	if entry.IsBlocked() { // YES: Can't Access Org Information
		r.Abort(5998, nil) // TODO: Specific Error
		return
	}
}

func AssertObjectUserAdmin(r rpf.GINProcessor, c *gin.Context) {
	// Get Request Org's USer Entry
	entry := r.MustGet("registry-object-user").(*orm.ObjectUserRegistry)

	// Is Admin User in Ob ject?
	if !entry.IsAdminUser() { // NO
		r.Abort(5998, nil) // TODO: Specific Error
		return
	}
}

func AssertObjectUserNotAdmin(r rpf.GINProcessor, c *gin.Context) {
	// Get Request Org's USer Entry
	entry := r.MustGet("registry-object-user").(*orm.ObjectUserRegistry)

	// Is Admin User in Ob ject?
	if entry.IsAdminUser() { // NO
		r.Abort(5998, nil) // TODO: Specific Error
		return
	}
}

func AssertUserHasAllRolesInObject(r rpf.GINProcessor, c *gin.Context) {
	// Get Request Org's User Entry
	entry := r.Get("registry-object-user").(*orm.ObjectUserRegistry)
	required := r.Get("roles-required").([]uint32)

	// Loop Through Required Roles
	pass := true
	for _, r := range required {
		if !entry.HasRole(r) {
			pass = false
			break
		}
	}

	// Does the User have the Required Roles?
	if !pass { // NO: Fail Request (Permission Denied)
		r.Abort(4003, nil)
		return
	}
}

func AssertUserHasOneRoleInObject(r rpf.GINProcessor, c *gin.Context) {
	// Get Request Org's User Entry
	entry := r.Get("registry-object-user").(*orm.ObjectUserRegistry)
	required := r.Get("roles-required").([]uint32)

	// Loop Through Possible Roles
	pass := false
	for _, r := range required {
		if entry.HasRole(r) {
			pass = true
			break
		}
	}

	// Does the User have the Required Roles?
	if !pass { // NO: Fail Request (Permission Denied)
		r.Abort(5998, nil) // TODO: Choose Correct Error - User Doesn't Have Required Roles
		return
	}
}
