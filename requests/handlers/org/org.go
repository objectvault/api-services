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
	"github.com/objectvault/api-services/requests/rpf/org"
	"github.com/objectvault/api-services/requests/rpf/session"
	"github.com/objectvault/api-services/requests/rpf/shared"

	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

// TODO: Each Organization can Have it's Own Password Policy (Minimimum Length, Character User, Max Valid Days, etc.)

// User Has Access and Roles in Organization

func Get(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.ORG", c, 1000, shared.JSONResponse)

	// Required Roles : Organization Access with Read Function
	roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_ORG, orm.FUNCTION_READ)}

	// Base Validation for Org Request
	org.AddinGroupValidateOrgRequest(request, func(o string) interface{} {
		switch o {
		case "check-user-unlocked":
			return true
		case "check-user-roles":
			return true
		case "roles":
			return roles
		}

		return nil
	})

	// Request Processing Chain
	request.Append(
		// PREPARE Response
		func(r rpf.GINProcessor, c *gin.Context) {
			org := r.MustGet("registry-org").(*orm.OrgRegistry)
			r.SetLocal("org-id", org.ID())
		},
		org.DBGetOrgByID,
		// Export Results //
		org.ExportOrganizationFull,
	)

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}

func Put(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.ORG", c, 1000, shared.JSONResponse)

	// Required Roles : Organization Access with Read Function
	roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_CONF, orm.FUNCTION_MODIFY)}

	// Base Validation for Org Request
	org.AddinGroupValidateOrgRequest(request, func(o string) interface{} {
		switch o {
		case "check-user-unlocked":
			return true
		case "check-user-roles":
			return true
		case "roles":
			return roles
		}

		return nil
	})

	// Request Processing Chain
	request.Append(
		// EXTRACT : JSON Body
		shared.RequestExtractJSON,
		// GET Organization Object
		func(r rpf.GINProcessor, c *gin.Context) {
			org := r.MustGet("registry-org").(*orm.OrgRegistry)
			r.SetLocal("org-id", org.ID())
		},
		org.DBGetOrgByID, // Get Organization Value
		// Create Organization From Post //
		org.ManagerUpdateFromJSON,
		org.DBUpdateOrg,
		org.DBRegistryOrgUpdateFromOrg,
		// TODO: If Registry Update - Have to Update Org Information in (registry_users_orgs) - IDEA Background Job to Update All Shards
		// Request Response //
		org.ExportOrganizationBasic,
		session.SaveSession, // Update Session Cookie
	)

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}
