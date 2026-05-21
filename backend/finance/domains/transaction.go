package domains

import (
	"context"
	response "sen1or/letslive/finance/response"
	"time"

	"github.com/gofrs/uuid/v5"
)

type TransactionType string

const (
	TransactionTypeReward     TransactionType = "reward"
	TransactionTypePurchase   TransactionType = "purchase"
	TransactionTypeTrade      TransactionType = "trade"
	TransactionTypeDonate     TransactionType = "donate"
	TransactionTypeRefund     TransactionType = "refund"
	TransactionTypeFee        TransactionType = "fee"
	TransactionTypeAdjustment TransactionType = "adjustment"
)

type ProcessStatus string

const (
	ProcessStatusCreated    ProcessStatus = "created"
	ProcessStatusProcessing ProcessStatus = "processing"
	ProcessStatusCompleted  ProcessStatus = "completed"
	ProcessStatusFailed     ProcessStatus = "failed"
	ProcessStatusCancelled  ProcessStatus = "cancelled"
)

type Transaction struct {
	Id        uuid.UUID       `json:"id" db:"id"`
	Type      TransactionType `json:"type" db:"type"`
	Reference *string         `json:"reference" db:"reference"`
	Status    ProcessStatus   `json:"status" db:"status"`
	ActorId   *uuid.UUID      `json:"actorId" db:"actor_id"`
	Metadata  *string         `json:"metadata" db:"metadata"`
	CreatedAt time.Time       `json:"createdAt" db:"created_at"`
}

type LedgerEntry struct {
	Id            uuid.UUID `json:"id" db:"id"`
	TransactionId uuid.UUID `json:"transactionId" db:"transaction_id"`
	AccountId     uuid.UUID `json:"accountId" db:"account_id"`
	CurrencyCode  string    `json:"currencyCode" db:"currency_code"`
	Amount        int64     `json:"-" db:"amount"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
}

// LedgerEntryDraft is an unsaved ledger entry passed to repository transactional writes.
// Id and CreatedAt are populated by the database.
type LedgerEntryDraft struct {
	AccountId    uuid.UUID
	CurrencyCode string
	Amount       int64
}

type TransactionRepository interface {
	Create(ctx context.Context, tx Transaction) (*Transaction, *response.Response[any])
	GetById(ctx context.Context, id uuid.UUID) (*Transaction, *response.Response[any])
	GetEntriesForAccount(ctx context.Context, transactionId uuid.UUID, accountId uuid.UUID) ([]LedgerEntry, *response.Response[any])
	ListByActor(ctx context.Context, actorId uuid.UUID, page int, limit int) ([]Transaction, int, *response.Response[any])
	// CompleteWithEntries inserts ledger entries and transitions transaction status -> completed atomically.
	// The DB zero-sum trigger enforces sum(entries.amount) = 0 on the status transition.
	CompleteWithEntries(ctx context.Context, transactionId uuid.UUID, entries []LedgerEntryDraft) *response.Response[any]
}
