"use client";

import Image from "next/image";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import useT from "@/hooks/use-translation";

export default function GlobalError() {
    const { t } = useT(["error", "common"]);

    return (
        <div className="min-h-screen w-full flex flex-col items-center justify-center p-4">
            <div className="max-w-[600px] text-center flex items-center justify-center">
                <Image
                    src="/images/pc-error.png"
                    alt="500 Error Illustration"
                    width={400}
                    height={300}
                    className="w-full h-auto mb-6"
                    priority
                />
                <div>
                    <h1 className="text-4xl font-bold mb-3">{t("general_title")}</h1>
                    <p className="text-muted-foreground text-lg mb-6">{t("general_description")}</p>
                    <Button asChild>
                        <Link href="/">{t("common:go_home")}</Link>
                    </Button>
                </div>
            </div>
        </div>
    );
}