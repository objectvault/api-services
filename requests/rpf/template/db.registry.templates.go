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

	rpf "github.com/objectvault/goginrpf"
)

func DBGetObjectTemplates(r rpf.GINProcessor, c *gin.Context) {
	// Get Object Identifier
	obj := r.MustGet("object-id").(uint64)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Object Shard
	db, err := dbm.Connect(obj)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// List Registered Org Users
	q := r.MustGet("query-conditions").(*orm.QueryConditions)
	templates, err := orm.QueryRegisteredObjectTemplates(db, obj, q, true)

	// Failed Retrieving User?
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Save List
	r.Set("registry-object-templates", templates)
}
