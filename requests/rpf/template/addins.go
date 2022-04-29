// cSpell:ignore goginrpf, gonic, paulo ferreira
package template

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

	"github.com/objectvault/api-services/requests/rpf/shared"
)

func AddinGetTemplate(g rpf.GINGroupProcessor, opts shared.TAddinCallbackOptions) rpf.GINGroupProcessor {

	g.Append(
		ExtractGINParameterTemplate,
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Source Object ID Variable
			source := shared.HelperAddinOptionsCallback(opts, "source-object", "").(string)
			r.SetLocal("object-id", r.MustGet(source))
			r.SetLocal("template-name", r.MustGet("request-template"))
		},
		AssertTemplateInObject,
		DBGetTemplate,
	)

	// OPTION: Check if Exporting List Directly? (DEFAULT: Check)
	if shared.HelperAddinOptionsCallback(opts, "export", false).(bool) {
		g.Append(ExportTemplate) // Export Template
	}

	return g
}

func AddinGetObjectTemplateList(g rpf.GINGroupProcessor, opts shared.TAddinCallbackOptions) rpf.GINGroupProcessor {

	g.Append(
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Source Object ID Variable
			source := shared.HelperAddinOptionsCallback(opts, "source-object", "").(string)
			r.SetLocal("object-id", r.MustGet(source))
		},
		// Extract Query Parameters //
		func(r rpf.GINProcessor, c *gin.Context) {
			gQuery := shared.GroupExtractQueryConditions(r, nil, func(f string) string {
				// MAP JSON to ORM FIELDS
				if f == "template" {
					return "template"
				}
				// ELSE: Invalid Field
				return ""
			}, nil)

			gQuery.Run()
			if !r.IsFinished() { // YES
				// Save Query Settings as Global
				gQuery.LocalToGlobal("query-conditions")
			}
		},
		// Query Organization
		DBGetObjectTemplates,
	)

	// OPTION: Check if Exporting List Directly? (DEFAULT: Check)
	if shared.HelperAddinOptionsCallback(opts, "export", false).(bool) {
		g.Append(ExportRegistryTemplateList) // Make sure user is Unblocked
	}

	return g
}
