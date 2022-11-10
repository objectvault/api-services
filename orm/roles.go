// cSpell:ignore ferreira, paulo
package orm

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
	"strconv"
	"strings"
)

type I_Roles interface {
	Roles() []uint32
	IsRolesEmpty() bool
	HasRole(role uint32) bool
	HasExactRole(role uint32) bool
	GetCategoryRole(category uint16) uint32
	GetSubCategoryRole(subcategory uint16) uint32
	AddRole(role uint32) bool
	AddRoles(roles []uint32) bool
	RemoveRole(role uint32) bool
	RemoveCategory(category uint16) bool
	RemoveExactRole(role uint32) bool
	RemoveRoles(roles []uint32) bool
	RemoveAllRoles() bool
	RolesFromCSV(csv string) bool
	RolesToCSV() string
}

// FUNCTIONS
const FUNCTION_READ = 0x0001
const FUNCTION_LIST = 0x0002
const FUNCTION_CREATE = 0x0100
const FUNCTION_UPDATE = 0x0200
const FUNCTION_DELETE = 0x0400

// COMBINED READ and LIST
const FUNCTION_READ_LIST = 0x0003

/* NOTES: Function
 * Does UPDATE Function Automatically imply read?
 * In most cases, updating objects requires that you load(read) the object,
 * and then modify it???
 * Can an object be updated, if you can't read?
 * Can an object be created, if you can't read?
 */

// FUNCTION GROUPS
const FUNCTION_ALL = 0xFFFF
const FUNCTION_READONLY = 0x00FF
const FUNCTION_MODIFY = 0xFF00

// Object: CATEGORIES //
const CATEGORY_SYSTEM = 0x0100 // SYSTEM ROLES
const CATEGORY_ORG = 0x0200    // ORGANIZATION ROLES
const CATEGORY_STORE = 0x0300  // STORE ROLES

// Management: SUB-CATEGORIES //
const SUBCATEGORY_CONF = 0x0001     // Configuration Management
const SUBCATEGORY_USER = 0x0002     // User Management
const SUBCATEGORY_ROLES = 0x0003    // Roles Management
const SUBCATEGORY_INVITE = 0x0004   // Invitation Management
const SUBCATEGORY_ORG = 0x0005      // Organization Management
const SUBCATEGORY_STORE = 0x0006    // Store Management
const SUBCATEGORY_OBJECT = 0x0007   // Store Object Management
const SUBCATEGORY_TEMPLATE = 0x0008 // Template Management

func RoleIsValid(role uint32) bool {
	return (RoleCategory(role) != 0) &&
		(RoleFunctions(role) != 0)
}

func Role(category, function uint16) uint32 {
	return uint32(category)<<16 | uint32(function)
}

func RoleCategory(role uint32) uint16 {
	return uint16((role & 0xFFFF0000) >> 16)
}

func RoleSubCategory(role uint32) uint16 {
	return uint16((role & 0x00FF0000) >> 16)
}

func RoleFunctions(role uint32) uint16 {
	return uint16(role & 0x0000FFFF)
}

func RoleMatchCategory(category uint16, role uint32) bool {
	ct := RoleCategory(role)
	return category == ct
}

func RoleMatchSubCategory(subcategory uint16, role uint32) bool {
	ct := RoleSubCategory(role)
	return (subcategory & 0x00FF) == ct
}

func RoleMatchFunctions(from, to uint32) bool {
	ff := RoleFunctions(from)
	ft := RoleFunctions(to)
	return (ff & ft) == ff
}

func RoleMatchExactFunctions(from, to uint32) bool {
	ff := RoleFunctions(from)
	ft := RoleFunctions(to)
	return ff == ft
}

func RoleAddFunctions(functions uint32, role uint32) uint32 {
	sf := RoleFunctions(functions)
	df := RoleFunctions(role)
	return Role(RoleCategory(role), sf|df)
}

func RoleRemoveFunctions(functions, role uint32) uint32 {
	sf := RoleFunctions(functions)
	df := RoleFunctions(role)
	return Role(RoleCategory(role), df&^sf)
}

func RoleMatch(from, to uint32) bool {
	return RoleMatchCategory(RoleCategory(from), to) && RoleMatchFunctions(from, to)
}

func RoleExactMatch(from, to uint32) bool {
	return RoleMatchCategory(RoleCategory(from), to) && RoleMatchExactFunctions(from, to)
}

type S_Roles struct {
	I_Roles
	roles []uint32
}

func (o *S_Roles) Roles() []uint32 {
	return o.roles
}

func (o *S_Roles) IsRolesEmpty() bool {
	return len(o.roles) == 0
}

func (o *S_Roles) HasRole(role uint32) bool {
	for _, r := range o.roles {
		if RoleMatch(role, r) {
			return true
		}
	}

	return false
}

func (o *S_Roles) HasCategory(category uint16) bool {
	for _, r := range o.roles {
		if RoleMatchCategory(category, r) {
			return true
		}
	}

	return false
}

func (o *S_Roles) HasSubCategory(subcategory uint16) bool {
	for _, r := range o.roles {
		if RoleMatchSubCategory(subcategory, r) {
			return true
		}
	}

	return false
}

func (o *S_Roles) HasExactRole(role uint32) bool {
	for _, r := range o.roles {
		if RoleExactMatch(role, r) {
			return true
		}
	}

	return false
}

func (o *S_Roles) AddRole(role uint32) bool {
	if RoleIsValid(role) {
		for i, r := range o.roles {
			ct := RoleCategory(role)
			if RoleMatchCategory(ct, r) {
				if !RoleMatchFunctions(role, r) {
					rn := RoleAddFunctions(role, r)
					if r != rn {
						o.roles[i] = rn
						return true
					}
				}

				return false
			}
		}

		o.roles = append(o.roles, role)
		return true
	}

	return false
}

func (o *S_Roles) GetCategoryRole(category uint16) uint32 {
	for _, r := range o.roles {
		if RoleMatchCategory(category, r) {
			return r
		}
	}

	return 0
}

func (o *S_Roles) GetSubCategoryRole(subcategory uint16) uint32 {
	for _, r := range o.roles {
		if RoleMatchSubCategory(subcategory, r) {
			return r
		}
	}

	return 0
}

func (o *S_Roles) AddRoles(roles []uint32) bool {
	modified := false
	for _, r := range roles {
		modified = o.AddRole(r) || modified
	}

	return modified
}

func (o *S_Roles) RemoveRole(role uint32) bool {
	if RoleIsValid(role) {
		for i, r := range o.roles {
			ct := RoleCategory(role)
			if RoleMatchCategory(ct, r) {
				if !RoleMatchFunctions(role, r) {
					rn := RoleRemoveFunctions(role, r)
					if (rn > 0) && (rn != r) {
						o.roles[i] = rn
						return true
					} else if rn == 0 {
						o.roles = o.removeRoleAtIndice(i)
						return true
					}
				}
			}
		}
	}

	return false
}

func (o *S_Roles) RemoveCategory(category uint16) bool {
	if category > 0 {
		for i, r := range o.roles {
			if RoleMatchCategory(category, r) {
				o.roles = o.removeRoleAtIndice(i)
				return true
			}
		}
	}

	return false
}

func (o *S_Roles) RemoveExactRole(role uint32) bool {
	if RoleIsValid(role) {
		for i, r := range o.roles {
			if RoleExactMatch(role, r) {
				o.roles = o.removeRoleAtIndice(i)
				return true
			}
		}
	}

	return false
}

func (o *S_Roles) RemoveRoles(roles []uint32) bool {
	modified := false
	for _, r := range roles {
		modified = o.RemoveRole(r) || modified
	}

	return modified
}

func (o *S_Roles) RemoveAllRoles() bool {
	if len(o.roles) == 0 {
		return false
	}

	o.roles = []uint32{}
	return true
}

func (o *S_Roles) RolesFromCSV(csv string) bool {
	o.roles = []uint32{}

	list := strings.Split(csv, ",")
	for i := 0; i < len(list); i++ {
		rs := strings.TrimSpace(list[i])
		if rs != "" {
			rui, err := strconv.ParseUint(rs, 10, 32)
			if err == nil && RoleIsValid(uint32(rui)) {
				o.roles = append(o.roles, uint32(rui))
			}
		}
	}

	return true
}

func (o *S_Roles) RolesToCSV() string {
	if len(o.roles) == 0 {
		return ""
	} else {
		s := make([]string, len(o.roles))
		for i, v := range o.roles {
			s[i] = strconv.FormatUint(uint64(v), 10)
		}
		return strings.Join(s, ",")
	}
}

func (o *S_Roles) removeRoleAtIndice(i int) []uint32 {
	l := len(o.roles)
	if l == 1 {
		return []uint32{}
	}

	if i == 0 {
		return o.roles[1:]
	} else if i == (l - 1) {
		return o.roles[:l-1]
	} else {
		return append(o.roles[:i-1], o.roles[i+1:]...)
	}
}
