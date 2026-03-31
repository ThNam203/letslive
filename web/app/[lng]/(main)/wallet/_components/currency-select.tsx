"use client";

import { cn } from "@/utils/cn";
import { CurrencyCode } from "@/types/wallet";
import useT from "@/hooks/use-translation";

interface Props {
    selected: CurrencyCode;
    onChange: (code: CurrencyCode) => void;
}

export default function CurrencySelect({ selected, onChange }: Props) {
    const { t } = useT("wallet");

    const options = [
        {
            code: CurrencyCode.SPARK,
            name: t("wallet:currency.spark"),
            description: t("wallet:currency.spark_description"),
            icon: "&#9889;",
        },
        {
            code: CurrencyCode.FLARE,
            name: t("wallet:currency.flare"),
            description: t("wallet:currency.flare_description"),
            icon: "&#128142;",
        },
    ];

    return (
        <div className="grid grid-cols-2 gap-3">
            {options.map((opt) => (
                <button
                    key={opt.code}
                    type="button"
                    onClick={() => onChange(opt.code)}
                    className={cn(
                        "border-border flex items-center gap-3 rounded-lg border p-4 text-left transition-colors",
                        selected === opt.code
                            ? "border-primary bg-primary/5 ring-primary ring-1"
                            : "hover:bg-background-hover",
                    )}
                >
                    <span
                        className="text-2xl"
                        dangerouslySetInnerHTML={{ __html: opt.icon }}
                    />
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
