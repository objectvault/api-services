// cSpell:ignore ginrpf, gonic, paulo ferreira
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
	"github.com/gin-gonic/gin"

	rpf "github.com/objectvault/goginrpf"

	"github.com/objectvault/api-services/orm"
	"github.com/objectvault/api-services/requests/rpf/session"
	"github.com/objectvault/api-services/requests/rpf/shared"
	"github.com/objectvault/api-services/requests/rpf/store"
	"github.com/objectvault/api-services/requests/rpf/template"
)

// List Active Store Templates
func ListStoreTemplates(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.STORE.TEMPLATES", c, 1000, shared.JSONResponse)

	// Required Roles : Store Object Access with List Function
	roles := []uint32{orm.Role(orm.CATEGORY_STORE|orm.SUBCATEGORY_OBJECT, orm.FUNCTION_LIST)}

	// Basic Request Validate
	store.AddinGroupValidateStoreRequest(request, func(o string) interface{} {
		if o == "roles" {
			return roles
		}

		return nil
	})

	// Addin Get Template List for Object and Export
	template.AddinGetObjectTemplateList(request, func(o string) interface{} {
		switch o {
		case "source-object": // Source ID Parameter
			return "request-store"
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

// Retrieve Store Template Definition
func GetTemplate(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.STORE.TEMPLATE", c, 1000, shared.JSONResponse)

	// Required Roles : Store Object Access with Read Function
	roles := []uint32{orm.Role(orm.CATEGORY_STORE|orm.SUBCATEGORY_OBJECT, orm.FUNCTION_READ)}

	// Basic Request Validate
	store.AddinGroupValidateStoreRequest(request, func(o string) interface{} {
		if o == "roles" {
			return roles
		}

		return nil
	})

	// Addin Get Template and Export
	template.AddinGetTemplate(request, func(o string) interface{} {
		switch o {
		case "source-object": // Source ID Parameter
			return "request-store"
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

// Add Template to Store
func AddStoreTemplate(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("POST.STORE.TEMPLATE", c, 1000, shared.JSONResponse)

	// Required Roles : Store Object Access with Read Function
	roles := []uint32{orm.Role(orm.CATEGORY_STORE|orm.SUBCATEGORY_TEMPLATE, orm.FUNCTION_CREATE)}

	// Basic Request Validate
	store.AddinGroupValidateStoreRequest(request, func(o string) interface{} {
		if o == "roles" {
			return roles
		}

		return nil
	})

	// Request Process //
	request.Append(
		template.ExtractGINParameterTemplate,
		func(r rpf.GINProcessor, c *gin.Context) {
			// Set Object to To Search
			r.SetLocal("object-id", r.MustGet("org-id"))
		},
		template.DBGetObjectTemplateRegistry,
		func(r rpf.GINProcessor, c *gin.Context) {
			// Set Object
			r.SetLocal("object-id", r.MustGet("store-id"))
			// Get Organization Entry
			registry := r.MustGet("registry-object-template").(*orm.ObjectTemplateRegistry)
			r.SetLocal("template-title", registry.Title())
		},
		template.DBRegisterTemplateWithObject,
		template.ExportRegistryTemplate,
		func(r rpf.GINProcessor, c *gin.Context) {
			r.Abort(5999, nil)
		},
		session.SaveSession, // Update Session Cookie
	)
}

// Delete Template from Store
func DeleteStoreTemplate(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("DELETE.STORE.TEMPLATE", c, 1000, shared.JSONResponse)

	// Required Roles : Store Object Access with Read Function
	roles := []uint32{orm.Role(orm.CATEGORY_STORE|orm.SUBCATEGORY_TEMPLATE, orm.FUNCTION_DELETE)}

	// Basic Request Validate
	store.AddinGroupValidateStoreRequest(request, func(o string) interface{} {
		if o == "roles" {
			return roles
		}

		return nil
	})

	// Request Process //
	request.Append(
		template.ExtractGINParameterTemplate,
		func(r rpf.GINProcessor, c *gin.Context) {
			// Set Object to To Search
			r.SetLocal("object-id", r.MustGet("store-id"))
		},
		template.AssertTemplateInObject,
		template.DBDeleteTemplateFromObject,
		// Request Response //
		func(r rpf.GINProcessor, c *gin.Context) {
			// TODO What Value to Return?
			r.SetResponseDataValue("ok", true)
		},
		session.SaveSession, // Update Session Cookie
	)
}
