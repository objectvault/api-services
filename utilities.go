// cSpell:ignore gonic, orgs, paulo, ferreira

/*
 * This file is part of the ObjectVault Project.
 * Copyright (C) 2020-2022 Paulo Ferreira <vault at sourcenotes.org>
 *
 * This work is published under the GNU AGPLv3.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */
package main

// LOCAL HELPER Methods //
// If Empty String return nil (Allows Conversion Betwee String and String Pointer)
func nilOnEmpty(s string) *string {
	if s == "" {
		return nil
	}

	return &s
}

// If nil String Pointer return Empty String (Allows Conversion Betwee String and String Pointer)
func emptyOnNil(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}
