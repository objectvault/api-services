package queue

// cSpell:ignore amqp

import (
	"github.com/gin-gonic/gin"

	rpf "github.com/objectvault/goginrpf"

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

func QueueActionMessage(r rpf.GINProcessor, c *gin.Context) {
	// Get the Required Activation Message
	m := r.MustGet("queue-action")

	// Open Queue Connection
	mq := c.MustGet("mq-connection").(*queue.AMQPServerConnection)

	// Can we open a connection?
	_, err := mq.OpenConnection()
	if err != nil { // NO: Abort
		r.SetResponseCode(2490)
		return
	}

	// Default Action Queue
	q := "action-incoming"

	// Do we have a non standard queue set?
	nsq := r.Get("queue")
	if nsq != nil {
		q = nsq.(string)
	}

	// Published Message to Queue?
	err = mq.QueuePublishJSON("write", q, m)
	if err != nil { // NO: Failed to Publish Message to Queue
		r.SetResponseCode(2490)
	}

	// Close the Connection
	mq.CloseConnection()
}

func SendQueueMessage(r rpf.GINProcessor, c *gin.Context) {
	// Get the Required Activation Message
	m := r.MustGet("queue-message")

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
	err = mq.QueuePublishJSON("write", q, m)
	if err != nil { // NO: Failed to Publish Message to Queue
		r.SetResponseCode(2490)
	}

	// Close the Connection
	mq.CloseConnection()
}
