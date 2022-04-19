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
	sharedorg "github.com/objectvault/api-services/requests/rpf/org"
	"github.com/objectvault/api-services/requests/rpf/session"
	"github.com/objectvault/api-services/requests/rpf/shared"
	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

func GetOrgLockState(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.ORG.LOCK", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Extract Route Parameters
		sharedorg.ExtractGINParameterOrg,
		// Validate Basic Request Settings
		sharedorg.AssertNotSystemOrgRequest, // CAN'T MODIFY SYSTEM ORGS STATE
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Request Organization ID
			oid := r.MustGet("request-org").(uint64)

			// Required Roles : Organization Configuration Read
			roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_ORG, orm.FUNCTION_READONLY)}

			// Initialize Request
			org.GroupOrgRequestInitialize(r, oid, roles, false).
				Run()
		},
		// CALCULATE RESPONSE //
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-org").(*orm.OrgRegistry)
			r.SetResponseDataValue("locked", registry.HasAnyStates(orm.STATE_READONLY))
		},
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

func PutOrgLockState(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.ORG.LOCK", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Extract Route Parameters
		sharedorg.ExtractGINParameterOrg,
		// Validate Basic Request Settings
		sharedorg.AssertNotSystemOrgRequest, // CAN'T MODIFY SYSTEM ORGS STATE
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Request Organization ID
			oid := r.MustGet("request-org").(uint64)

			// Required Roles : Organization Configuration Modify
			roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_ORG, orm.FUNCTION_MODIFY)}

			// Initialize Request (Block Changes to System Organization)
			org.GroupOrgRequestInitialize(r, oid, roles, true).
				Run()
		},
		// Extract : GIN Parameter 'bool' //
		shared.ExtractGINParameterBooleanValue,
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
			// Return Current States
			r.SetResponseDataValue("state", registry.State())
		},
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

func GetOrgBlockState(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.ORG.BLOCK", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Extract Route Parameters
		sharedorg.ExtractGINParameterOrg,
		// Validate Basic Request Settings
		sharedorg.AssertNotSystemOrgRequest, // CAN'T MODIFY SYSTEM ORGS STATE
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Request Organization ID
			oid := r.MustGet("request-org").(uint64)

			// Required Roles : Organization Configuration Read
			roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_ORG, orm.FUNCTION_READONLY)}

			// Initialize Request
			org.GroupOrgRequestInitialize(r, oid, roles, false).
				Run()
		},
		// CALCULATE RESPONSE //
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-org").(*orm.OrgRegistry)
			r.SetResponseDataValue("blocked", registry.HasAnyStates(orm.STATE_BLOCKED))
		},
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

func PutOrgBlockState(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.ORG.BLOCK", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Extract Route Parameters
		sharedorg.ExtractGINParameterOrg,
		// Validate Basic Request Settings
		sharedorg.AssertNotSystemOrgRequest, // CAN'T MODIFY SYSTEM ORGS STATE
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Request Organization ID
			oid := r.MustGet("request-org").(uint64)

			// Required Roles : Organization Configuration Modify
			roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_ORG, orm.FUNCTION_MODIFY)}

			// Initialize Request (Block Changes to System Organization)
			org.GroupOrgRequestInitialize(r, oid, roles, true).
				Run()
		},
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
			// Return Current States
			r.SetResponseDataValue("state", registry.State())
		},
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

func GetOrgState(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.ORG.STATE", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Extract Route Parameters
		sharedorg.ExtractGINParameterOrg,
		// Validate Basic Request Settings
		sharedorg.AssertNotSystemOrgRequest, // CAN'T MODIFY SYSTEM ORGS STATE
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Request Organization ID
			oid := r.MustGet("request-org").(uint64)

			// Required Roles : Organization Configuration Read
			roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_ORG, orm.FUNCTION_READONLY)}

			// Initialize Request
			org.GroupOrgRequestInitialize(r, oid, roles, false).
				Run()
		},
		// CALCULATE RESPONSE //
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-org").(*orm.OrgRegistry)
			r.SetResponseDataValue("state", registry.State())
		},
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

func PutOrgState(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.ORG.STATE", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Extract Route Parameters
		sharedorg.ExtractGINParameterOrg,
		// Validate Basic Request Settings
		sharedorg.AssertNotSystemOrgRequest, // CAN'T MODIFY SYSTEM ORGS STATE
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Request Organization ID
			oid := r.MustGet("request-org").(uint64)

			// Required Roles : Organization Configuration Modify
			roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_ORG, orm.FUNCTION_MODIFY)}

			// Initialize Request (Block Changes to System Organization)
			org.GroupOrgRequestInitialize(r, oid, roles, true).
				Run()
		},
		// Extract : GIN Parameter 'uint' //
		shared.ExtractGINParameterUINTValue,
		// UPDATE Registry Entry
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-org").(*orm.OrgRegistry)
			states := r.MustGet("request-value").(uint64)

			// Can Only Update Function States (not System States)
			states = states & orm.STATE_MASK_FUNCTIONS

			registry.SetStates(uint16(states))
		},
		org.DBRegistryOrgUpdate,
		// CALCULATE RESPONSE //
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-org").(*orm.OrgRegistry)
			r.SetResponseDataValue("state", registry.State())
		},
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

func DeleteOrgState(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("DELETE.ORG.STATE", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Extract Route Parameters
		sharedorg.ExtractGINParameterOrg,
		// Validate Basic Request Settings
		sharedorg.AssertNotSystemOrgRequest, // CAN'T MODIFY SYSTEM ORGS STATE
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Request Organization ID
			oid := r.MustGet("request-org").(uint64)

			// Required Roles : Organization Configuration Modify
			roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_ORG, orm.FUNCTION_MODIFY)}

			// Initialize Request (Block Changes to System Organization)
			org.GroupOrgRequestInitialize(r, oid, roles, true).
				Run()
		},
		// Extract : GIN Parameter 'uint' //
		shared.ExtractGINParameterUINTValue,
		// UPDATE Registry Entry
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-org").(*orm.OrgRegistry)
			states := r.MustGet("request-value").(uint64)

			// Can Only Update Function States (not System States)
			states = states & orm.STATE_MASK_FUNCTIONS

			registry.ClearStates(uint16(states))
		},
		org.DBRegistryOrgUpdate,
		// CALCULATE RESPONSE //
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-org").(*orm.OrgRegistry)
			r.SetResponseDataValue("state", registry.State())
		},
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}
