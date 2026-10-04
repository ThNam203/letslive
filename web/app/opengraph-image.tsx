import { ImageResponse } from "next/og";
import { myGetT } from "@/lib/i18n";

// the default link-preview card: Next attaches it to every route that does
// not set openGraph.images itself, so pages with nothing to illustrate still
// share as a branded card rather than a bare URL
export const alt = "Let's Live";
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

export default async function OpenGraphImage() {
    const { t } = await myGetT("common");

    return new ImageResponse(
        (
            <div
                style={{
                    width: "100%",
                    height: "100%",
                    display: "flex",
                    flexDirection: "column",
                    alignItems: "center",
                    justifyContent: "center",
                    gap: 24,
                    // --background and --primary from globals.css
                    background:
                        "linear-gradient(135deg, hsl(222, 47%, 11%) 0%, hsl(276, 91%, 24%) 100%)",
                    color: "#ffffff",
                    fontFamily: "sans-serif",
                }}
            >
                <div style={{ fontSize: 112, fontWeight: 700 }}>
                    {t("common:app_title")}
                </div>
                <div
                    style={{
                        fontSize: 40,
                        color: "hsl(0, 0%, 80%)",
                        maxWidth: 900,
                        textAlign: "center",
                    }}
                >
                    {t("common:app_description")}
                </div>
            </div>
        ),
        size,
    );
}
