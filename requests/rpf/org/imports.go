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
	// Extract and Validat JSON Message
	m := r.MustGet("request-json").(xjson.T_xMap)
	vmap := xjson.S_xJSONMap{Source: m}

	// Create Organization
	o := &orm.Organization{}

	// Organization ALIAS
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
		o.SetAlias(v.(string))
		return nil
	})

	// OPTIONAL: Organization Name
	vmap.Optional("title", nil, xjson.F_xToTrimmedString, nil, func(v interface{}) error {
		if v != nil {
			o.SetName(v.(string))
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

	// Save Organization
	r.Set("org", o)
}

func SystemUpdateFromJSON(r rpf.GINProcessor, c *gin.Context) {
	// Extract and Validate JSON MEssage
	m := r.MustGet("request-json").(xjson.T_xMap)
	vmap := xjson.S_xJSONMap{Source: m}

	// Get Organization
	o := r.MustGet("org").(*orm.Organization)

	// OPTIONAL: Organization Name
	vmap.Optional("alias", nil, func(v interface{}) (interface{}, error) {
		v, e := xjson.F_xToTrimmedString(v)
		if e != nil {
			return nil, e
		}

		s := strings.ToLower(v.(string))
		if !utils.IsValidOrgAlias(s) {
			return nil, errors.New("Valus is does not contains a valid alias")
		}
		return s, nil
	}, nil, func(v interface{}) error {
		if v != nil {
			o.SetAlias(v.(string))
		}
		return nil
	})

	// OPTIONAL: Organization Name
	vmap.Optional("name", nil, xjson.F_xToTrimmedString, nil, func(v interface{}) error {
		if v != nil {
			o.SetName(v.(string))
		}
		return nil
	})
}

func ManagerUpdateFromJSON(r rpf.GINProcessor, c *gin.Context) {
	// Extract and Validate JSON MEssage
	m := r.MustGet("request-json").(xjson.T_xMap)
	vmap := xjson.S_xJSONMap{Source: m}

	// Get Organization
	o := r.MustGet("org").(*orm.Organization)
	// STANDARD Organization Updates (can't modify alias)

	// OPTIONAL: Organization Name
	vmap.Optional("name", nil, xjson.F_xToTrimmedString, nil, func(v interface{}) error {
		if v != nil {
			o.SetName(v.(string))
		}
		return nil
	})
}
