import { ImageResponse } from "next/og";
import { GetUserById } from "@/lib/api/user";
import { GetLivestreamOfUser } from "@/lib/api/livestream";

export const runtime = "nodejs";
export const alt = "User profile";
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

export default async function Image({
    params,
}: {
    params: Promise<{ lng: string; userId: string }>;
}) {
    const { userId } = await params;
    let username = "User";
    let bio = "";
    let live = false;
    try {
        const [u, l] = await Promise.all([
            GetUserById(userId),
            GetLivestreamOfUser(userId).catch(() => null),
        ]);
        if (u.success && u.data) {
            username = u.data.username || username;
            bio = u.data.bio || "";
        }
        if (l && l.success && l.data && !l.data.endedAt) live = true;
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
                    alignItems: "flex-start",
                    background:
                        "linear-gradient(135deg, #0f172a 0%, #7c3aed 100%)",
                    color: "white",
                    padding: 80,
                    fontFamily: "sans-serif",
                }}
            >
                {live && (
                    <div
                        style={{
                            background: "#ef4444",
                            color: "white",
                            padding: "8px 20px",
                            borderRadius: 999,
                            fontSize: 28,
                            fontWeight: 700,
                            marginBottom: 24,
                        }}
                    >
                        LIVE
                    </div>
                )}
                <div style={{ fontSize: 96, fontWeight: 800, letterSpacing: -2 }}>
                    {username}
                </div>
                {bio && (
                    <div
                        style={{
                            marginTop: 24,
                            fontSize: 32,
                            opacity: 0.85,
                            maxWidth: 1000,
                        }}
                    >
                        {bio.slice(0, 180)}
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
