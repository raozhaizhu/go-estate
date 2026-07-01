package worker

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
)

func (p *redisTaskProcessor) HandleDeleteSessionsTask(ctx context.Context, t *asynq.Task) error {
	var payload DeleteSessionsPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	// log.Println("获取新任务, 荷载为: ", payload)

	if err := p.sessionCache.BatchDelete(ctx, payload.JTIs); err != nil {
		return err
	}

	return nil
}
