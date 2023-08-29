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

	"github.com/objectvault/api-services/orm"
	"github.com/objectvault/api-services/requests/rpf/org"
	"github.com/objectvault/api-services/requests/rpf/session"
	"github.com/objectvault/api-services/requests/rpf/shared"
)

func GetOrgLockState(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.SYSTEM.ORG.LOCK", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{}

	// Required Roles : System Organization Role with Read Function
	roles := []uint32{orm.Role(orm.CATEGORY_SYSTEM|orm.SUBCATEGORY_ORG, orm.FUNCTION_READ)}

	// Validate Permissions
	org.AddinGroupValidateOrgRequest(request, func(o string) interface{} {
		switch o {
		case "system-organization":
			return true
		case "roles":
			return roles
		}

		return nil
	})

	// Resolve Request
	request.Append(
		// Extract : GIN Parameter 'org' //
		org.ExtractGINParameterOrgID,
		// SYSTEM Organization Can't be Locked/Blocked
		org.AssertNotSystemOrgRequest,
		// REQUEST: Get Organization (by Way of Registry) //
		func(r rpf.GINProcessor, c *gin.Context) {
			r.Unset("registry-org")
			r.SetLocal("org-id", r.MustGet("request-org"))
		},
		org.DBRegistryOrgFindByID,
		// CALCULATE RESPONSE //
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-org").(*orm.OrgRegistry)
			r.SetResponseDataValue("locked", registry.HasAnyStates(orm.STATE_READONLY))
		},
	)

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}

func PutOrgLockState(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.SYSTEM.ORG.LOCK", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{}

	// Required Roles : System Organization Role with Update Function
	roles := []uint32{orm.Role(orm.CATEGORY_SYSTEM|orm.SUBCATEGORY_ORG, orm.FUNCTION_UPDATE)}

	// Do Basic ORG Request Validation
	org.AddinGroupValidateOrgRequest(request, func(o string) interface{} {
		switch o {
		case "system-organization":
			return true
		case "roles":
			return roles
		}

		return nil
	})

	// Process Request
	request.Append(
		// Extract : GIN Parameter 'org' //
		org.ExtractGINParameterOrgID,
		// SYSTEM Organization Can't be Locked/Blocked
		org.AssertNotSystemOrgRequest,
		// Extract : GIN Parameter 'bool' //
		shared.ExtractGINParameterBooleanValue,
		// REQUEST: Get Organization (by Way of Registry) //
		func(r rpf.GINProcessor, c *gin.Context) {
			r.Unset("registry-org")
			r.SetLocal("org-id", r.MustGet("request-org"))
		},
		org.DBRegistryOrgFindByID,
		// UPDATE Registry Entry
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-org").(*orm.OrgRegistry)
			lock := r.MustGet("request-value").(bool)

			if lock {
				registry.SetStates(orm.STATE_READONLY)
			} else {
				registry.ClearStates(orm.STATE_READONLY)
			}
		},
		org.DBRegistryOrgUpdate,
		// CALCULATE RESPONSE //
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-org").(*orm.OrgRegistry)
			r.SetResponseDataValue("locked", registry.HasAnyStates(orm.STATE_READONLY))
		},
	)

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}

func GetOrgBlockState(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.SYSTEM.ORG.BLOCK", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{}

	// Required Roles : System Organization Role with Read Function
	roles := []uint32{orm.Role(orm.CATEGORY_SYSTEM|orm.SUBCATEGORY_ORG, orm.FUNCTION_READ)}

	// Do Basic ORG Request Validation
	org.AddinGroupValidateOrgRequest(request, func(o string) interface{} {
		switch o {
		case "system-organization":
			return true
		case "roles":
			return roles
		}

		return nil
	})

	// Validate User
	request.Append(
		// Extract : GIN Parameter 'org' //
		org.ExtractGINParameterOrgID,
		// SYSTEM Organization Can't be Locked/Blocked
		org.AssertNotSystemOrgRequest,
		// REQUEST: Get Organization (by Way of Registry) //
		func(r rpf.GINProcessor, c *gin.Context) {
			r.SetLocal("org-id", r.MustGet("request-org"))
		},
		org.DBRegistryOrgFindByID,
		// CALCULATE RESPONSE //
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-org").(*orm.OrgRegistry)
			r.SetResponseDataValue("blocked", registry.HasAnyStates(orm.STATE_BLOCKED))
		},
	)

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}

func PutOrgBlockState(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.SYSTEM.ORG.BLOCK", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{}

	// Required Roles : System Organization Role with Update Function
	roles := []uint32{orm.Role(orm.CATEGORY_SYSTEM|orm.SUBCATEGORY_ORG, orm.FUNCTION_UPDATE)}

	// Do Basic ORG Request Validation
	org.AddinGroupValidateOrgRequest(request, func(o string) interface{} {
		switch o {
		case "system-organization":
			return true
		case "roles":
			return roles
		}

		return nil
	})

	// Validate User
	request.Append(
		// Extract : GIN Parameter 'org' //
		org.ExtractGINParameterOrgID,
		// SYSTEM Organization Can't be Locked/Blocked
		org.AssertNotSystemOrgRequest,
		// Extract : GIN Parameter 'bool' //
		shared.ExtractGINParameterBooleanValue,
		// REQUEST: Get Organization (by Way of Registry) //
		func(r rpf.GINProcessor, c *gin.Context) {
			r.Unset("registry-org")
			r.SetLocal("org-id", r.MustGet("request-org"))
		},
		org.DBRegistryOrgFindByID,
		// UPDATE Registry Entry
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-org").(*orm.OrgRegistry)
			lock := r.MustGet("request-value").(bool)

			if lock {
				registry.SetStates(orm.STATE_BLOCKED)
			} else {
				registry.ClearStates(orm.STATE_BLOCKED)
			}
		},
		org.DBRegistryOrgUpdate,
		// CALCULATE RESPONSE //
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-org").(*orm.OrgRegistry)
			r.SetResponseDataValue("blocked", registry.HasAnyStates(orm.STATE_BLOCKED))
		},
	)

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}
