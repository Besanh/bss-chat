package model

import (
	"errors"

	"github.com/uptrace/bun"
)

type ChatQueue struct {
	*Base
	bun.BaseModel       `bun:"table:chat_queue,alias:cq"`
	TenantId            string               `json:"tenant_id" bun:"tenant_id,type:uuid,notnull"`
	QueueName           string               `json:"queue_name" bun:"queue_name,type:text,notnull"`
	Description         string               `json:"description" bun:"description,type:text"`
	ChatRoutingId       string               `json:"chat_routing_id" bun:"chat_routing_id,type:uuid,notnull"`
	ChatRouting         *ChatRouting         `json:"chat_routing" bun:"rel:has-one,join:chat_routing_id=id"`
	ConnectionQueues    *ConnectionQueue     `json:"chat_connection_queues" bun:"rel:has-one,join:id=queue_id"`
	ChatQueueUser       []*ChatQueueUser     `json:"chat_queue_user" bun:"rel:has-many,join:id=queue_id"`
	ManageQueueId       string               `json:"manage_queue_id" bun:"manage_queue_id,type:uuid,default:null"`
	ChatManageQueueUser *ChatManageQueueUser `json:"manage_queue_user" bun:"rel:has-one,join:manage_queue_id=id"`
	Status              string               `json:"status" bun:"status,notnull"`
}

type ChatQueueRequest struct {
	QueueName     string   `json:"queue_name"`
	Description   string   `json:"description"`
	ConnectionId  []string `json:"connection_id"`
	ChatRoutingId string   `json:"chat_routing_id"`
	Status        string   `json:"status"`
}

func (m *ChatQueueRequest) Validate() error {
	if len(m.QueueName) < 1 {
		return errors.New("queue name is required")
	}
	if len(m.ChatRoutingId) < 1 {
		return errors.New("chat routing id is required")
	}
	return nil
}
