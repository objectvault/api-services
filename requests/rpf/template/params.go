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
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/objectvault/api-services/requests/rpf/utils"

	rpf "github.com/objectvault/goginrpf"
)

func ExtractGINParameterTemplate(r rpf.GINProcessor, c *gin.Context) {
	// Initial Post Parameter Tests
	name, message := utils.ValidateGinParameter(c, "template", true, true, false)
	if message != "" {
		fmt.Println(message)
		r.Abort(3100, nil)
		return
	}

	// See if it is valid
	n, message := utils.ValidateTemplateName(name)
	if message != "" {
		fmt.Println(message)
		r.Abort(3100, nil)
		return
	}

	r.SetLocal("request-template", n.(string))
}
