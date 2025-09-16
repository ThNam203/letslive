import IconLoader from "@/components/icons/loader";
import { myGetT } from "@/lib/i18n";

export default async function LoadingPage() {
  const { t } = await myGetT("common")

  return (
    <div className="min-h-screen w-full flex flex-col items-center justify-center p-4 bg-gradient-to-b from-background to-muted/20">
      <div className="max-w-[500px] text-center space-y-6 px-4">
        <div className="flex items-center justify-center">
          <IconLoader className="h-12 w-12 text-primary animate-spin" />
        </div>
        <div className="space-y-2">
          <h1 className="text-3xl font-bold tracking-tight">{t("loading")}</h1>
          <p className="text-muted-foreground">{t("please_wait_while_loading")}</p>
        </div>
        <div className="space-y-2 pt-4">
          <div className="h-2.5 bg-muted rounded-full w-full animate-pulse"></div>
          <div className="h-2.5 bg-muted rounded-full w-3/4 mx-auto animate-pulse"></div>
          <div className="h-2.5 bg-muted rounded-full mx-auto animate-pulse"></div>
        </div>
      </div>
    </div>
  )
}