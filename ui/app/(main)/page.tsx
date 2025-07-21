import LivestreamsPreviewView from "../../components/livestream/livesteams-preview";
import { PopularVODView } from "../../components/livestream/popular-vod-view";

export default function HomePage() {
  return (
    <div className="flex flex-col w-full max-h-full px-8 py-4 overflow-y-auto overflow-x-hidden text-xs">
      <LivestreamsPreviewView />

      <h1 className="font-semibold text-xl my-2">Popular VODs</h1>
      <PopularVODView />
    </div>
  );
}
