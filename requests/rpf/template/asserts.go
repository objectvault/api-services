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

	"github.com/objectvault/api-services/orm"
)

func AssertTemplateInObject(r rpf.GINProcessor, c *gin.Context) {
	// Get Object Identifier
	obj := r.MustGet("object-id").(uint64)

	// Get Template
	template := r.MustGet("request-template").(string)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Object Shard
	db, e := dbm.Connect(obj)
	if e != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// List Registered Org Users
	exists, e := orm.ExistsRegisteredObjectTemplate(db, obj, template)

	// Failed Retrieving User?
	if e != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Is template associated with the Object?
	if !exists { // NO
		r.Abort(4400, nil)
		return
	}
}
