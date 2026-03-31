package general

import (
	"sen1or/letslive/user/handlers/basehandler"

	"github.com/jackc/pgx/v5/pgxpool"
)

type GeneralHandler struct {
	basehandler.BaseHandler
	DB *pgxpool.Pool
}

func NewGeneralHandler(db *pgxpool.Pool) *GeneralHandler {
	return &GeneralHandler{DB: db}
}
