"use client";

import { cn } from "@/utils/cn";
import { PaymentProvider } from "@/types/wallet";
import useT from "@/hooks/use-translation";

interface Props {
    selected: PaymentProvider;
    onChange: (provider: PaymentProvider) => void;
}

export default function ProviderSelect({ selected, onChange }: Props) {
    const { t } = useT("wallet");

    const options = [
        {
            provider: PaymentProvider.STRIPE,
            name: t("wallet:deposit.provider.stripe"),
            description: t("wallet:deposit.provider.stripe_description"),
        },
        {
            provider: PaymentProvider.PAYPAL,
            name: t("wallet:deposit.provider.paypal"),
            description: t("wallet:deposit.provider.paypal_description"),
        },
    ];

    return (
        <div className="space-y-2">
            {options.map((opt) => (
                <button
                    key={opt.provider}
                    type="button"
                    onClick={() => onChange(opt.provider)}
                    className={cn(
                        "border-border flex w-full items-center gap-3 rounded-lg border p-4 text-left transition-colors",
                        selected === opt.provider
                            ? "border-primary bg-primary/5 ring-primary ring-1"
                            : "hover:bg-background-hover",
                    )}
                >
                    <div>
                        <p className="text-foreground text-sm font-medium">
                            {opt.name}
                        </p>
                        <p className="text-muted-foreground text-xs">
                            {opt.description}
                        </p>
                    </div>
                </button>
            ))}
        </div>
    );
}
