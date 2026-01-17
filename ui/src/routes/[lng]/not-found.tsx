"use client";

import Image from "next/image";
import { Button } from "@/components/ui/button";
import useT from "@/hooks/use-translation";
import Link from "next/link";

export default function NotFound() {
    const { t } = useT(["common"]);

    return (
        <div className="flex min-h-screen w-full flex-col items-center justify-center p-4">
            <div className="flex max-w-[600px] items-center justify-center text-center">
                <Image
                    src="/images/pc-error.png"
                    alt="404 Error Illustration"
                    width={400}
                    height={300}
                    className="mb-6 h-auto w-full"
                    priority
                />
                <div>
                    <h1 className="mb-3 text-4xl font-bold">Oops!</h1>
                    <p className="text-muted-foreground mb-6 text-lg">
                        {t("common:page_not_found")}
                    </p>
                    <Button asChild>
                        <Link href="/">{t("common:go_home")}</Link>
                    </Button>
                </div>
            </div>
        </div>
    );
}
