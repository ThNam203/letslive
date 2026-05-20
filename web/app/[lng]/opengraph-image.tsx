import { ImageResponse } from "next/og";
import { myGetT } from "@/lib/i18n";

export const runtime = "nodejs";
export const alt = "Let's Live";
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

export default async function Image() {
    const { t } = await myGetT("common");
    const title = t("app_title");
    const description = t("app_description");

    return new ImageResponse(
        (
            <div
                style={{
                    width: "100%",
                    height: "100%",
                    display: "flex",
                    flexDirection: "column",
                    justifyContent: "center",
                    alignItems: "center",
                    background:
                        "linear-gradient(135deg, #0f172a 0%, #4338ca 100%)",
                    color: "white",
                    padding: 80,
                    fontFamily: "sans-serif",
                }}
            >
                <div style={{ fontSize: 96, fontWeight: 800, letterSpacing: -2 }}>
                    {title}
                </div>
                <div
                    style={{
                        marginTop: 24,
                        fontSize: 32,
                        textAlign: "center",
                        opacity: 0.85,
                        maxWidth: 1000,
                    }}
                >
                    {description}
                </div>
            </div>
        ),
        size,
    );
}
