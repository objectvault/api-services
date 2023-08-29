// cSpell:ignore goginrpf, gonic, paulo ferreira
package store

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
	"github.com/objectvault/api-services/requests/rpf/org"
	"github.com/objectvault/api-services/requests/rpf/session"
	"github.com/objectvault/api-services/requests/rpf/user"

	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

func GroupAssertUserStorePermissions(parent rpf.GINProcessor, userid uint64, storeid uint64, roles []uint32, checkUserLock bool, checkUserIsAdmin bool) *rpf.ProcessorGroup {
	// Create Processing Group
	group := &rpf.ProcessorGroup{}
	group.Parent = parent

	// User to
	group.SetLocal("user-id", userid)

	// Store
	group.SetLocal("request-store", storeid)

	// Required Roles
	group.SetLocal("roles-required", roles)

	group.Chain = rpf.ProcessChain{
		DBStoreUserGet, // Get Store Entry for User
	}

	// Check if User is Store Locked?
	if checkUserLock { // YES
		group.Chain = append(group.Chain,
			object.AssertObjectUserBlocked, // ASSERT User is Active in Organization
		)
	}

	// Check if User is Organization Locked?
	if checkUserIsAdmin { // YES
		group.Chain = append(group.Chain,
			object.AssertObjectUserAdmin, // ASSERT User is Admin
		)
	}

	// Check if User is  Locked?
	if checkUserLock { // YES
		group.Chain = append(group.Chain,
			AssertStoreUserUnblocked, // ASSERT User is Active in Store
		)
	}

	// FINALLY: Verify User Roles in Store
	if roles != nil {
		group.Chain = append(group.Chain,
			AssertUserHasAllRolesInStore, // ASSERT User has required Access Roles
		)
	}

	return group
}

/* NOTE:
 * Can't Move to Org Package as it Creates a Package Cycle
 * ORG PACKAGE requires STORE PACKAGE requires ORG PACKAGE
 */

func GroupOrgStoreRequestInitialize(parent rpf.GINProcessor, roles []uint32, loadRegistry, checkStoreLock bool) *rpf.ProcessorGroup {
	// Create Processing Group
	group := &rpf.ProcessorGroup{}
	group.Parent = parent

	group.Chain = rpf.ProcessChain{
		// Extract : GIN Parameters 'org' and 'store' //
		org.ExtractGINParameterOrg,
		ExtractGINParameterStore,
		func(r rpf.GINProcessor, c *gin.Context) {
			id := r.MustGet("request-org")
			r.Set("org", id)
		},
		// Get Session User
		func(r rpf.GINProcessor, c *gin.Context) {
			// Is User Session?
			gSessionUser := session.GroupGetSessionUser(r, true, true)
			gSessionUser.Run()

			// User Session Requirements Passed?
			if !r.IsFinished() { // YES: Save User Information
				userID := gSessionUser.MustGet("user-id").(uint64)
				r.Set("user-id", userID)
			}
		},
		// REQUEST: Get Organization (by Way of Registry) //
		org.DBRegistryOrgFind,
		// ASSERT: Can't Use this Request for System Organization
		org.AssertNotSystemOrgRegistry,
		// Validate Session Users Permission
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Request Organization
			o := r.MustGet("registry-org").(*orm.OrgRegistry)
			r.Set("org-id", o.ID())
			r.LocalToGlobal("registry-org")

			// Get Session User
			userID := r.MustGet("user-id").(uint64)

			// Check User has Permissions in Organization
			org.GroupAssertUserOrganizationPermissions(r, userID, o.ID(), roles, true, true, false).
				Run()
		},
	}

	// Load Org Store Registry Entry
	if loadRegistry { // YES
		group.Chain = append(group.Chain,
			org.DBOrgStoreFind,
			func(r rpf.GINProcessor, c *gin.Context) {
				r.LocalToGlobal("registry-store")

				// Save Store ID
				s := r.MustGet("registry-store").(*orm.OrgStoreRegistry)
				r.Set("request-store", s.Store())
			},
		)
	}

	// Check if Store is Globally Locked?
	if checkStoreLock { // YES
		group.Chain = append(group.Chain,
			AssertStoreUnblocked, // Assert Store is Active
		)
	}

	return group
}

func GroupStoreRequestInitialize(parent rpf.GINProcessor, storeID uint64, roles []uint32) *rpf.ProcessorGroup {
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
			}
		},
		// Validate Session Users Permission
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Session User
			userID := r.MustGet("user-id").(uint64)

			// Check User has Permissions in Store
			g := GroupAssertUserStorePermissions(r, userID, storeID, roles, true, false)
			g.Run()

			// Not Finished?
			if !r.IsFinished() { // YES
				// Save Store<-->User Registry Entry
				g.LocalToGlobal("registry-store-user")

				// Save Store Information
				r.Set("request-store", storeID)
			}
		},
	}

	return group
}

func GroupStoreUserRequestInitialize(parent rpf.GINProcessor, roles []uint32) *rpf.ProcessorGroup {
	// Create Processing Group
	group := &rpf.ProcessorGroup{}
	group.Parent = parent

	group.Chain = rpf.ProcessChain{
		// Extract : GIN Parameters 'org' and 'store' //
		ExtractGINParameterStore,
		user.ExtractGINParameterUser,
		// Get Session User
		func(r rpf.GINProcessor, c *gin.Context) {
			// Is User Session?
			gSessionUser := session.GroupGetSessionUser(r, true, true)
			gSessionUser.Run()

			// User Session Requirements Passed?
			if !r.IsFinished() { // YES: Save User Information
				userID := gSessionUser.MustGet("user-id").(uint64)
				r.SetLocal("user-id", userID)
			}
		},
		// Validate Session Users Permission
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Request Store
			sid := r.MustGet("request-store").(uint64)

			// Get Session User
			uid := r.MustGet("user-id").(uint64)

			// Check User has Permissions in Store
			GroupAssertUserStorePermissions(r, uid, sid, roles, true, false).
				Run()
		},
		func(r rpf.GINProcessor, c *gin.Context) {
			r.LocalToGlobal("store")
			r.LocalToGlobal("request-user")
		},
	}

	return group
}

func GroupStoreUserAdminRequestInitialize(parent rpf.GINProcessor, roles []uint32) *rpf.ProcessorGroup {
	// Create Processing Group
	group := &rpf.ProcessorGroup{}
	group.Parent = parent

	group.Chain = rpf.ProcessChain{
		// Extract : GIN Parameters 'org' and 'store' //
		ExtractGINParameterStore,
		user.ExtractGINParameterUser,
		// Get Session User
		func(r rpf.GINProcessor, c *gin.Context) {
			// Is User Session?
			gSessionUser := session.GroupGetSessionUser(r, true, true)
			gSessionUser.Run()

			// User Session Requirements Passed?
			if !r.IsFinished() { // YES: Save User Information
				userID := gSessionUser.MustGet("user-id").(uint64)
				r.SetLocal("user-id", userID)
			}
		},
		// Validate Session Users Permission
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Request Store
			sid := r.MustGet("request-store").(uint64)

			// Get Session User
			uid := r.MustGet("user-id").(uint64)

			// Check User has Permissions in Store
			GroupAssertUserStorePermissions(r, uid, sid, roles, true, true).
				Run()
		},
		func(r rpf.GINProcessor, c *gin.Context) {
			r.LocalToGlobal("store")
			r.LocalToGlobal("request-user")
		},
	}

	return group
}
