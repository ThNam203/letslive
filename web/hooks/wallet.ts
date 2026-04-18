import { create } from "zustand";
import {
    AccountBalance,
    CurrencyCode,
    Transaction,
    WalletOverview,
} from "../types/wallet";

export type WalletState = {
    wallet: WalletOverview | null;
    transactions: Transaction[];
    isLoading: boolean;
    transactionsPage: number;
    transactionsTotal: number;

    setWallet: (wallet: WalletOverview | null) => void;
    setBalances: (balances: AccountBalance[]) => void;
    setTransactions: (txns: Transaction[]) => void;
    appendTransactions: (txns: Transaction[]) => void;
    setIsLoading: (loading: boolean) => void;
    setTransactionsPage: (page: number) => void;
    setTransactionsTotal: (total: number) => void;

    getBalance: (currency: CurrencyCode) => string;
};

const useWallet = create<WalletState>((set, get) => ({
    wallet: null,
    transactions: [],
    isLoading: false,
    transactionsPage: 0,
    transactionsTotal: 0,

    setWallet: (wallet) => set({ wallet }),
    setBalances: (balances) =>
        set((state) => ({
            wallet: state.wallet ? { ...state.wallet, balances } : null,
        })),
    setTransactions: (transactions) => set({ transactions }),
    appendTransactions: (txns) =>
        set((state) => ({
            transactions: [...state.transactions, ...txns],
        })),
    setIsLoading: (isLoading) => set({ isLoading }),
    setTransactionsPage: (transactionsPage) => set({ transactionsPage }),
    setTransactionsTotal: (transactionsTotal) => set({ transactionsTotal }),

    getBalance: (currency) => {
        const balances = get().wallet?.balances ?? [];
        const found = balances.find((b) => b.currencyCode === currency);
        return found?.balance ?? "0";
    },
}));

export default useWallet;
