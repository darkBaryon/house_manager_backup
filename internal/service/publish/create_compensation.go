package publish

import (
	"context"
	"fmt"

	"house-manager/pkg/errcode"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func rollbackCreateFailure(ctx context.Context, rollback func(context.Context, bson.ObjectID) error, id bson.ObjectID, entityName string, cause error) error {
	if rollback == nil || id.IsZero() {
		return cause
	}
	if err := rollback(ctx, id); err != nil {
		return errcode.InternalError.WithError(fmt.Errorf("%s side effects failed: %v; rollback failed: %w", entityName, cause, err))
	}
	return cause
}
