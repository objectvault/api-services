// cSpell:ignore bson, paulo ferreira
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

func boolToMySQL(f bool) uint8 {
	if f {
		return 1
	}

	return 0
}

func mySQLtoBool(v uint8) bool {
	if v == 0 {
		return false
	}

	return true
}
