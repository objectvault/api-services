// cSpell:ignore ferreira, paulo
package action

/*
 * This file is part of the ObjectVault Project.
 * Copyright (C) 2020-2022 Paulo Ferreira <vault at sourcenotes.org>
 *
 * This work is published under the GNU AGPLv3.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

const STATE_REGISTERED = 0x0000 // DEFAULT: Action Registered
const STATE_QUEUED = 0x0010     // Action Queued
const STATE_PROCESSED = 0x00ff  // Action Processed
