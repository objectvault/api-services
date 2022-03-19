// cSpell:ignore goginrpf, gonic, paulo ferreira
package object

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
	"fmt"
	"strings"

	"github.com/objectvault/api-services/requests/rpf/utils"
	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

func ExtractGINParameterObjectID(r rpf.GINProcessor, c *gin.Context) {
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

	r.SetLocal("request-object-id", uint64(*id))
}
