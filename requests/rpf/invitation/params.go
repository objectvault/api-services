// cSpell:ignore goginrpf, gonic, orgs, paulo, ferreira
package invitation

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
	"strconv"
	"strings"

	"github.com/objectvault/api-services/requests/rpf/utils"

	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

func ExtractGINParameterUID(r rpf.GINProcessor, c *gin.Context) {
	// Initial Post Parameter Tests
	uid, message := utils.ValidateGinParameter(c, "uid", true, true, false)
	if message != "" {
		fmt.Println(message)
		r.Abort(3100, nil)
		return
	}

	// Convert to Lower Case
	uid = strings.ToLower(uid)
	if !utils.IsValidUID(uid) { // NO: Invalid Invitation ID
		r.Abort(5998, nil) // TODO: Error Code [Invalid Invitation UID]
		return
	}

	r.SetLocal("request-invite-uid", uid)
}

func ExtractGINParameterInvitationID(r rpf.GINProcessor, c *gin.Context) {
	// Initial Post Parameter Tests
	sid, message := utils.ValidateGinParameter(c, "id", true, true, false)
	if message != "" {
		fmt.Println(message)
		r.Abort(3100, nil)
		return
	}

	// See if it is valid
	if !utils.IsValidID(sid) {
		fmt.Println(message)
		r.Abort(3100, nil)
		return
	}

	uid, e := strconv.ParseUint(sid, 10, 64)
	if e != nil {
		r.Abort(5998, nil) // TODO: Error Code [Invalid Invitation UID]
		return
	}

	r.SetLocal("request-invite-id", uid)
}
