package content

import (
	"context"

	"github.com/gofrs/uuid/v5"
)

type ContentGateway interface {
	SetAuthorDisabled(ctx context.Context, userId uuid.UUID, disabled bool) error
	ReplaceDisabledAuthors(ctx context.Context, userIds []uuid.UUID) error
}
