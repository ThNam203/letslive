import LivestreamsPreviewView from "@/components/livestream/livesteams-preview";
import { PopularVODView } from "@/components/livestream/popular-vod-view";
import { myGetT } from "@/lib/i18n";
import { I18N_FALLBACK_LNG, I18N_LANGUAGES } from "@/lib/i18n/settings";
import type { Metadata } from "next";

function sleep(ms: number) {
    return new Promise((resolve) => setTimeout(resolve, ms));
}

export async function generateMetadata({
    params,
}: {
    params: Promise<{ lng: string }>;
}): Promise<Metadata> {
    const { lng } = await params;
    const { t } = await myGetT("common");
    const title = t("meta_home_title");
    const description = t("app_description");
    const languages = Object.fromEntries(
        I18N_LANGUAGES.map((l) => [l, `/${l}`]),
    );
    languages["x-default"] = `/${I18N_FALLBACK_LNG}`;
    return {
        title,
        description,
        alternates: { canonical: `/${lng}`, languages },
        openGraph: { title, description, url: `/${lng}` },
    };
}

export default async function HomePage() {
    const { t } = await myGetT("common");

    return (
        <div className="flex max-h-full w-full flex-col overflow-x-hidden overflow-y-auto px-8 py-4 text-xs">
            <h1 className="my-2 text-xl font-semibold">{t("livestreams")}</h1>
            <LivestreamsPreviewView />

            <h1 className="my-2 text-xl font-semibold">{t("videos")}</h1>
            <PopularVODView />
        </div>
    );
}
