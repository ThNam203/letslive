"use client";

import { cn } from "@/utils/cn";
import { CurrencyCode } from "@/types/wallet";
import useT from "@/hooks/use-translation";

interface Props {
    currencyCode: CurrencyCode;
    balance: string;
    className?: string;
}

const currencyConfig: Record<CurrencyCode, { gradient: string; icon: string }> =
    {
        [CurrencyCode.SPARK]: {
            gradient: "from-amber-500 to-orange-600",
            icon: "&#9889;",
        },
        [CurrencyCode.FLARE]: {
            gradient: "from-purple-500 to-indigo-600",
            icon: "&#128142;",
        },
    };

export default function BalanceCard({
    currencyCode,
    balance,
    className,
}: Props) {
    const { t } = useT("wallet");
    const config = currencyConfig[currencyCode];

    const nameKey =
        currencyCode === CurrencyCode.SPARK
            ? "wallet:currency.spark"
            : "wallet:currency.flare";
    const descKey =
        currencyCode === CurrencyCode.SPARK
            ? "wallet:currency.spark_description"
            : "wallet:currency.flare_description";

    return (
        <div
            className={cn(
                "relative overflow-hidden rounded-xl p-6 text-white",
                "bg-gradient-to-br",
                config.gradient,
                className,
            )}
        >
            <div className="absolute top-3 right-4 text-3xl opacity-30">
                <span dangerouslySetInnerHTML={{ __html: config.icon }} />
            </div>
            <p className="text-sm font-medium opacity-80">{t(nameKey)}</p>
            <p className="mt-1 text-3xl font-bold tracking-tight">
                {formatBalance(balance)}
            </p>
            <p className="mt-2 text-xs opacity-60">{t(descKey)}</p>
        </div>
    );
}

function formatBalance(value: string): string {
    const num = parseFloat(value);
    if (isNaN(num)) return "0";
    return num.toLocaleString(undefined, {
        minimumFractionDigits: 0,
        maximumFractionDigits: 2,
    });
}
