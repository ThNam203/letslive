import LivestreamsPreviewView from "@/components/livestream/livesteams-preview";
import { PopularVODView } from "@/components/livestream/popular-vod-view";
import { myGetT } from "@/lib/i18n";

export default async function HomePage() {
  const { t } = await myGetT("translation");
  return (
    <div className="flex flex-col w-full max-h-full px-8 py-4 overflow-y-auto overflow-x-hidden text-xs">
      <h1 className="font-semibold text-xl my-2">{t("home.livestreams")}</h1>
      <LivestreamsPreviewView />

      <h1 className="font-semibold text-xl my-2">{t("home.videos")}</h1>
      <PopularVODView />
    </div>
  );
}
