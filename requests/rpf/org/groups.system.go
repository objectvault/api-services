// cSpell:ignore ginrpf, gonic, paulo ferreira
package org

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
	"github.com/objectvault/api-services/requests/rpf/session"

	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

func GroupSystemOrgRequestInitialize(parent rpf.GINProcessor, roles []uint32) *rpf.ProcessorGroup {
	// Create Processing Group
	group := &rpf.ProcessorGroup{}
	group.Parent = parent

	group.Chain = rpf.ProcessChain{
		// Extract : GIN Parameter 'org' //
		ExtractGINParameterOrg,
		func(r rpf.GINProcessor, c *gin.Context) {
			r.Set("org-id", uint64(0))
		},
		// Get Session User
		func(r rpf.GINProcessor, c *gin.Context) {
			// Is User Session?
			gSessionUser := session.GroupGetSessionUser(r, true, false)
			gSessionUser.Run()

			// User Session Requirements Passed?
			if !r.IsFinished() { // YES: Save User Information
				user_id := gSessionUser.MustGet("user-id").(uint64)
				r.Set("user-id", user_id)
			}
		},
		// REQUEST: Get Organization //
		DBRegistryOrgFindByID,
		// Validate Session Users Permission
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Session User
			org := r.MustGet("registry-org").(*orm.OrgRegistry)

			// Get Session User
			user_id := r.MustGet("user-id").(uint64)

			// Check User has Permissions in Organization
			GroupAssertUserOrganizationPermissions(r, user_id, org.ID(), roles, true, true, false).
				Run()
		},
		func(r rpf.GINProcessor, c *gin.Context) {
			r.LocalToGlobal("registry-org")
		},
	}

	return group
}
