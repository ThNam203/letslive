import LivestreamsPreviewView from "@/components/livestream/livesteams-preview";
import { PopularVODView } from "@/components/livestream/popular-vod-view";
import { myGetT } from "@/lib/i18n";
function sleep(ms: number) {
    return new Promise((resolve) => setTimeout(resolve, ms));
}

export default async function HomePage() {
    const { t } = await myGetT("common");

    return (
        <div className="flex max-h-full w-full flex-col overflow-y-auto overflow-x-hidden px-8 py-4 text-xs">
            <h1 className="my-2 text-xl font-semibold">{t("livestreams")}</h1>
            <LivestreamsPreviewView />

            <h1 className="my-2 text-xl font-semibold">{t("videos")}</h1>
            <PopularVODView />
        </div>
    );
}
