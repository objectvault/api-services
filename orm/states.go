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

// FUNCTIONS
const STATE_INACTIVE = 0x0001 // User Locked Out of System (USE: Too Many Failed Password Attempts)
const STATE_BLOCKED = 0x0002  // User/Organization Blocked (USE: Administrator Blocked User Access)
const STATE_READONLY = 0x0004 // User/Organization Disabled All Modification Roles

// MARKERS
const STATE_SYSTEM = 0x1000 // SYSTEM User/Organization
const STATE_DELETE = 0x2000 // Object Marked for Deletion

// MASKS
const STATE_MASK_MARKERS = 0xF000   // MASK UPPER Bits
const STATE_MASK_FUNCTIONS = 0x0FFF // MASK LOWER BITS

func HasAnyStates(state, test uint16) bool {
	return (state & test) != 0
}

func HasAllStates(state, test uint16) bool {
	return (state & test) == test
}

func SetStates(state, set uint16) uint16 {
	return (state | set)
}

func ClearStates(state, clear uint16) uint16 {
	return state &^ clear
}

type States interface {
	State() uint16
	HasAnyStates(states uint16) bool
	HasAllStates(states uint16) bool
	SetStates(states uint16)
	ClearStates(states uint16)
}
