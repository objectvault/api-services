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

// Map ORM List Entry to Export List Entry
type TMapListEntryORMtoExport = func(interface{}) interface{}

// Map ORM Field Name to External Name
type TMapFieldNameORMToExternal = func(p string) string
