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

import "regexp"

// REGEXP - ID
var rMatchID = regexp.MustCompile("^[0-9][0-9]*$")

// REGEXP - ID
var rMatchHexID = regexp.MustCompile("^:[0-9a-f][0-9a-f]*$")

// REGEXP - Alias
var rMatchAlias = regexp.MustCompile("^[a-z][a-z0-9_.-]+$")

// REGEXP Organization Alias
// https://stackoverflow.com/questions/106179/regular-expression-to-match-dns-hostname-or-ip-address
// MATCH DNS as per RFC 1123
var rMatchOrgAlias = regexp.MustCompile("^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9_\\-]*[a-zA-Z0-9])\\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9_\\-]*[A-Za-z0-9])$")

// REGEXP - Password Hash (Unsalted) = SHA256
var rMatchPasswordHash = regexp.MustCompile("^[a-f0-9]{64}$")

// REGEXP - Invitation Unique ID SHA1
var rMatchUID = regexp.MustCompile("^[a-f0-9]{40}$")

// https://github.com/badoux/checkmail
var rMatchEmail = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// Roles CSV String
var rMatchRolesCSV = regexp.MustCompile(`^\s*\d+(\s*,\s*\d+)*\s*$`)

func IsValidEmail(v string) bool {
	return rMatchEmail.Match([]byte(v))
}

func IsValidUserName(v string) bool {
	return rMatchAlias.Match([]byte(v))
}

func IsValidOrgAlias(v string) bool {
	return rMatchOrgAlias.Match([]byte(v))
}

func IsValidStoreAlias(v string) bool {
	return rMatchAlias.MatchString(v)
}

func IsValidID(v string) bool {
	return rMatchID.MatchString(v)
}

func IsValidHexID(v string) bool {
	return rMatchHexID.MatchString(v)
}

func IsValidRolesCSV(v string) bool {
	return rMatchRolesCSV.MatchString(v)
}

func IsValidUserID(id string) bool {
	return IsValidID(id) || IsValidUserName(id) || IsValidEmail(id)
}

func IsValidOrgReference(id string) bool {
	return IsValidHexID(id) || IsValidOrgAlias(id)
}

func IsValidStoreID(id string) bool {
	return IsValidID(id) || IsValidStoreAlias(id)
}

func IsValidPasswordHash(p string) bool {
	return rMatchPasswordHash.Match([]byte(p))
}

func IsValidUID(p string) bool {
	return rMatchUID.Match([]byte(p))
}

func IsValidTemplateName(v string) bool {
	return rMatchAlias.MatchString(v)
}
