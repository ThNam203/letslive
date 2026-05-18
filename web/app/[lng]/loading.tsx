import IconLoader from "@/components/icons/loader";
import { myGetT } from "@/lib/i18n";
import { cookies } from "next/headers";
import { I18N_COOKIE_NAME, I18N_FALLBACK_LNG } from "@/lib/i18n/settings";

export default async function LoadingPage() {
    const cookieStore = await cookies();
    const lng = cookieStore.get(I18N_COOKIE_NAME)?.value ?? I18N_FALLBACK_LNG;
    const { t } = await myGetT(lng, "common");

    return (
        <div className="from-background to-muted/20 flex min-h-screen w-full flex-col items-center justify-center bg-gradient-to-b p-4">
            <div className="max-w-[500px] space-y-6 px-4 text-center">
                <div className="flex items-center justify-center">
                    <IconLoader className="text-primary h-12 w-12 animate-spin" />
                </div>
                <div className="space-y-2">
                    <h1 className="text-3xl font-bold tracking-tight">
                        {t("loading")}
                    </h1>
                    <p className="text-muted-foreground">
                        {t("please_wait_while_loading")}
                    </p>
                </div>
                <div className="space-y-2 pt-4">
                    <div className="bg-muted h-2.5 w-full animate-pulse rounded-full"></div>
                    <div className="bg-muted mx-auto h-2.5 w-3/4 animate-pulse rounded-full"></div>
                    <div className="bg-muted mx-auto h-2.5 animate-pulse rounded-full"></div>
                </div>
            </div>
        </div>
    );
}
