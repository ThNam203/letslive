package payment

import (
	"context"
	"sen1or/letslive/finance/domains"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/gofrs/uuid/v5"
)

func (r postgresPaymentRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domains.ProcessStatus) error {
	cmd, err := r.dbConn.Exec(ctx, `update payments set status = $1 where id = $2`, status, id)
	if err != nil {
		logger.Errorf(ctx, "db update error [updatepaymentstatus: %v]", err)
		return domains.ErrDatabaseQuery
	}
	if cmd.RowsAffected() == 0 {
		return domains.ErrPaymentNotFound
	}
	return nil
}
