import Image from "next/image";
import { Button } from "@/components/ui/button";
import { myGetT } from "@/lib/i18n";
import Link from "next/link";
import { cookies } from "next/headers";
import { I18N_COOKIE_NAME, I18N_FALLBACK_LNG } from "@/lib/i18n/settings";

export default async function NotFound() {
    const cookieStore = await cookies();
    const lng = cookieStore.get(I18N_COOKIE_NAME)?.value ?? I18N_FALLBACK_LNG;
    const { t } = await myGetT(lng, ["common", "accessibility"]);

    return (
        <div className="flex min-h-screen w-full flex-col items-center justify-center p-4">
            <div className="flex max-w-[600px] items-center justify-center text-center">
                <Image
                    src="/images/pc-error.webp"
                    alt={t("accessibility:error_404_illustration")}
                    width={400}
                    height={300}
                    className="mb-6 h-auto w-full"
                    priority
                />
                <div>
                    <h1 className="mb-3 text-4xl font-bold">
                        {t("common:oops")}
                    </h1>
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
