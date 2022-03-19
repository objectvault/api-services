// cSpell:ignore goginrpf, gonic, paulo ferreira
package entry

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

func AssertFolderObject(r rpf.GINProcessor, c *gin.Context) {
	// Get Object ID
	o := r.MustGet("store-object").(*orm.StoreObject)

	// Is Folder Object?
	if o.Type() != orm.OBJECT_TYPE_FOLDER { // NO
		r.Abort(4998 /* TODO: Error [Not Folder Object] */, nil)
		return
	}
}

func AssertNotFolderObject(r rpf.GINProcessor, c *gin.Context) {
	// Get Object ID
	o := r.MustGet("store-object").(*orm.StoreObject)

	// Is Folder Object?
	if o.Type() == orm.OBJECT_TYPE_FOLDER { // YES
		r.Abort(4998 /* TODO: Error [Folder Object] */, nil)
		return
	}
}

func AssertNotRootFolder(r rpf.GINProcessor, c *gin.Context) {
	// Get Object ID
	oid := r.MustGet("request-entry-id").(uint32)

	// Is Root Object ID
	if oid == 0 { // YES: Can't Access Root Object Information
		r.Abort(4001, nil)
		return
	}
}
