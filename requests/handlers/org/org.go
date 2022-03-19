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
	"github.com/objectvault/api-services/requests/rpf/shared"

	sharedorg "github.com/objectvault/api-services/requests/rpf/org"
	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

// TODO: Each Organization can Have it's Own Password Policy (Minimimum Length, Character User, Max Valid Days, etc.)

// User Has Access and Roles in Organization

func Get(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.ORG", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Extract Route Parameters
		sharedorg.ExtractGINParameterOrg,
		// Validate Basic Request Settings
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Request Organization ID
			oid := r.MustGet("request-org").(uint64)

			// Required Roles : Organization Access with Read Function
			roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_ORG, orm.FUNCTION_READ)}

			// Initialize Request
			sharedorg.GroupOrgRequestInitialize(r, oid, roles, false).
				Run()
		},
		// PREPARE Response
		func(r rpf.GINProcessor, c *gin.Context) {
			org := r.MustGet("registry-org").(*orm.OrgRegistry)
			r.SetLocal("org-id", org.ID())
		},
		sharedorg.DBGetOrgByID,
		// Export Results //
		sharedorg.ExportOrganizationFull,
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

func Put(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.ORG", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Extract Route Parameters
		sharedorg.ExtractGINParameterOrg,
		// Validate Basic Request Settings
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Request Organization ID
			oid := r.MustGet("request-org").(uint64)

			// Required Roles : Organization Access with Read Function
			roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_CONF, orm.FUNCTION_MODIFY)}

			// Initialize Request
			sharedorg.GroupOrgRequestInitialize(r, oid, roles, false).
				Run()
		},
		// EXTRACT : JSON Body
		shared.RequestExtractJSON,
		// GET Organization Object
		func(r rpf.GINProcessor, c *gin.Context) {
			org := r.MustGet("registry-org").(*orm.OrgRegistry)
			r.SetLocal("org-id", org.ID())
		},
		sharedorg.DBGetOrgByID, // Get Organization Value
		// Create Organization From Post //
		sharedorg.ManagerUpdateFromJSON,
		sharedorg.DBUpdateOrg,
		sharedorg.DBRegistryOrgUpdateFromOrg,
		// TODO: If Registry Update - Have to Update Org Information in (registry_users_orgs) - IDEA Background Job to Update All Shards
		// Request Response //
		sharedorg.ExportOrganizationBasic,
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}
