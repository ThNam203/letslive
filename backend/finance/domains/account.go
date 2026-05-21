package domains

import (
	"context"
	response "sen1or/letslive/finance/response"
	"time"

	"github.com/gofrs/uuid/v5"
)

type AccountType string

const (
	AccountTypeUserWallet AccountType = "user_wallet"
	AccountTypePlatform   AccountType = "platform"
	AccountTypeEscrow     AccountType = "escrow"
	AccountTypeFee        AccountType = "fee"
)

type AccountStatus string

const (
	AccountStatusActive AccountStatus = "active"
	AccountStatusFrozen AccountStatus = "frozen"
	AccountStatusClosed AccountStatus = "closed"
)

type Account struct {
	Id        uuid.UUID     `json:"id" db:"id"`
	Type      AccountType   `json:"type" db:"type"`
	OwnerId   *uuid.UUID    `json:"ownerId" db:"owner_id"`
	Status    AccountStatus `json:"status" db:"status"`
	CreatedAt time.Time     `json:"createdAt" db:"created_at"`
}

type AccountBalance struct {
	AccountId    uuid.UUID  `json:"accountId" db:"account_id"`
	CurrencyCode string     `json:"currencyCode" db:"currency_code"`
	Balance      int64      `json:"-" db:"balance"`
	LastEntryId  *uuid.UUID `json:"lastEntryId" db:"last_entry_id"`
	UpdatedAt    time.Time  `json:"updatedAt" db:"updated_at"`
}

type AccountRepository interface {
	GetUserWalletByOwnerId(ctx context.Context, ownerId uuid.UUID) (*Account, *response.Response[any])
	CreateUserWallet(ctx context.Context, ownerId uuid.UUID) (*Account, *response.Response[any])
	GetById(ctx context.Context, id uuid.UUID) (*Account, *response.Response[any])
	GetBalances(ctx context.Context, accountId uuid.UUID) ([]AccountBalance, *response.Response[any])
}
