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
	"fmt"
	"strings"

	"github.com/objectvault/api-services/requests/rpf/utils"
	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

func ExtractGINParameterOrg(r rpf.GINProcessor, c *gin.Context) {
	// Initial Post Parameter Tests
	id, message := utils.ValidateGinParameter(c, "org", true, true, false)
	if message != "" {
		fmt.Println(message)
		r.Abort(3100, nil)
		return
	}

	// Cleanup ID
	id = strings.TrimSpace(id)
	id = strings.ToLower(id)

	// See if it is valid
	iid, message := utils.ValidateOrgReference(id)
	if message != "" {
		fmt.Println(message)
		r.Abort(3100, nil)
		return
	}

	r.SetLocal("request-org", iid)
}

func ExtractGINParameterOrgID(r rpf.GINProcessor, c *gin.Context) {
	// Initial Post Parameter Tests
	id, message := utils.ValidateGinParameter(c, "org", true, true, false)
	if message != "" {
		fmt.Println(message)
		r.Abort(3100, nil)
		return
	}

	// See if it is valid
	iid, message := utils.ValidateOrgID(id)
	if message != "" {
		fmt.Println(message)
		r.Abort(3100, nil)
		return
	}

	r.SetLocal("request-org", iid)
}
