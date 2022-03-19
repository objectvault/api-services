// cSpell:ignore goginrpf, gonic, paulo ferreira
package store

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
	"errors"
	"fmt"
	"strings"

	"github.com/objectvault/api-services/orm"
	"github.com/objectvault/api-services/requests/rpf/utils"
	"github.com/objectvault/api-services/xjson"

	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

func CreateFromJSON(r rpf.GINProcessor, c *gin.Context) {
	// Extract and Validat Post Parameters
	m := r.MustGet("request-json").(xjson.T_xMap)
	vmap := xjson.S_xJSONMap{Source: m}

	// Create Store Entry Based On JSON
	e := &orm.Store{}

	// Store ALIAS
	vmap.Required("alias", nil, func(v interface{}) (interface{}, error) {
		v, e := xjson.F_xToTrimmedString(v)
		if e != nil {
			return nil, e
		}

		s := strings.ToLower(v.(string))
		if !utils.IsValidOrgAlias(s) {
			return nil, errors.New("Valus is does not contains a valid alias")
		}
		return s, nil
	}, func(v interface{}) error {
		e.SetAlias(v.(string))
		return nil
	})

	// OPTIONAL: Organization Name
	vmap.Optional("title", nil, xjson.F_xToTrimmedString, nil, func(v interface{}) error {
		if v != nil {
			e.SetName(v.(string))
		}
		return nil
	})

	// Did we have an Error Processing the Map?
	if vmap.Error != nil {
		fmt.Println(vmap.Error)
		fmt.Println(vmap.StringSrc())
		r.Abort(5202, nil)
		return
	}

	// Save Store
	r.Set("store", e)
}

func UpdateFromJSON(r rpf.GINProcessor, c *gin.Context) {

	// Extract and Validate JSON MEssage
	m := r.MustGet("request-json").(xjson.T_xMap)
	vmap := xjson.S_xJSONMap{Source: m}

	// Get Store
	e := r.MustGet("store").(*orm.Store)

	// OPTIONAL: Store Name
	vmap.Optional("alias", nil, func(v interface{}) (interface{}, error) {
		v, e := xjson.F_xToTrimmedString(v)
		if e != nil {
			return nil, e
		}

		s := strings.ToLower(v.(string))
		if !utils.IsValidStoreAlias(s) {
			return nil, errors.New("Valus is does not contains a valid alias")
		}
		return s, nil
	}, nil, func(v interface{}) error {
		if v != nil {
			e.SetAlias(v.(string))
		}
		return nil
	})

	// OPTIONAL: Store Name
	vmap.Optional("name", nil, xjson.F_xToTrimmedString, nil, func(v interface{}) error {
		if v != nil {
			e.SetName(v.(string))
		}
		return nil
	})

	// Did we have an Error Processing the Map?
	if vmap.Error != nil {
		fmt.Println(vmap.Error)
		fmt.Println(vmap.StringSrc())
		r.Abort(5202, nil)
		return
	}
}
