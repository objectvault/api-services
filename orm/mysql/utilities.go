package mysql

/*
 * This file is part of the ObjectVault Project.
 * Copyright (C) 2020-2022 Paulo Ferreira <vault at sourcenotes.org>
 *
 * This work is published under the GNU AGPLv3.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

func BoolToMySQL(f bool) uint8 {
	if f {
		return 1
	}

	return 0
}

func MySQLtoBool(v uint8) bool {
	return v != 0
}
