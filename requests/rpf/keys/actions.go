// cSpell:ignore goginrpf, gonic, paulo ferreira
package keys

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
	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
	"github.com/objectvault/api-services/orm"
)

func KeyExtractBytes(r rpf.GINProcessor, c *gin.Context) {
	// Get User Identifier
	key := r.MustGet("key-object").(*orm.Key)
	pick := r.MustGet("key-pick").([]byte)

	bytes, e := key.DecryptKey(pick)
	if e != nil { // ERROR: Unexpected
		r.Abort(5900, nil)
		return
	}

	r.SetLocal("key-bytes", bytes) // Key Global ID
}
