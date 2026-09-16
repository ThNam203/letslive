package transaction

import (
	"context"
	"sen1or/letslive/finance/domains"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/gofrs/uuid/v5"
)

// UpdateStatus performs a status-only transition. The transactions_status_update_only
// trigger rejects any change away from 'completed', so terminal ledgered state is safe.
func (r postgresTransactionRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domains.ProcessStatus) error {
	cmd, err := r.dbConn.Exec(ctx, `update transactions set status = $1 where id = $2`, status, id)
	if err != nil {
		logger.Errorf(ctx, "db update error [updatetransactionstatus: %v]", err)
		return domains.ErrDatabaseQuery
	}
	if cmd.RowsAffected() == 0 {
		return domains.ErrTransactionNotFound
	}
	return nil
}
