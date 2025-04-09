import LivestreamsPreviewView from "../../components/main_page/LivesteamsPreviewView";
import { PopularVODView } from "../../components/main_page/popular-vod-view";

export default function HomePage() {
  return (
    <div className="flex flex-col w-full max-h-full px-8 py-4 overflow-y-scroll overflow-x-hidden text-xs">
      <LivestreamsPreviewView />

      <h1 className="font-semibold text-xl my-2">Popular VODs</h1>
      <PopularVODView />
    </div>
  );
}
