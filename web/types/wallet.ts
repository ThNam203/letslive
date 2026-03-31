// ---------------------------------------------------------------------------
// Currency
// ---------------------------------------------------------------------------

export enum CurrencyCode {
    SPARK = "SPARK",
    FLARE = "FLARE",
}

export type Currency = {
    code: CurrencyCode;
    name: string;
    precision: number;
};

// ---------------------------------------------------------------------------
// Account & Balance
// ---------------------------------------------------------------------------

export enum AccountType {
    USER_WALLET = "user_wallet",
    PLATFORM = "platform",
    ESCROW = "escrow",
    FEE = "fee",
}

export enum AccountStatus {
    ACTIVE = "active",
    FROZEN = "frozen",
    CLOSED = "closed",
}

export type Account = {
    id: string;
    ownerId: string;
    type: AccountType;
    status: AccountStatus;
    createdAt: string;
    updatedAt: string;
};

export type AccountBalance = {
    accountId: string;
    currencyCode: CurrencyCode;
    balance: string; // decimal string for precision
    lastEntryId: string | null;
};

export type WalletOverview = {
    account: Account;
    balances: AccountBalance[];
};

// ---------------------------------------------------------------------------
// Transaction
// ---------------------------------------------------------------------------

export enum TransactionType {
    REWARD = "reward",
    PURCHASE = "purchase",
    TRADE = "trade",
    DONATE = "donate",
    REFUND = "refund",
    FEE = "fee",
    ADJUSTMENT = "adjustment",
}

export enum TransactionStatus {
    CREATED = "created",
    PROCESSING = "processing",
    COMPLETED = "completed",
    FAILED = "failed",
    CANCELLED = "cancelled",
}

export type Transaction = {
    id: string;
    type: TransactionType;
    status: TransactionStatus;
    reference: string | null;
    description: string | null;
    actorId: string;
    metadata: Record<string, any> | null;
    createdAt: string;
    updatedAt: string;
    entries: LedgerEntry[] | null;
};

export type LedgerEntry = {
    id: string;
    transactionId: string;
    accountId: string;
    currencyCode: CurrencyCode;
    amount: string; // positive = credit, negative = debit
    createdAt: string;
};

// ---------------------------------------------------------------------------
// Payment (deposit / withdrawal)
// ---------------------------------------------------------------------------

export enum PaymentStatus {
    PENDING = "pending",
    PROCESSING = "processing",
    COMPLETED = "completed",
    FAILED = "failed",
    CANCELLED = "cancelled",
}

export enum PaymentProvider {
    STRIPE = "stripe",
    PAYPAL = "paypal",
}

export type Payment = {
    id: string;
    transactionId: string | null;
    provider: PaymentProvider;
    providerReference: string | null;
    currencyCode: CurrencyCode;
    amount: string;
    status: PaymentStatus;
    createdAt: string;
    updatedAt: string;
};

// ---------------------------------------------------------------------------
// Request / Response helpers
// ---------------------------------------------------------------------------

export type CreateDepositRequest = {
    provider: PaymentProvider;
    currencyCode: CurrencyCode;
    amount: string;
};

export type DepositResponse = {
    payment: Payment;
    checkoutUrl: string;
};
