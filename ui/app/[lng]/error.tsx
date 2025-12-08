"use client";

import Image from "next/image";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import useT from "@/hooks/use-translation";

export default function GlobalError() {
    const { t } = useT(["error", "common"]);

    return (
        <div className="flex min-h-screen w-full flex-col items-center justify-center p-4">
            <div className="flex max-w-[600px] items-center justify-center text-center">
                <Image
                    src="/images/pc-error.png"
                    alt="500 Error Illustration"
                    width={400}
                    height={300}
                    className="mb-6 h-auto w-full"
                    priority
                />
                <div>
                    <h1 className="mb-3 text-4xl font-bold">
                        {t("error:general_title")}
                    </h1>
                    <p className="text-muted-foreground mb-6 text-lg">
                        {t("error:general_description")}
                    </p>
                    <Button asChild>
                        <Link href="/">{t("common:go_home")}</Link>
                    </Button>
                </div>
            </div>
        </div>
    );
}
