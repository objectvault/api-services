// cSpell:ignore dword, ferreira, otype, paulo, qword
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

import (
	"math/rand"
	"time"
)

// KNOWN TYPES
const OTYPE_NOTSET = uint16(0x0000)     // Object Type NOT SET
const OTYPE_USER = uint16(0x0001)       // USER Object
const OTYPE_ORG = uint16(0x0002)        // ORGANIZATION Object
const OTYPE_STORE = uint16(0x0003)      // STORE Object
const OTYPE_ACTION = uint16(0x00FB)     // ACTION Object
const OTYPE_REQUEST = uint16(0x00FC)    // REQUEST Object
const OTYPE_KEY = uint16(0x00FD)        // KEY Object
const OTYPE_INVITATION = uint16(0x00FE) // INVITATION Object
const OTYPE_OTHER = uint16(0x00FF)      // OTHER Object Type

// BITS for VARIOUS SIZES
const QWORD_BITS = 64
const DWORD_BITS = 32
const WORD_BITS = 16
const BYTE_BITS = 8

// SEPARATION of PARTS
const SHARD_QWORD_MASK = uint64(0xF0000FFF00000000) // UPPER DWORD
const SHARD_QWORD_OFFSET = 32                       // BIT OFFSERR in QWORD
const LOCAL_QWORD_MASK = uint64(0x00000000FFFFFFFF) // LOWER DWORD

// SHARD GROUP BIT MANAGEMENT
const SHARD_GROUP_BITS = 4                                             // NUMBER of Bits Used for SHARD GROUP
const SHARD_GROUP_DWORD_MASK = uint32(0xF0000000)                      // ID for SHARD GROUP
const SHARD_GROUP_DWORD_OFFSET_MASK = uint32(0x0000000F)               // ID for SHARD GROUP
const SHARD_GROUP_WORD_OFFSET_MASK = uint16(0x000F)                    // ID for SHARD GROUP
const SHARD_GROUP_DWORD_OFFSET = DWORD_BITS - SHARD_GROUP_BITS         // BIT OFFSET in DWORD (uint32)
const SHARD_GROUP_QWORD_OFFSET = SHARD_GROUP_DWORD_OFFSET + DWORD_BITS // BIT OFFSET in QWORD (uint64)

// SHARD ID BIT MANAGEMENT
const SHARD_ID_BITS = 12                                            // NUMBER of Bits Used for SHARD ID
const SHARD_ID_DWORD_MASK = uint32(0x00000FFF)                      // ID of SHARD in GROUP
const SHARD_ID_DWORD_OFFSET = 0                                     // BIT OFFSET in DWORD (uint32)
const SHARD_ID_QWORD_OFFSET = SHARD_GROUP_DWORD_OFFSET + DWORD_BITS // BIT OFFSET in QWORD (uint64)

// OBJECT TYPE BIT MANAGEMENT
const OTYPE_BITS = 8                                       // OBJECT TYPE in GLOBAL ID
const OTYPE_DWORD_MASK = uint32(0x00FF0000)                // OBJECT TYPE BITS
const OTYPE_DWORD_OFFSET = 16                              // BIT OFFSET in DWORD (uint32)
const OTYPE_WORD_OFFSET_MASK = uint16(0x00FF)              // OBJECT TYPE for GLOBAL ID
const OTYPE_QWORD_OFFSET = OTYPE_DWORD_OFFSET + DWORD_BITS // BIT OFFSET in QWORD (uint64)

// LOCAL ID BIT MANAGEMENT
const LOCAL_ID_BITS = 32        // NUMBER of BITS Used for LOCAL ID
const LOCAL_ID_DWORD_OFFSET = 0 // BIT OFFSET in DWORD (uint32)
const LOCAL_ID_QWORD_OFFSET = 0 // BIT OFFSET in QWORD (uint64)

// GLOBAL OBJECT CONSTANTS
const SYSTEM_ADMINISTRATOR = uint64(0x1000000000000) // STAR System Administrator
const SYSTEM_ORGANIZATION = uint64(0x2000000000000)  // Global Management Organization

func ShardInfoFromID(global uint64) uint32 {
	qword := global & SHARD_QWORD_MASK
	dword := uint32(qword >> SHARD_QWORD_OFFSET)
	return dword
}

func ShardGroupFromID(global uint64) uint16 {
	dword := ShardInfoFromID(global)
	group := uint16((dword & SHARD_GROUP_DWORD_MASK) >> SHARD_GROUP_DWORD_OFFSET)
	return group
}

func ShardFromID(global uint64) uint32 {
	dword := ShardInfoFromID(global)
	id := (dword & SHARD_ID_DWORD_MASK)
	return id
}

func ObjectTypeFromID(global uint64) uint16 {
	dword := uint32(global >> SHARD_QWORD_OFFSET)
	id := uint16((dword & OTYPE_DWORD_MASK) >> OTYPE_DWORD_OFFSET)
	return id
}

func LocalIDFromID(global uint64) uint32 {
	qword := global & LOCAL_QWORD_MASK
	id := uint32(qword)
	return id
}

func IsObjectOfType(global uint64, t uint16) bool {
	return ObjectTypeFromID(global) == t
}

// Local Random Number Generator
var randgen *rand.Rand

func RandomShardID() uint32 {
	// Do we need to initialize a Random Number Generator?
	if randgen == nil { // YES
		randgen = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	shard := randgen.Uint32() & SHARD_ID_DWORD_MASK
	return shard
}

func RandomGlobalID(group uint16, otype uint16, id uint32) uint64 {
	// Do we need to initialize a Random Number Generator?
	if randgen == nil { // YES
		randgen = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	g := uint32(group & SHARD_GROUP_WORD_OFFSET_MASK)
	t := uint32(otype & OTYPE_WORD_OFFSET_MASK)
	shard := randgen.Uint32() & SHARD_ID_DWORD_MASK

	return shardGlobalID(g, t, shard, id)
}

func ShardGlobalID(group uint16, otype uint16, shard uint32, id uint32) uint64 {
	g := uint32(group & SHARD_GROUP_WORD_OFFSET_MASK)
	t := uint32(otype & OTYPE_WORD_OFFSET_MASK)
	s := shard & SHARD_ID_DWORD_MASK
	return shardGlobalID(g, t, s, id)
}

func shardGlobalID(group uint32, otype uint32, shard uint32, id uint32) uint64 {
	g := group << SHARD_GROUP_DWORD_OFFSET
	t := otype << OTYPE_DWORD_OFFSET
	gid := uint64(g | t | shard)
	gid = gid << DWORD_BITS
	gid = gid | uint64(id)
	return gid
}
