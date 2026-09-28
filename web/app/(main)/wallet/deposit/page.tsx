"use client";

import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { toast } from "@/components/utils/toast";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import useT from "@/hooks/use-translation";
import useUser from "@/hooks/user";
import { CreateDeposit } from "@/lib/api/wallet";
import { unwrapResponse } from "@/lib/api/api-error";
import { CurrencyCode, PaymentProvider } from "@/types/wallet";
import CurrencySelect from "../_components/currency-select";
import ProviderSelect from "../_components/provider-select";
import IconLoader from "@/components/icons/loader";

const MIN_DEPOSIT = 1;
const MAX_DEPOSIT = 10000;

export default function DepositPage() {
    const router = useRouter();
    const { t } = useT(["wallet", "api-response", "fetch-error"]);
    const user = useUser((s) => s.user);

    const [currency, setCurrency] = useState<CurrencyCode>(CurrencyCode.SPARK);
    const [provider, setProvider] = useState<PaymentProvider>(
        PaymentProvider.STRIPE,
    );
    const [amount, setAmount] = useState("");
    const [error, setError] = useState<string | null>(null);

    const validateAmount = (value: string): string | null => {
        const num = parseFloat(value);
        if (!value || isNaN(num) || num <= 0) {
            return t("wallet:deposit.error_invalid_amount");
        }
        if (num < MIN_DEPOSIT) {
            return t("wallet:deposit.error_min_amount", {
                amount: MIN_DEPOSIT,
            });
        }
        if (num > MAX_DEPOSIT) {
            return t("wallet:deposit.error_max_amount", {
                amount: MAX_DEPOSIT,
            });
        }
        return null;
    };

    const depositMutation = useMutation({
        mutationFn: async () =>
            unwrapResponse(
                await CreateDeposit({
                    provider,
                    currencyCode: currency,
                    amount,
                }),
            ),
        onSuccess: (data) => {
            toast.success(t("wallet:deposit.success"));
            if (data?.checkoutUrl) {
                window.open(data.checkoutUrl, "_blank");
            } else {
                router.push("./overview");
            }
        },
    });

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (!user) return;

        const validationError = validateAmount(amount);
        if (validationError) {
            setError(validationError);
            return;
        }
        setError(null);
        depositMutation.mutate();
    };

    const parsedAmount = parseFloat(amount);
    const isValidPreview = !isNaN(parsedAmount) && parsedAmount > 0;

    return (
        <form onSubmit={handleSubmit} className="space-y-8">
            <section>
                <h2 className="text-foreground mb-1 text-xl font-semibold">
                    {t("wallet:deposit.title")}
                </h2>
                <p className="text-muted-foreground mb-6 text-sm">
                    {t("wallet:deposit.description")}
                </p>

                <div className="space-y-6">
                    {/* Currency selection */}
                    <div>
                        <label className="text-foreground mb-2 block text-sm font-medium">
                            {t("wallet:deposit.select_currency")}
                        </label>
                        <CurrencySelect
                            selected={currency}
                            onChange={setCurrency}
                        />
                    </div>

                    {/* Amount */}
                    <div>
                        <label className="text-foreground mb-2 block text-sm font-medium">
                            {t("wallet:deposit.amount")}
                        </label>
                        <Input
                            type="number"
                            min={MIN_DEPOSIT}
                            max={MAX_DEPOSIT}
                            step="0.01"
                            placeholder={t("wallet:deposit.amount_placeholder")}
                            value={amount}
                            onChange={(e) => {
                                setAmount(e.target.value);
                                setError(null);
                            }}
                            className="border-border"
                        />
                        {error && (
                            <p className="text-destructive mt-1 text-xs">
                                {error}
                            </p>
                        )}
                        <p className="text-muted-foreground mt-1 text-xs">
                            {t("wallet:deposit.min_amount", {
                                amount: MIN_DEPOSIT,
                            })}{" "}
                            &middot;{" "}
                            {t("wallet:deposit.max_amount", {
                                amount: MAX_DEPOSIT.toLocaleString(),
                            })}
                        </p>
                    </div>

                    {/* Provider selection */}
                    <div>
                        <label className="text-foreground mb-2 block text-sm font-medium">
                            {t("wallet:deposit.select_provider")}
                        </label>
                        <ProviderSelect
                            selected={provider}
                            onChange={setProvider}
                        />
                    </div>
                </div>
            </section>

            {/* Summary */}
            {isValidPreview && (
                <section className="border-border rounded-lg border p-4">
                    <h3 className="text-foreground mb-3 text-sm font-semibold">
                        {t("wallet:deposit.summary")}
                    </h3>
                    <div className="text-muted-foreground space-y-1 text-sm">
                        <div className="flex justify-between">
                            <span>{t("wallet:deposit.you_receive")}</span>
                            <span className="text-foreground font-medium">
                                {parsedAmount.toLocaleString()} {currency}
                            </span>
                        </div>
                    </div>
                </section>
            )}

            <div className="flex justify-end">
                <Button
                    type="submit"
                    disabled={depositMutation.isPending || !amount}
                >
                    {depositMutation.isPending && <IconLoader />}
                    {depositMutation.isPending
                        ? t("wallet:deposit.confirming")
                        : t("wallet:deposit.confirm")}
                </Button>
            </div>
        </form>
    );
}
