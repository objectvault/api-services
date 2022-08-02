// cSpell:ignore ferreira, paulo
package request

/*
 * This file is part of the ObjectVault Project.
 * Copyright (C) 2020-2022 Paulo Ferreira <vault at sourcenotes.org>
 *
 * This work is published under the GNU AGPLv3.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

const STATE_ACTIVE = 0x0000 // DEFAULT: Request Created
const STATE_QUEUED = 0x00F0 // Request Expired
const STATE_CLOSED = 0x00FF // Request Closed
