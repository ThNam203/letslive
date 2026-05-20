import type { Metadata } from "next";
import { GetUserById } from "@/lib/api/user";
import { GetLivestreamOfUser } from "@/lib/api/livestream";
import { myGetT } from "@/lib/i18n";
import { I18N_FALLBACK_LNG, I18N_LANGUAGES } from "@/lib/i18n/settings";

type Params = Promise<{ lng: string; userId: string }>;

export async function generateMetadata({
    params,
}: {
    params: Params;
}): Promise<Metadata> {
    const { lng, userId } = await params;
    const { t } = await myGetT("common");

    let username = "";
    let bio = "";
    let avatar: string | undefined;
    let liveTitle: string | null = null;

    try {
        const [userRes, liveRes] = await Promise.all([
            GetUserById(userId),
            GetLivestreamOfUser(userId).catch(() => null),
        ]);
        if (userRes.success && userRes.data) {
            username = userRes.data.username || "";
            bio = userRes.data.bio || "";
            avatar = userRes.data.profilePicture;
        }
        if (liveRes && liveRes.success && liveRes.data) {
            liveTitle = liveRes.data.title || null;
        }
    } catch (err) {
        console.error("user metadata fetch failed", err);
    }

    if (!username) {
        return { robots: { index: false, follow: false } };
    }

    const title = liveTitle
        ? t("meta_user_live", { username, title: liveTitle })
        : t("meta_user_offline", { username });
    const description = t("meta_user_description", { username, bio });
    const path = `/users/${userId}`;
    const languages = Object.fromEntries(
        I18N_LANGUAGES.map((l) => [l, `/${l}${path}`]),
    );
    languages["x-default"] = `/${I18N_FALLBACK_LNG}${path}`;

    return {
        title,
        description,
        alternates: { canonical: `/${lng}${path}`, languages },
        openGraph: {
            type: "profile",
            title,
            description,
            url: `/${lng}${path}`,
            images: avatar ? [{ url: avatar }] : undefined,
        },
        twitter: {
            card: "summary_large_image",
            title,
            description,
            images: avatar ? [avatar] : undefined,
        },
    };
}

export default function UserLayout({ children }: { children: React.ReactNode }) {
    return children;
}
