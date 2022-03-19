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

	"github.com/objectvault/api-services/requests/rpf/session"
	"github.com/objectvault/api-services/requests/rpf/shared"
	"github.com/objectvault/api-services/requests/rpf/store"
	"github.com/objectvault/api-services/requests/rpf/template"
)

// TODO IMPLEMENT: List Active Store Templates
func GetTemplates(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.STORE.TEMPLATES", c, 1000, shared.JSONResponse)

	// Basic Request Validate
	store.AddinGroupValidateStoreRequest(request, func(o string) interface{} {
		if o == "check-user-roles" { // NO Roles Check Required
			return false
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

// TODO IMPLEMENT: READ Store Template Definition
func GetTemplate(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.STORE.TEMPLATE", c, 1000, shared.JSONResponse)

	// Basic Request Validate
	store.AddinGroupValidateStoreRequest(request, func(o string) interface{} {
		if o == "check-user-roles" { // NO Roles Check Required
			return false
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
