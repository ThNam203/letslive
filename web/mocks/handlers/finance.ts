import { http } from "msw";
import { API_BASE, ok, notFound, created } from "../utils";
import {
    walletAccount,
    walletBalances,
    currencies,
    transactions,
    payments,
    ME_USER_ID,
    uid,
    now,
} from "../db";
import {
    Currency,
    WalletOverview,
    Transaction,
    Payment,
    DepositResponse,
    CreateDepositRequest,
    PaymentStatus,
    TransactionStatus,
    TransactionType,
    CurrencyCode,
    PaymentProvider,
} from "@/types/wallet";

export const financeHandlers = [
    http.get(`${API_BASE}/wallet`, () => {
        const overview: WalletOverview = {
            account: walletAccount,
            balances: walletBalances,
        };
        return ok<WalletOverview>(overview);
    }),

    http.get(`${API_BASE}/currencies`, () => {
        return ok<Currency[]>(currencies);
    }),

    http.get(`${API_BASE}/transactions`, ({ request }) => {
        const url = new URL(request.url);
        const page = parseInt(url.searchParams.get("page") ?? "0");
        const pageSize = parseInt(url.searchParams.get("page_size") ?? "20");
        const sorted = [...transactions].sort(
            (a, b) =>
                new Date(b.createdAt).getTime() -
                new Date(a.createdAt).getTime(),
        );
        const slice = sorted.slice(page * pageSize, page * pageSize + pageSize);
        return ok<Transaction[]>(slice, {
            page,
            page_size: pageSize,
            total: transactions.length,
        });
    }),

    http.get(
        `${API_BASE}/transactions/:transactionId`,
        ({ params }) => {
            const { transactionId } = params as { transactionId: string };
            const tx = transactions.find((t) => t.id === transactionId);
            if (!tx)
                return notFound(
                    "res_err_transaction_failed",
                    "Transaction not found",
                );
            return ok<Transaction>(tx);
        },
    ),

    http.post(`${API_BASE}/deposits`, async ({ request }) => {
        const body = (await request.json()) as CreateDepositRequest;

        const paymentId = `pay-${uid()}`;
        const txId = `tx-${uid()}`;

        const newPayment: Payment = {
            id: paymentId,
            transactionId: txId,
            provider: body.provider ?? PaymentProvider.STRIPE,
            providerReference: `pi_mock_${uid()}`,
            currencyCode: body.currencyCode ?? CurrencyCode.SPARK,
            amount: body.amount,
            status: PaymentStatus.CREATED,
            createdAt: now(),
            updatedAt: now(),
        };
        payments.push(newPayment);

        const newTx: Transaction = {
            id: txId,
            type: TransactionType.PURCHASE,
            status: TransactionStatus.PROCESSING,
            reference: paymentId,
            description: `Deposit via ${body.provider}`,
            actorId: ME_USER_ID,
            metadata: { provider: body.provider, amount: body.amount },
            createdAt: now(),
            updatedAt: now(),
            entries: null,
        };
        transactions.push(newTx);

        // Update balance optimistically (mock — real backend would confirm via webhook)
        const balance = walletBalances.find(
            (b) => b.currencyCode === body.currencyCode,
        );
        if (balance) {
            balance.balance = (
                parseFloat(balance.balance) + parseFloat(body.amount)
            ).toFixed(2);
        }

        const depositResponse: DepositResponse = {
            payment: newPayment,
            checkoutUrl: "",
        };
        return created<DepositResponse>(depositResponse);
    }),

    http.get(`${API_BASE}/payments`, ({ request }) => {
        const url = new URL(request.url);
        const page = parseInt(url.searchParams.get("page") ?? "0");
        const pageSize = parseInt(url.searchParams.get("page_size") ?? "20");
        const sorted = [...payments].sort(
            (a, b) =>
                new Date(b.createdAt).getTime() -
                new Date(a.createdAt).getTime(),
        );
        const slice = sorted.slice(page * pageSize, page * pageSize + pageSize);
        return ok<Payment[]>(slice, {
            page,
            page_size: pageSize,
            total: payments.length,
        });
    }),

    http.get(`${API_BASE}/payments/:paymentId`, ({ params }) => {
        const { paymentId } = params as { paymentId: string };
        const payment = payments.find((p) => p.id === paymentId);
        if (!payment)
            return notFound("res_err_payment_not_found", "Payment not found");
        return ok<Payment>(payment);
    }),
];
