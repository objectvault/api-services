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
	"github.com/objectvault/api-services/orm"
	"github.com/objectvault/api-services/requests/rpf/shared"
	rpf "github.com/objectvault/goginrpf"
)

func ExportRegistryTemplateList(r rpf.GINProcessor, c *gin.Context) {
	// Get Registry Entries
	templates := r.Get("registry-object-templates").(orm.TQueryResults)

	list := &shared.ExportList{
		List: templates,
		ValueMapper: func(v interface{}) interface{} {
			return &FullTemplateRegistryToJSON{
				Registry: v.(*orm.ObjectTemplateRegistry),
			}
		},
		FieldMapper: func(f string) string {
			// MAP ORM FIELDS to JSON
			switch f {
			case "id_object":
				return "object"
			case "template":
				return "name"
			case "title":
				return "title"
			default:
				return ""
			}
		},
	}

	r.SetResponseDataValue("templates", list)
}

func ExportRegistryTemplate(r rpf.GINProcessor, c *gin.Context) {
	// Get Registry Entries
	registry := r.MustGet("registry-object-template").(*orm.ObjectTemplateRegistry)

	v := &FullTemplateRegistryToJSON{
		Registry: registry,
	}

	r.SetResponseDataValue("templates", v)
}

func ExportTemplate(r rpf.GINProcessor, c *gin.Context) {
	// Get Organization Information
	t := r.MustGet("template").(*orm.Template)

	// Transform for Export
	v := &FullTemplateToJSON{
		Template: t,
	}

	r.SetResponseDataValue("template", v)
}
