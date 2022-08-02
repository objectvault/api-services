// cSpell:ignore goginrpf, gonic, paulo ferreira
package shared

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

func ExtractGINParameterBooleanValue(r rpf.GINProcessor, c *gin.Context) {
	// Initial Post Parameter Tests
	v, message := utils.ValidateGinParameter(c, "bool", true, true, false)
	if message != "" {
		fmt.Println(message)
		r.Abort(3100, nil)
		return
	}

	// Cleanup Value
	v = strings.TrimSpace(v)
	v = strings.ToLower(v)

	switch v {
	case "0":
		fallthrough
	case "f":
		fallthrough
	case "false":
		r.SetLocal("request-value", false)
	case "1":
		fallthrough
	case "t":
		fallthrough
	case "true":
		r.SetLocal("request-value", true)
	default:
		fmt.Println("Route parameter [bool] does not contain a 'boolean' value")
		r.Abort(3100, nil)
		return
	}
}

func ExtractGINParameterIntValue(r rpf.GINProcessor, c *gin.Context) {
	// Initial Post Parameter Tests
	v, message := utils.ValidateGinParameter(c, "int", true, true, false)
	if message != "" {
		fmt.Println(message)
		r.Abort(3100, nil)
		return
	}

	// Cleanup Value
	v = strings.TrimSpace(v)
	i, err := strconv.ParseInt(v, 10, 64)

	if err != nil {
		fmt.Println("Route parameter [int] does not contain an 'integer' value")
		r.Abort(3100, nil)
		return
	}

	r.SetLocal("request-value", i)
}

func ExtractGINParameterUINTValue(r rpf.GINProcessor, c *gin.Context) {
	// Initial Post Parameter Tests
	v, message := utils.ValidateGinParameter(c, "uint", true, true, false)
	if message != "" {
		fmt.Println(message)
		r.Abort(3100, nil)
		return
	}

	// Cleanup Value
	v = strings.TrimSpace(v)
	u, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		fmt.Println("Route parameter [uint] does not contain an 'unsigned integer' value")
		r.Abort(3100, nil)
		return
	}

	r.SetLocal("request-value", u)
}

func ExtractGINParameterEmail(r rpf.GINProcessor, c *gin.Context) {
	// Initial Post Parameter Tests
	email, message := utils.ValidateGinParameter(c, "email", true, true, false)
	if message != "" {
		fmt.Println(message)
		r.Abort(3100, nil)
		return
	}

	// Cleanup Email
	email = strings.TrimSpace(email)
	email = strings.ToLower(email)

	// See if it is valid
	email, message = utils.ValidateEmailFormat(email)
	if message != "" {
		fmt.Println(message)
		r.Abort(3100, nil)
		return
	}

	r.SetLocal("request-email", email)
}

func ExtractGINParameterGUID(r rpf.GINProcessor, c *gin.Context) {
	// Initial Post Parameter Tests
	guid, message := utils.ValidateGinParameter(c, "guid", true, true, false)
	if message != "" {
		fmt.Println(message)
		r.Abort(3100, nil)
		return
	}

	// Cleanup GUID
	guid = strings.TrimSpace(guid)
	guid = strings.ToLower(guid)

	// See if it is valid
	guid, message = utils.ValidateGUIDFormat(guid)
	if message != "" {
		fmt.Println(message)
		r.Abort(3100, nil)
		return
	}

	r.SetLocal("request-guid", guid)
}
