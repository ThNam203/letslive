import type { Metadata } from "next";
import { GetVODInformation } from "@/lib/api/vod";
import { GetUserById } from "@/lib/api/user";
import { myGetT } from "@/lib/i18n";
import { getSiteUrl, toAbsoluteUrl } from "@/utils/siteUrl";
import GLOBAL from "@/global";
import VODView from "./vod-view";

type VODPageParams = { userId: string; vodId: string };

const OG_IMAGE_WIDTH = 1280;
const OG_IMAGE_HEIGHT = 720;
const OG_DESCRIPTION_MAX_LENGTH = 200;

function truncate(text: string, max: number): string {
    const trimmed = text.trim();
    return trimmed.length <= max ? trimmed : `${trimmed.slice(0, max - 1)}…`;
}

export async function generateMetadata({
    params,
}: {
    params: Promise<VODPageParams>;
}): Promise<Metadata> {
    const { userId, vodId } = await params;
    const { t } = await myGetT("common");
    const siteUrl = await getSiteUrl();
    const appTitle = t("common:app_title");
    const pageUrl = `${siteUrl}/users/${userId}/vods/${vodId}`;

    const [vod, user] = await Promise.all([
        GetVODInformation(vodId)
            .then((res) => res.data ?? null)
            .catch(() => null),
        GetUserById(userId)
            .then((res) => res.data ?? null)
            .catch(() => null),
    ]);

    // a private or missing VOD must not leak its title into a chat preview
    if (!vod || vod.visibility !== "public") {
        return {
            // absolute: nothing to name, so skip the "<page> | <app>" template
            title: { absolute: appTitle },
            robots: { index: false, follow: false },
            openGraph: { title: appTitle, url: pageUrl, siteName: appTitle },
        };
    }

    const vodTitle = vod.title?.trim() ?? "";
    const title = vodTitle || appTitle;
    const author = user?.username?.trim() ?? "";
    const description = vod.description?.trim()
        ? truncate(vod.description, OG_DESCRIPTION_MAX_LENGTH)
        : author
          ? t("common:vod_share_description", { username: author })
          : appTitle;

    const imageUrl =
        toAbsoluteUrl(vod.thumbnailUrl, siteUrl) ??
        `${GLOBAL.API_URL}/files/livestreams/${vod.id}/thumbnail.jpeg`;

    return {
        // an untitled VOD has nothing to prefix the app name with
        title: vodTitle || { absolute: appTitle },
        description,
        alternates: { canonical: pageUrl },
        openGraph: {
            type: "video.other",
            url: pageUrl,
            siteName: appTitle,
            title,
            description,
            images: [
                {
                    url: imageUrl,
                    width: OG_IMAGE_WIDTH,
                    height: OG_IMAGE_HEIGHT,
                    alt: title,
                },
            ],
        },
        twitter: {
            card: "summary_large_image",
            title,
            description,
            images: [imageUrl],
        },
    };
}

export default function VODPage() {
    return <VODView />;
}
