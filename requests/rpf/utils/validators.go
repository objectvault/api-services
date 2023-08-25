// cSpell:ignore gonic, orgs, paulo, ferreira
package utils

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

	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

func ValidateHash(value string) (string, string) {
	if !IsValidPasswordHash(value) {
		return "", "Parameter Does NOT Contain Valid a Password HASH"
	}

	return value, ""
}

func ValidateEmailFormat(value string) (string, string) {
	if !IsValidEmail(value) {
		return "", "Parameter Does NOT Contain Valid a Email Address"
	}

	return value, ""
}

func ValidateGUIDFormat(value string) (string, string) {
	if !IsValidGUID(value) {
		return "", "Parameter Does NOT Contain Valid a GUID"
	}

	return value, ""
}

// Value Contains a Value that can pass as a User Reference
func ValidateUserID(value string) (interface{}, string) {
	// Is Valid Store ID
	if IsValidHexID(value) { // YES
		id, e := strconv.ParseUint(value[1:], 16, 64)
		if e == nil {
			return id, ""
		}
	}

	// ELSE: No
	return "", "Parameter Does NOT Contain Valid a User ID"
}

// Value Contains a Value that can pass as a User Reference
func ValidateUserReference(value string) (interface{}, string) {
	// Is Valid Store ID
	if IsValidHexID(value) { // YES
		id, e := strconv.ParseUint(value[1:], 16, 64)
		if e == nil {
			return id, ""
		}
	} else if IsValidUserName(value) || IsValidEmail(value) { // ELSE: Valid Alias or Email
		return value, ""
	}
	// ELSE: No
	return "", "Parameter Does NOT Contain Valid a User Reference"
}

// Value Contains a Value that can pass as a User Reference
func ValidateOrgID(value string) (interface{}, string) {
	// Is Valid Store ID
	if IsValidHexID(value) { // YES
		id, e := strconv.ParseUint(value[1:], 16, 64)
		if e == nil {
			return id, ""
		}
	}

	// ELSE: No
	return "", "Parameter Does NOT Contain Valid a Organization ID"
}

// Value Contains a Value that can pass as a Organization Reference
func ValidateOrgReference(value string) (interface{}, string) {
	// Is Valid Store ID
	if IsValidHexID(value) { // YES
		id, e := strconv.ParseUint(value[1:], 16, 64)
		if e == nil {
			return id, ""
		}
	} else if IsValidOrgAlias(value) { // ELSE: Valid Alias
		return value, ""
	}
	// ELSE: No
	return "", "Parameter Does NOT Contain Valid a Organization Reference"
}

func ValidateStoreReference(value string) (interface{}, string) {
	// Is Valid Store ID
	if IsValidHexID(value) { // YES
		id, e := strconv.ParseUint(value[1:], 16, 64)
		if e == nil {
			return id, ""
		}
	} else if IsValidStoreAlias(value) { // ELSE: Valid Alias
		return value, ""
	}
	// ELSE: No
	return "", "Parameter Does NOT Contain Valid a Store Reference"
}

func ValidateStoreID(value string) (interface{}, string) {
	// Is Valid Store ID
	if IsValidHexID(value) { // YES
		id, e := strconv.ParseUint(value[1:], 16, 64)
		if e == nil {
			return id, ""
		}
	}

	// ELSE: No
	return "", "Parameter Does NOT Contain Valid a Store ID"
}

func ValidateObjectID(name string, value string) (*uint64, string) {
	// Is Valid Hex Value?
	if IsValidHexID(value) { // YES
		id, e := strconv.ParseUint(value[1:], 16, 64)
		if e == nil {
			return &id, ""
		}
	}
	// ELSE: Is is it an unsigned integer?
	i, msg := ValidateUintParameter(name, value, false)
	if msg != "" { // NO
		return nil, msg
	}
	// ELSE: Valid UINT
	return i, ""
}

// Value Contains a Value that can pass as a Template Name
func ValidateTemplateName(value string) (interface{}, string) {
	// Cleanup Name
	value = strings.TrimSpace(value)
	value = strings.ToLower(value)

	// Is Valid Template Name
	if IsValidTemplateName(value) { // ELSE: Valid Alias or Email
		return value, ""
	}
	// ELSE: No
	return "", "Parameter Does NOT Contain Valid a Template Name"
}

func ValidateUintParameter(name string, value string, allowEmpty bool) (*uint64, string) {
	i, msg := ValidateIntParameter(name, value, allowEmpty)
	if msg != "" {
		return nil, msg
	}

	if *i < 0 {
		return nil, fmt.Sprintf("Required Parameter '%s' is not an unsigned integer", name)
	}

	u := uint64(*i)
	return &u, ""
}

// Basic Integer Parameter Validator
func ValidateIntParameter(name string, value string, allowEmpty bool) (*int64, string) {
	// Does it have a value?
	if value == "" { // NO
		// Is no Value Allowed
		if !allowEmpty { // NO: Error
			return nil, fmt.Sprintf("Required Parameter '%s' is missing", name)
		}
		// ELSE: No Value Set
		return nil, ""
	}

	// Convert String to Int
	i, err := strconv.ParseInt(value, 10, 64)

	// Is Valid Value?
	if err != nil { // NO: Error
		return nil, fmt.Sprintf("Required Parameter '%s' is not an integer", name)
	}

	// ELSE: Non Empty String Value
	return &i, ""
}

// Basic String Parameter Validator
func ValidateStringParameter(name string, value string, trim bool, allowEmpty bool) (string, string) {
	// Should we Trim Value?
	if trim { // YES
		value = strings.TrimSpace(value)
	}

	// Does it have a value?
	if value == "" { // NO
		// Is no Value Allowed
		if !allowEmpty { // NO: Error
			return "", fmt.Sprintf("Required Parameter '%s' is empty", name)
		}
		// ELSE: Empty String Allowed
	}
	// ELSE: Non Empty String Value
	return value, ""
}

// Basic GIN Request Parameter Validator
func ValidateGinParameter(c *gin.Context, name string, required bool, trim bool, allowEmpty bool) (string, string) {
	// Get Parameter
	value, exists := c.Params.Get(name)

	// Does it Exist?
	if !exists { // NO
		// Is it required?
		if required { // YES: Error
			return "", fmt.Sprintf("Missing Required ROUTE Parameter '%s'", name)
		}
		// ELSE: Not Required
		return "", ""
	}

	// Basic Value Validation
	return ValidateStringParameter(name, value, trim, allowEmpty)
}

// Basic GIN URL Parameter Validator
func ValidateURLParameter(c *gin.Context, name string, required bool, trim bool, allowEmpty bool) (string, string) {
	// Get Parameter
	value, exists := c.GetQuery(name)

	// Does it Exist?
	if !exists { // NO
		// Is it required?
		if required { // YES: Error
			return "", fmt.Sprintf("Missing Required URL Parameter '%s'", name)
		}
		// ELSE: Not Required
		return "", ""
	}

	// Basic Value Validation
	return ValidateStringParameter(name, value, trim, allowEmpty)
}

// Basic GIN Form Parameter Validator
func ValidateFormParameter(c *gin.Context, name string, required bool, trim bool, allowEmpty bool) (string, string) {
	value, exists := c.GetPostForm(name)

	// Does it Exist?
	if !exists { // NO
		// Is it required?
		if required { // YES: Error
			return "", "Missing Required FORM Parameter"
		}
		// ELSE: Not Required
		return "", ""
	}

	// Basic Value Validation
	return ValidateStringParameter(name, value, trim, allowEmpty)
}

// SINGLE/MULTI FIELD VALIDATION RPF HANDLERS //
func RPFReadyVFields(r rpf.GINProcessor, c *gin.Context) {
	// Initialize Field Validation
	r.Set("v_fields", make(map[string]string))
}

func RPFTestVFields(code int, r rpf.GINProcessor, c *gin.Context) {
	// Get Fields Error Message Map
	fields := r.Get("v_fields").(map[string]string)

	// Do we have any errors
	if len(fields) > 0 { // YES: Exit
		data := gin.H{
			"fields": fields,
		}
		r.Abort(code, &data)
	}
}
