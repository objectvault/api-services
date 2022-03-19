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
	"github.com/objectvault/api-services/common"
	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

func JSONResponse(r rpf.GINProcessor, c *gin.Context) {
	// Convert Status Code to HTTP Status Code and Message
	httpCode, msg := common.CodeToMessage(r.ResponseCode())

	// Create Response Message
	message := gin.H{
		"version": gin.H{
			"major": 1,
			"minor": 0,
		},
		"code":    r.ResponseCode(),
		"message": msg,
	}

	// Do we have any Extra Data?
	if r.ResponseData() != nil { // YES: Add to Message
		message["data"] = r.ResponseData()
	}

	// Set Request Response
	c.JSON(httpCode, message)
}
