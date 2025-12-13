import IconLoader from "@/components/icons/loader";
import { myGetT } from "@/lib/i18n";

export default async function LoadingPage() {
    const { t } = await myGetT("common");

    return (
        <div className="flex min-h-screen w-full flex-col items-center justify-center bg-gradient-to-b from-background to-muted/20 p-4">
            <div className="max-w-[500px] space-y-6 px-4 text-center">
                <div className="flex items-center justify-center">
                    <IconLoader className="h-12 w-12 animate-spin text-primary" />
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
                    <div className="h-2.5 w-full animate-pulse rounded-full bg-muted"></div>
                    <div className="mx-auto h-2.5 w-3/4 animate-pulse rounded-full bg-muted"></div>
                    <div className="mx-auto h-2.5 animate-pulse rounded-full bg-muted"></div>
                </div>
            </div>
        </div>
    );
}
