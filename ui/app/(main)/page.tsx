import LivestreamsPreviewView from "../../components/main_page/LivesteamsPreviewView";
import { PopularVODView } from "../../components/main_page/popular-vod-view";

export default function HomePage() {
    return (
        <div className="flex flex-col w-full max-h-full px-8 py-4 overflow-y-scroll overflow-x-hidden text-xs">
            <h1 className="font-semibold text-xl">Livestreaming</h1>
            <p>How to start your livestream: </p>
            <p>Open OBS &rarr; Settings &rarr; Stream </p>
            <p>
                Enter: &quot;Server: rtmp://{process.env.NEXT_PUBLIC_ENVIRONMENT === "production" ? "sen1or-huly.com" : "localhost"}:1935, StreamKey:
                Your key in Security Setting&quot;
            </p>
            <p className="mb-2">Start livestream</p>
            <LivestreamsPreviewView />

            <h1 className="font-semibold text-xl my-2">Popular VODs</h1>
            <PopularVODView />
        </div>
    );
}