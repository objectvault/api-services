// cSpell:ignore orgid, userid
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
	"github.com/objectvault/api-services/requests/rpf/object"
	"github.com/objectvault/api-services/requests/rpf/session"
	"github.com/objectvault/api-services/requests/rpf/user"

	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

func GroupAssertUserOrganizationPermissions(parent rpf.GINProcessor, userid uint64, orgid uint64, roles []uint32, checkOrgLock bool, checkUserLock bool, checkUserIsAdmin bool) *rpf.ProcessorGroup {
	// Create Processing Group
	group := &rpf.ProcessorGroup{}
	group.Parent = parent

	// User to
	group.SetLocal("user-id", userid)

	// Organization
	group.SetLocal("org-id", orgid)

	// Required Roles
	group.SetLocal("roles-required", roles)

	// Check if Organization is Globally Locked?
	if checkOrgLock { // YES
		group.Chain = append(group.Chain,
			DBRegistryOrgFindByID, // Get Organization Registry Entry
			AssertOrgUnblocked,    // Assert ORG is GLOBALLY Active
		)
	}

	// Check if User is Organization Locked?
	if checkUserLock { // YES
		group.Chain = append(group.Chain,
			object.DBOrgUserFind,           // Get Organization Entry for User
			object.AssertObjectUserBlocked, // ASSERT User is Active in Organization
		)
	}

	// Check if User is Admin in Organization?
	if checkUserIsAdmin { // YES
		group.Chain = append(group.Chain,
			object.AssertObjectUserAdmin, // ASSERT If User is not Admin
		)
	}

	// FINALLY: Verify User Roles in Organization
	if roles != nil {
		group.Chain = append(group.Chain,
			object.AssertUserHasAllRolesInObject, // ASSERT User has required Access Roles
		)
	}

	return group
}

func GroupOrgRequestInitialize(parent rpf.GINProcessor, orgRef interface{}, roles []uint32, noSystemOrg bool) *rpf.ProcessorGroup {
	// Create Processing Group
	group := &rpf.ProcessorGroup{}
	group.Parent = parent

	group.Chain = rpf.ProcessChain{
		// Get Session User
		func(r rpf.GINProcessor, c *gin.Context) {
			// Is User Session?
			g := session.GroupGetSessionUser(r, true, true)
			g.Run()

			// User Session Requirements Passed?
			if !r.IsFinished() { // YES
				// Save User Information
				g.LocalToGlobal("user-id")
				g.LocalToGlobal("registry-user")

				// Set Organization Reference for Current Group
				r.SetLocal("org", orgRef)
			}
		},
		// REQUEST: Get Organization (by Way of Registry) //
		DBRegistryOrgFind,
		// ASSERT: Can't Use this Request for System Organization
		func(r rpf.GINProcessor, c *gin.Context) {
			if noSystemOrg {
				AssertNotSystemOrgRegistry(r, c)
			}
		},
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Session User
			user_id := r.MustGet("user-id").(uint64)

			// Get Organization
			org := r.MustGet("registry-org").(*orm.OrgRegistry)

			// Check User has Permissions in Organization
			g := GroupAssertUserOrganizationPermissions(r, user_id, org.ID(), roles, true, true, false)
			g.Run()

			// User Session Requirements Passed?
			if !r.IsFinished() { // YES
				// Save Organization Registry Entry
				r.LocalToGlobal("registry-org")

				// NOTE: request-org can be ShardID(uint64) or Org Alias(string)
				r.Set("request-org", orgRef)
				r.Set("org-id", org.ID())
			}
		},
	}

	return group
}

func GroupOrgUserRequestInitialize(parent rpf.GINProcessor, roles []uint32) *rpf.ProcessorGroup {
	// Create Processing Group
	group := &rpf.ProcessorGroup{}
	group.Parent = parent

	group.Chain = rpf.ProcessChain{
		// Extract : GIN Parameters 'org' and 'user' //
		ExtractGINParameterOrg,
		user.ExtractGINParameterUser,
		func(r rpf.GINProcessor, c *gin.Context) {
			id := r.MustGet("request-org")
			r.Set("org", id)
		},
		// Get Session User
		func(r rpf.GINProcessor, c *gin.Context) {
			// Is User Session?
			gSessionUser := session.GroupGetSessionUser(r, true, false)
			gSessionUser.Run()

			// User Session Requirements Passed?
			if !r.IsFinished() { // YES: Save User Information
				user_id := gSessionUser.MustGet("user-id").(uint64)
				r.SetLocal("user-id", user_id)
			}
		},
		// REQUEST: Get Organization (by Way of Registry) //
		DBRegistryOrgFind,
		// ASSERT: Can't Use this Request for System Organization
		// AssertNotSystemOrgRegistry,
		// Validate Session Users Permission
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Request Organization
			org := r.MustGet("registry-org").(*orm.OrgRegistry)

			// Get Session User
			user_id := r.MustGet("user-id").(uint64)

			// Check User has Permissions in System Organization
			g := GroupAssertUserOrganizationPermissions(r, user_id, org.ID(), roles, true, true, false)
			g.Run()

			if !r.IsFinished() {
				g.LocalToGlobal("registry-org-user")
			}
		},
		func(r rpf.GINProcessor, c *gin.Context) {
			r.LocalToGlobal("registry-org")
			r.LocalToGlobal("request-user")

			// Save Org ID
			org := r.MustGet("registry-org").(*orm.OrgRegistry)
			r.Set("org-id", org.ID())
		},
	}

	return group
}

func GroupOrgUserAdminRequestInitialize(parent rpf.GINProcessor, roles []uint32) *rpf.ProcessorGroup {
	// Create Processing Group
	group := &rpf.ProcessorGroup{}
	group.Parent = parent

	group.Chain = rpf.ProcessChain{
		// Extract : GIN Parameters 'org' and 'user' //
		ExtractGINParameterOrg,
		user.ExtractGINParameterUser,
		func(r rpf.GINProcessor, c *gin.Context) {
			id := r.MustGet("request-org")
			r.Set("org", id)
		},
		// Get Session User
		func(r rpf.GINProcessor, c *gin.Context) {
			// Is User Session?
			gSessionUser := session.GroupGetSessionUser(r, true, false)
			gSessionUser.Run()

			// User Session Requirements Passed?
			if !r.IsFinished() { // YES: Save User Information
				user_id := gSessionUser.MustGet("user-id").(uint64)
				r.SetLocal("user-id", user_id)
			}
		},
		// REQUEST: Get Organization (by Way of Registry) //
		DBRegistryOrgFind,
		// Validate Session Users Permission
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Request Organization
			org := r.MustGet("registry-org").(*orm.OrgRegistry)

			// Get Session User
			user_id := r.MustGet("user-id").(uint64)

			// Check User has Permissions in System Organization
			g := GroupAssertUserOrganizationPermissions(r, user_id, org.ID(), roles, true, true, true)
			g.Run()

			if !r.IsFinished() {
				g.LocalToGlobal("registry-org-user")
			}
		},
		func(r rpf.GINProcessor, c *gin.Context) {
			r.LocalToGlobal("registry-org")
			r.LocalToGlobal("request-user")

			// Save Org ID
			org := r.MustGet("registry-org").(*orm.OrgRegistry)
			r.Set("org-id", org.ID())
		},
	}

	return group
}
