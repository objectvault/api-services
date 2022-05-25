package queue

// cSpell:ignore amqp

import (
	"github.com/gin-gonic/gin"

	rpf "github.com/objectvault/goginrpf"

	"github.com/objectvault/queue-interface/messages"
	"github.com/objectvault/queue-interface/queue"
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

func SendQueueMessage(r rpf.GINProcessor, c *gin.Context) {
	// Get the Required Activation Message
	m := r.MustGet("queue-message")

	// Create Queue Message
	msg := messages.QueueMessage{}
	msg.SetID("TODO GENERATE ID")
	msg.SetMessage(m)

	// Open Queue Connection
	mq := c.MustGet("mq-connection").(*queue.AMQPServerConnection)

	// Can we open a connection?
	_, err := mq.OpenConnection()
	if err != nil { // NO: Abort
		r.SetResponseCode(2490)
		return
	}

	// Published Message to Queue?
	q := r.MustGet("queue").(string)
	err = mq.QueuePublishJSON("write", q, msg)
	if err != nil { // NO: Failed to Publish Message to Queue
		r.SetResponseCode(2490)
	}
}
