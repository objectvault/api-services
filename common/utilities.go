// cSpell:ignore gonic, orgs, paulo, ferreira
package common

/*
 * This file is part of the ObjectVault Project.
 * Copyright (C) 2020-2022 Paulo Ferreira <vault at sourcenotes.org>
 *
 * This work is published under the GNU AGPLv3.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

import "time"

// GLOBAL HELPERs //
// If Empty String return nil (Allows Conversion Betwee String and String Pointer)
func StringNilOnEmpty(s string) *string {
	if s == "" {
		return nil
	}

	return &s
}

// If nil String Pointer return Empty String (Allows Conversion Betwee String and String Pointer)
func StringEmptyOnNil(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}

// UTCTimeStamp Return UTC Time Stamp String in RFC 3339
func UTCTimeStamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}
