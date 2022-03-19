// cSpell:ignore goginrpf, gonic, paulo ferreira
package user

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
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/objectvault/api-services/requests/rpf/utils"
	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

func ExtractGINParameterUser(r rpf.GINProcessor, c *gin.Context) {
	// Initial Post Parameter Tests
	id, message := utils.ValidateGinParameter(c, "user", true, true, false)
	if message != "" {
		fmt.Println(message)
		r.Abort(3100, nil)
		return
	}

	iid, message := utils.ValidateUserReference(id)
	if message != "" {
		fmt.Println(message)
		r.Abort(3100, nil)
		return
	}

	r.SetLocal("request-user", iid)
}

func ExtractGINParameterObject(r rpf.GINProcessor, c *gin.Context) {
	// Initial Post Parameter Tests
	v, message := utils.ValidateGinParameter(c, "object", true, true, false)
	if message != "" {
		fmt.Println(message)
		r.Abort(3100, nil)
		return
	}

	// Cleanup ID
	v = strings.TrimSpace(v)

	// See if it is valid
	id, message := utils.ValidateObjectID("object", v)
	if message != "" {
		fmt.Println(message)
		r.Abort(3100, nil)
		return
	}

	r.SetLocal("request-object", *id)
}

func ExtractFormParameterCredentials(r rpf.GINProcessor, c *gin.Context) {
	// Initial Post Parameter Tests
	hash, message := utils.ValidateFormParameter(c, "credentials", true, true, false)
	if message != "" {
		r.Abort(3100, nil)
		return
	}

	hash, message = utils.ValidateHash(hash)
	if message != "" {
		r.Abort(3100, nil)
		return
	}

	bytes, e := hex.DecodeString(hash)
	if e != nil {
		r.Abort(3100, nil)
		return
	}

	r.SetLocal("user-credentials", bytes)
}
