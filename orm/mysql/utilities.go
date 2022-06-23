package mysql

import (
	"fmt"
	"time"
)

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

/* IMPORTANT NOTE:
 * MySQL stores timestamps in UTC, but serves them in local time, as set
 * on the MySQL Server HOST NODE.
 * This MEANS that the NODE on which the GO is Being Run has to be ON THE
 * SAME time zone setting as the MySQL HOST NODE
 */
func MySQLTimeStampToGoTime(t string) *time.Time {
	// Parse MySQL TimeStamp
	tm, e := time.Parse("2006-01-02 15:04:05", t)

	// NOTE: MySQL Server Should have Time Zone set to UTC
	// time.Parse() Returns time stamp relative to UTC if no Time Zone in String

	// Did parse genersate an error?
	if e != nil { // YES
		return nil
	}

	// Return Time Stam
	fmt.Println(tm)
	return &tm
}

func GoTimeToMySQLTimeStamp(t *time.Time) string {
	utc := t.UTC()
	formatted := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
		utc.Year(), utc.Month(), utc.Day(),
		utc.Hour(), utc.Minute(), utc.Second())
	return formatted
}
