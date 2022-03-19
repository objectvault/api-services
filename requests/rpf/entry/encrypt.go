// cSpell:ignore ginrpf, gonic, paulo ferreira
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

func DecryptStoreObject(r rpf.GINProcessor, c *gin.Context) {
	// Get Store Key and Object to Decrypt
	skey := r.MustGet("store-key").([]byte)
	o := r.MustGet("store-object").(*orm.StoreObject)

	// Decrypt Object
	t := &orm.StoreTemplateObject{}
	e := t.DecryptObject(skey, o.Object())
	if e != nil {
		r.Abort(4998 /* TODO: ERROR [Invalid Store Key] */, nil)
		return
	}

	r.SetLocal("store-template-object", t)
}

func EncryptStoreObject(r rpf.GINProcessor, c *gin.Context) {
	o := r.MustGet("store-object").(*orm.StoreObject)
	ot := r.MustGet("store-template-object").(*orm.StoreTemplateObject)
	skey := r.MustGet("store-key").([]byte)

	// Encrypt Object
	ebs, e := ot.EncryptObject(skey)
	if e != nil {
		r.Abort(4998 /* TODO: ERROR [Failed to Encrypot Object] */, nil)
	}

	o.SetObject(ebs)
}
