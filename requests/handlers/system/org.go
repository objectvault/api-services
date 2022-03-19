// cSpell:ignore goginrpf, gonic, orgs, paulo, ferreira
package system

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

	"github.com/objectvault/api-services/common"
	"github.com/objectvault/api-services/orm"
	"github.com/objectvault/api-services/requests/rpf/object"
	"github.com/objectvault/api-services/requests/rpf/org"
	"github.com/objectvault/api-services/requests/rpf/session"
	"github.com/objectvault/api-services/requests/rpf/shared"
)

func GetOrgProfile(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.SYSTEM.ORG", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Extract : GIN Parameter 'org' //
		org.ExtractGINParameterOrg,
		// Validate Session Users Permission
		func(r rpf.GINProcessor, c *gin.Context) {
			id := r.MustGet("request-org").(string)
			r.Set("org", id)
		},
		// Validate Session Users Permission
		func(r rpf.GINProcessor, c *gin.Context) {
			// Is User Session?
			gSessionUser := session.GroupGetSessionUser(r, true, false)
			gSessionUser.Run()
			if !r.IsFinished() { // YES
				// Get Session User
				user_id := gSessionUser.MustGet("user-id").(uint64)

				// Required Roles : Organization 0 Access with Role Orgs Update
				roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_ORG, orm.FUNCTION_READ)}

				// Check User has Permissions in System Organization
				org.GroupAssertUserOrganizationPermissions(r, user_id, common.SYSTEM_ORGANIZATION, roles, true, true, false).
					Run()

				// Session Requirements Passed?
				if !r.IsFinished() { // YES: Save User Information
					r.SetLocal("user-id", user_id)
				}
			}
		},
		// REQUEST: Get Organization (by Way of Registry) //
		org.DBRegistryOrgFind,
		func(r rpf.GINProcessor, c *gin.Context) {
			org := r.MustGet("registry-org").(*orm.OrgRegistry)
			r.SetLocal("org-id", org.ID())
		},
		org.DBGetOrgByID,
		// Export Results //
		org.ExportOrganizationFull,
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

func PostCreateOrg(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("POST.SYSTEM.ORG", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Validate Session Users Permission
		func(r rpf.GINProcessor, c *gin.Context) {
			// Is User Session?
			gSessionUser := session.GroupGetSessionUser(r, true, false)
			gSessionUser.Run()
			if !r.IsFinished() { // YES
				// Get Session User
				user := gSessionUser.MustGet("registry-user").(*orm.UserRegistry)

				// Required Roles : Organization 0 Access with Role Orgs Update
				roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_ORG, orm.FUNCTION_CREATE)}

				// Check User has Permissions in System Organization
				org.GroupAssertUserOrganizationPermissions(r, user.ID(), common.SYSTEM_ORGANIZATION, roles, true, true, false).
					Run()

				// Session Requirements Passed?
				if !r.IsFinished() { // YES: Save User Information
					gSessionUser.LocalToGlobal("registry-user")
					r.Set("user-id", user.ID())
				}
			}
		},
		// Create Organization From JSON Post //
		shared.RequestExtractJSON,
		org.CreateFromJSON,
		org.DBInsertOrg,
		org.DBRegisterOrg,
		func(r rpf.GINProcessor, c *gin.Context) {
			// Set Default Org Administration Roles
			roles := []uint32{0x201FFFF, 0x202FFFF, 0x203FFFF, 0x204FFFF, 0x205FFFF}
			r.SetLocal("register-roles", roles)
			r.SetLocal("register-org-admin", true)
		},
		object.DBRegisterUserWithOrg,
		object.DBRegisterOrgWithUser,
		// Request Response //
		org.ExportOrganizationFull,
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

func PutOrgProfile(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.SYSTEM.ORG", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Extract : GIN Parameter 'org' //
		org.ExtractGINParameterOrg,
		func(r rpf.GINProcessor, c *gin.Context) {
			id := r.MustGet("request-org").(string)
			r.Set("org", id)
		},
		// EXTRACT : JSON Body
		shared.RequestExtractJSON,
		// Validate Session Users Permission
		func(r rpf.GINProcessor, c *gin.Context) {
			// Is User Session?
			gSessionUser := session.GroupGetSessionUser(r, true, false)
			gSessionUser.Run()
			if !r.IsFinished() { // YES
				// Get Session User
				user_id := gSessionUser.MustGet("user-id").(uint64)

				// Required Roles : Organization 0 Access with Role Orgs Update
				roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_ORG, orm.FUNCTION_UPDATE)}

				// Check User has Permissions in System Organization
				org.GroupAssertUserOrganizationPermissions(r, user_id, common.SYSTEM_ORGANIZATION, roles, true, true, false).
					Run()

				// Session Requirements Passed?
				if !r.IsFinished() { // YES: Save User Information
					r.SetLocal("user-id", user_id)
				}
			}
		},
		// SEARCH Regisrty for Entry
		org.DBRegistryOrgFind,
		// GET Organization Object
		func(r rpf.GINProcessor, c *gin.Context) {
			org := r.MustGet("registry-org").(*orm.OrgRegistry)
			r.SetLocal("org-id", org.ID())
		},
		org.DBGetOrgByID, // Get Organization Value
		// Create Organization From Post //
		org.SystemUpdateFromJSON,
		org.DBUpdateOrg,
		org.DBRegistryOrgUpdateFromOrg,
		// TODO: If Registry Update - Have to Update Org Information in (registry_users_orgs) - IDEA Background Job to Update All Shards
		// Request Response //
		org.ExportOrganizationFull,
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

func DeleteOrg(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("DELETE.SYSTEM.ORG", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Extract : GIN Parameter 'org' //
		org.ExtractGINParameterOrg,
		func(r rpf.GINProcessor, c *gin.Context) {
			id := r.MustGet("request-org").(string)
			r.Set("org", id)
		},
		// Validate Session Users Permission
		func(r rpf.GINProcessor, c *gin.Context) {
			// Is User Session?
			gSessionUser := session.GroupGetSessionUser(r, true, false)
			gSessionUser.Run()
			if !r.IsFinished() { // YES
				// Get Session User
				user_id := gSessionUser.MustGet("user-id").(uint64)

				// Required Roles : Organization 0 Access with Role Orgs Update
				roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_ORG, orm.FUNCTION_DELETE)}

				// Check User has Permissions in System Organization
				org.GroupAssertUserOrganizationPermissions(r, user_id, common.SYSTEM_ORGANIZATION, roles, true, true, false).
					Run()

				// Session Requirements Passed?
				if !r.IsFinished() { // YES: Save User Information
					r.SetLocal("user-id", user_id)
				}
			}
		},
		// SEARCH Regisrty for Entry
		org.DBRegistryOrgFind,
		org.AssertNotSystemOrgRegistry,
		// REQUEST VALIDATION //
		func(r rpf.GINProcessor, c *gin.Context) {
			r.Abort(5999, nil)
		},
	}

	// Start Request Processing
	request.Run()
}
