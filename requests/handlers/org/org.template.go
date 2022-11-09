// cSpell:ignore addin, ginrpf, gonic, paulo, ferreira
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
	"github.com/gin-gonic/gin"

	rpf "github.com/objectvault/goginrpf"

	"github.com/objectvault/api-services/common"
	"github.com/objectvault/api-services/orm"
	"github.com/objectvault/api-services/requests/rpf/org"
	"github.com/objectvault/api-services/requests/rpf/session"
	"github.com/objectvault/api-services/requests/rpf/shared"
	"github.com/objectvault/api-services/requests/rpf/template"
)

// Service Handlers //

// List Active System Templates
func ListSystemTemplates(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.ORG.TEMPLATES.SYSTEM", c, 1000, shared.JSONResponse)

	// Required Roles : Organization Object Template Access with List Function
	roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_TEMPLATE, orm.FUNCTION_LIST)}

	// Basic Request Validate
	org.AddinGroupValidateOrgRequest(request, func(o string) interface{} {
		if o == "roles" { // Roles to Verify
			return roles
		}

		return nil
	})

	// ORGANIZATION for Request is System Organization //
	request.Append(
		func(r rpf.GINProcessor, c *gin.Context) {
			// List comes from System Organization
			r.SetLocal("system-org", common.SYSTEM_ORGANIZATION)
		})

	// Addin Get Template List for Object and Export
	template.AddinGetObjectTemplateList(request, func(o string) interface{} {
		switch o {
		case "source-object": // Source ID Parameter
			return "system-org"
		case "export": // Export List
			return true
		default: // Use Default
			return nil
		}
	})

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}

// List Active Organization Templates
func ListTemplates(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.ORG.TEMPLATES", c, 1000, shared.JSONResponse)

	// Required Roles : Organization Object Template Access with List Function
	roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_TEMPLATE, orm.FUNCTION_LIST)}

	// Basic Request Validate
	org.AddinGroupValidateOrgRequest(request, func(o string) interface{} {
		if o == "roles" { // Roles to Verify
			return roles
		}

		return nil
	})

	// Addin Get Template List for Object and Export
	template.AddinGetObjectTemplateList(request, func(o string) interface{} {
		switch o {
		case "source-object": // Source ID Parameter
			return "request-org"
		case "export": // Export List
			return true
		default: // Use Default
			return nil
		}
	})

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}

// Retrieve Organization Template Definition
func GetTemplate(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.ORG.TEMPLATE", c, 1000, shared.JSONResponse)

	// Required Roles : SYSTEM Organization Object Template Access with Read Function
	roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_TEMPLATE, orm.FUNCTION_READ)}

	// Basic Request Validate
	org.AddinGroupValidateOrgRequest(request, func(o string) interface{} {
		if o == "roles" { // Roles to Verify
			return roles
		}

		return nil
	})

	/*
		// Basic Request Validate
		org.AddinGroupValidateOrgRequest(request, func(o string) interface{} {
			if o == "check-user-roles" { // NO Roles Check Required
				return false
			}

			return nil
		})

		// ROLES Verification //
		request.Append(
			// VERIFY Users Organization Roles //
			func(r rpf.GINProcessor, c *gin.Context) {
				// Set Object to To Search for Roles (SYSTEM ORGANIZATION)
				r.SetLocal("object-id", common.SYSTEM_ORGANIZATION)
				r.SetLocal("roles-required", roles)
			},
			object.AssertUserHasAllRolesInObject, // ASSERT User has required Access Roles
		)
	*/

	// Addin Get Template and Export
	template.AddinGetTemplate(request, func(o string) interface{} {
		switch o {
		case "source-object": // Source ID Parameter
			return "request-org"
		case "export": // Export
			return true
		default: // Use Default
			return nil
		}
	})

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}

// Add Template to Organization
func AddTemplateToOrg(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("POST.STORE.TEMPLATE", c, 1000, shared.JSONResponse)

	// Required Roles : Organization Object Template Access with Create Function
	roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_TEMPLATE, orm.FUNCTION_CREATE)}

	// Basic Request Validate
	org.AddinGroupValidateOrgRequest(request, func(o string) interface{} {
		if o == "roles" { // Roles to Verify
			return roles
		}

		return nil
	})

	// Request Process //
	request.Append(
		// Get Template from Organization //
		template.ExtractGINParameterTemplate,
		template.DBGetTemplate,
		// Register Template with Store //
		func(r rpf.GINProcessor, c *gin.Context) {
			// Set Object
			r.SetLocal("object-id", r.MustGet("org-id"))
			// Get Template
			t := r.MustGet("template").(*orm.Template)
			r.SetLocal("template-title", t.Title())
		},
		template.DBRegisterTemplateWithObject,
		// Request Response //
		template.ExportRegistryTemplate,
	)

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}

// Delete Template from Organization
func DeleteTemplateFromOrg(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("DELETE.ORG.TEMPLATE", c, 1000, shared.JSONResponse)

	// Required Roles : Organization Object Template Access with Read Function
	roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_TEMPLATE, orm.FUNCTION_DELETE)}

	// Basic Request Validate
	org.AddinGroupValidateOrgRequest(request, func(o string) interface{} {
		if o == "roles" { // Roles to Verify
			return roles
		}

		return nil
	})

	// Request Process //
	request.Append(
		// Verify Template Exists in Organization //
		template.ExtractGINParameterTemplate,
		func(r rpf.GINProcessor, c *gin.Context) {
			// Set Object to To Search
			r.SetLocal("object-id", r.MustGet("org-id"))
			r.SetLocal("template-name", r.MustGet("request-template"))
		},
		template.AssertTemplateInObject,
		// Delete Template //
		template.DBDeleteTemplateFromObject,
		// Request Response //
		func(r rpf.GINProcessor, c *gin.Context) {
			// TODO What Value to Return?
			r.SetResponseDataValue("ok", true)
		},
	)

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}
