// cSpell:ignore goginrpf, gonic, paulo ferreira
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
	"github.com/objectvault/api-services/requests/rpf/shared"
	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

// ORGANIZATION //

func ExportSystemRegistryOrgList(r rpf.GINProcessor, c *gin.Context) {
	// Get Orgs Registry Entries
	orgs := r.Get("orgs").(orm.TQueryResults)

	list := &shared.ExportList{
		List: orgs,
		ValueMapper: func(v interface{}) interface{} {
			return &FullRegOrgToJSON{
				Registry: v.(*orm.OrgRegistry),
			}
		},
		FieldMapper: func(f string) string {
			switch f {
			case "id_org":
				return "id"
			case "orgname":
				return "alias"
			case "name":
				return "name"
			case "state":
				return "state"
			default:
				return ""
			}
		},
	}

	r.SetReponseDataValue("orgs", list)
}

func ExportOrganizationFull(r rpf.GINProcessor, c *gin.Context) {
	// Get Organization Information
	registry := r.MustGet("registry-org").(*orm.OrgRegistry)
	org := r.MustGet("org").(*orm.Organization)

	// Transform for Export
	d := &FullOrgToJSON{
		Registry:     registry,
		Organization: org,
	}

	r.SetReponseDataValue("organization", d)
}

func ExportOrganizationBasic(r rpf.GINProcessor, c *gin.Context) {
	// Get Organization Information
	registry := r.MustGet("registry-org").(*orm.OrgRegistry)
	org := r.MustGet("org").(*orm.Organization)

	// Transform for Export
	d := &BasicOrgToJSON{
		ID:           registry.ID(),
		Organization: *org,
	}

	r.SetReponseDataValue("organization", d)
}

// STORES //

func ExportRegistryOrgStoreList(r rpf.GINProcessor, c *gin.Context) {
	// Get Org Stores Registry Entries
	stores := r.Get("registry-stores").(orm.TQueryResults)

	list := &shared.ExportList{
		List: stores,
		ValueMapper: func(v interface{}) interface{} {
			return &FullRegOrgStoreToJSON{
				Registry: v.(*orm.OrgStoreRegistry),
			}
		},
		FieldMapper: func(f string) string {
			switch f {
			case "id_store":
				return "store"
			case "storename":
				return "alias"
			case "state":
				return "state"
			default:
				return ""
			}
		},
	}

	r.SetReponseDataValue("stores", list)
}
