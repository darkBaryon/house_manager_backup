package publish

import (
	"context"
	"fmt"

	"house-manager/pkg/errcode"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func rollbackCreateFailure(ctx context.Context, rollback func(context.Context, bson.ObjectID) error, id bson.ObjectID, entityName string, cause error) error {
	if rollback == nil || id.IsZero() {
		logPublishResult(ctx, entityName+".rollback.skipped", entityName+".rollback.skipped", cause, "entity_id", id.Hex())
		return cause
	}
	logPublishWarn(ctx, entityName+".rollback.start", "entity_id", id.Hex(), "cause", cause)
	if err := rollback(ctx, id); err != nil {
		logPublishError(ctx, entityName+".rollback.failed", "entity_id", id.Hex(), "cause", cause, "error", err)
		return errcode.InternalError.WithError(fmt.Errorf("%s side effects failed: %v; rollback failed: %w", entityName, cause, err))
	}
	logPublishInfo(ctx, entityName+".rollback.success", "entity_id", id.Hex())
	return cause
}
