// cSpell:ignore goginrpf, gonic, paulo ferreira
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
	"github.com/objectvault/api-services/orm"

	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

// STORE //

func ExportStoreFull(r rpf.GINProcessor, c *gin.Context) {
	// Get Organization Information
	registry := r.MustGet("registry-store").(*orm.OrgStoreRegistry)
	store := r.MustGet("store").(*orm.Store)

	// Transform for Export
	d := &FullStoreToJSON{
		Registry: registry,
		Store:    store,
	}

	r.SetResponseDataValue("store", d)
}

func ExportStoreBasic(r rpf.GINProcessor, c *gin.Context) {
	// Get Organization Information
	store_id := r.MustGet("store-id").(uint64)
	store := r.MustGet("store").(*orm.Store)

	// Transform for Export
	d := &BasicStoreToJSON{
		ID:    store_id,
		Store: store,
	}

	r.SetResponseDataValue("store", d)
}

// REGISTRY: STORE <--> USER //
/*
func ExportFullStoreUser(r rpf.GINProcessor, c *gin.Context) {
	r.SetLocal("registry-object-user", r.MustGet("registry-store-user"))
	object.ExportRegistryObjUserFull(r, c)
}

func ExportStoreUser(r rpf.GINProcessor, c *gin.Context) {
	r.SetLocal("registry-object-user", r.MustGet("registry-store-user"))
	object.ExportRegistryObjUserBasic(r, c)
}
*/
