import { ApiResponse } from "@/types/fetch-response";
import {
    CreateDepositRequest,
    Currency,
    DepositResponse,
    Payment,
    Transaction,
    WalletOverview,
} from "@/types/wallet";
import { fetchClient } from "@/utils/fetchClient";

// ---------------------------------------------------------------------------
// Wallet / Account
// ---------------------------------------------------------------------------

export async function GetMyWallet(): Promise<ApiResponse<WalletOverview>> {
    return fetchClient<ApiResponse<WalletOverview>>(`/finance/wallet`);
}

// ---------------------------------------------------------------------------
// Currencies
// ---------------------------------------------------------------------------

export async function GetCurrencies(): Promise<ApiResponse<Currency[]>> {
    return fetchClient<ApiResponse<Currency[]>>(`/finance/currencies`);
}

// ---------------------------------------------------------------------------
// Transactions
// ---------------------------------------------------------------------------

export async function GetTransactions(
    page: number = 0,
    pageSize: number = 20,
): Promise<ApiResponse<Transaction[]>> {
    return fetchClient<ApiResponse<Transaction[]>>(
        `/finance/transactions?page=${page}&page_size=${pageSize}`,
    );
}

export async function GetTransactionById(
    transactionId: string,
): Promise<ApiResponse<Transaction>> {
    return fetchClient<ApiResponse<Transaction>>(
        `/finance/transactions/${transactionId}`,
    );
}

// ---------------------------------------------------------------------------
// Deposits (payment gateway)
// ---------------------------------------------------------------------------

export async function CreateDeposit(
    data: CreateDepositRequest,
): Promise<ApiResponse<DepositResponse>> {
    return fetchClient<ApiResponse<DepositResponse>>(`/finance/deposits`, {
        method: "POST",
        body: JSON.stringify(data),
    });
}

export async function GetPayments(
    page: number = 0,
    pageSize: number = 20,
): Promise<ApiResponse<Payment[]>> {
    return fetchClient<ApiResponse<Payment[]>>(
        `/finance/payments?page=${page}&page_size=${pageSize}`,
    );
}

export async function GetPaymentById(
    paymentId: string,
): Promise<ApiResponse<Payment>> {
    return fetchClient<ApiResponse<Payment>>(`/finance/payments/${paymentId}`);
}
