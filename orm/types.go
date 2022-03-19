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

// Map External Field Name to ORM Field Name
type TMapFieldNameExternalToORM = func(p string) string

// Map External Field Value to ORM Field Value
type TMapFieldValueExternalToORM = func(string, interface{}) interface{}
