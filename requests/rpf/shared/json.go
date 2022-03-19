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
	"encoding/json"
	"fmt"
	"strings"

	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

func RequestExtractJSON(r rpf.GINProcessor, c *gin.Context) {
	// Is JSON Body?
	app := c.GetHeader("content-type")
	fmt.Println(app)
	if (app == "") || (strings.Index(app, "application/json") < 0) {
		r.Abort(5201, nil)
		return
	}

	// Extract JSON Object from Request
	var x map[string]interface{}
	e := c.BindJSON(&x)
	if e != nil {
		r.Abort(5201, nil)
		return
	}

	// Set JSON Message
	r.SetLocal("request-json", x)
}

func JSONStringify(r rpf.GINProcessor, c *gin.Context) {
	o := r.Get("json")
	if o == nil {
		r.SetLocal("json-string", "")
		return
	}

	// Is Map Interface?
	j, ok := o.(map[string]interface{})
	if !ok { // NO: Abort
		r.Abort(4998 /* TODO: Error [Invalid JSON Object] */, nil)
		return
	}

	// Converted to String?
	s, err := json.Marshal(j)
	if err != nil { // NO: Abort
		r.Abort(4998 /* TODO: Error [Invalid JSON Object] */, nil)
		return
	}

	// Set JSON String
	r.SetLocal("json-string", string(s))
}
