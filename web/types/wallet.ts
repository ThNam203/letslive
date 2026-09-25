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
    /** null for platform, escrow and fee accounts — only user wallets have an owner */
    ownerId: string | null;
    type: AccountType;
    status: AccountStatus;
    createdAt: string;
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
    DEPOSIT = "deposit",
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
    /** null for transactions initiated by a service rather than a user */
    actorId: string | null;
    metadata: Record<string, unknown> | null;
    createdAt: string;
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
    CREATED = "created",
    PROCESSING = "processing",
    COMPLETED = "completed",
    FAILED = "failed",
    CANCELLED = "cancelled",
}

export enum PaymentProvider {
    STRIPE = "stripe",
    /** registered on the dev profile only */
    MOCK = "mock",
}

export type Payment = {
    id: string;
    transactionId: string;
    provider: PaymentProvider;
    providerReference: string;
    currencyCode: CurrencyCode;
    amount: string;
    status: PaymentStatus;
    createdAt: string;
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
