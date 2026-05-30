package payment

import (
	"context"
	"sen1or/letslive/finance/domains"
	"sen1or/letslive/finance/response"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/gofrs/uuid/v5"
)

func (r postgresPaymentRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domains.ProcessStatus) *response.Response[any] {
	cmd, err := r.dbConn.Exec(ctx, `update payments set status = $1 where id = $2`, status, id)
	if err != nil {
		logger.Errorf(ctx, "db update error [updatepaymentstatus: %v]", err)
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}
	if cmd.RowsAffected() == 0 {
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_PAYMENT_NOT_FOUND,
			nil,
			nil,
			nil,
		)
	}
	return nil
}
