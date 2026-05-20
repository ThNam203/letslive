import { ImageResponse } from "next/og";
import { GetVODInformation } from "@/lib/api/vod";
import { GetUserById } from "@/lib/api/user";

export const runtime = "nodejs";
export const alt = "Video";
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

export default async function Image({
    params,
}: {
    params: Promise<{ lng: string; userId: string; vodId: string }>;
}) {
    const { userId, vodId } = await params;
    let title = "Video";
    let username = "";
    try {
        const [v, u] = await Promise.all([
            GetVODInformation(vodId),
            GetUserById(userId).catch(() => null),
        ]);
        if (v.success && v.data) title = v.data.title || title;
        if (u && u.success && u.data) username = u.data.username || "";
    } catch {}

    return new ImageResponse(
        (
            <div
                style={{
                    width: "100%",
                    height: "100%",
                    display: "flex",
                    flexDirection: "column",
                    justifyContent: "center",
                    background:
                        "linear-gradient(135deg, #1e1b4b 0%, #db2777 100%)",
                    color: "white",
                    padding: 80,
                    fontFamily: "sans-serif",
                }}
            >
                <div
                    style={{
                        fontSize: 28,
                        opacity: 0.7,
                        marginBottom: 16,
                        textTransform: "uppercase",
                        letterSpacing: 4,
                    }}
                >
                    Video
                </div>
                <div
                    style={{
                        fontSize: 72,
                        fontWeight: 800,
                        lineHeight: 1.1,
                        letterSpacing: -1,
                        maxWidth: 1040,
                    }}
                >
                    {title.slice(0, 140)}
                </div>
                {username && (
                    <div
                        style={{
                            marginTop: 32,
                            fontSize: 32,
                            opacity: 0.9,
                        }}
                    >
                        by {username}
                    </div>
                )}
                <div
                    style={{
                        marginTop: "auto",
                        fontSize: 24,
                        opacity: 0.6,
                    }}
                >
                    Let&apos;s Live
                </div>
            </div>
        ),
        size,
    );
}
