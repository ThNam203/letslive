import { myGetT } from "@/lib/i18n";
import { formatLocaleDate } from "@/utils/timeFormats";

const POLICY_LAST_UPDATED = "2026-08-19";

export async function generateMetadata() {
    const { t } = await myGetT("legal");
    return {
        title: t("legal:cookie_policy_title"),
    };
}

export default async function CookiePolicyPage() {
    const { t, i18n } = await myGetT("legal");
    const lastUpdated = formatLocaleDate(
        new Date(POLICY_LAST_UPDATED),
        i18n.resolvedLanguage,
        { dateStyle: "long" },
    );

    const rows = [
        {
            name: "ACCESS_TOKEN",
            purpose: t("legal:cookie_policy_row_access_purpose"),
            duration: t("legal:cookie_policy_row_access_duration"),
        },
        {
            name: "REFRESH_TOKEN",
            purpose: t("legal:cookie_policy_row_refresh_purpose"),
            duration: t("legal:cookie_policy_row_refresh_duration"),
        },
        {
            name: "lng",
            purpose: t("legal:cookie_policy_row_lng_purpose"),
            duration: t("legal:cookie_policy_row_lng_duration"),
        },
    ];

    return (
        <div className="mx-auto w-full max-w-3xl px-4 py-10">
            <h1 className="mb-2 text-3xl font-bold">
                {t("legal:cookie_policy_title")}
            </h1>
            <p className="text-muted-foreground mb-8 text-sm">
                {t("legal:cookie_policy_updated", { date: lastUpdated })}
            </p>

            <section className="mb-8">
                <h2 className="mb-2 text-xl font-semibold">
                    {t("legal:cookie_policy_intro_title")}
                </h2>
                <p className="text-muted-foreground">
                    {t("legal:cookie_policy_intro_body")}
                </p>
            </section>

            <section className="mb-8">
                <h2 className="mb-2 text-xl font-semibold">
                    {t("legal:cookie_policy_table_title")}
                </h2>
                <div className="border-border overflow-x-auto rounded-lg border">
                    <table className="w-full text-left text-sm">
                        <thead className="bg-muted">
                            <tr>
                                <th className="p-3 font-medium">
                                    {t("legal:cookie_policy_table_name")}
                                </th>
                                <th className="p-3 font-medium">
                                    {t("legal:cookie_policy_table_purpose")}
                                </th>
                                <th className="p-3 font-medium">
                                    {t("legal:cookie_policy_table_duration")}
                                </th>
                                <th className="p-3 font-medium">
                                    {t("legal:cookie_policy_table_type")}
                                </th>
                            </tr>
                        </thead>
                        <tbody>
                            {rows.map((row) => (
                                <tr
                                    key={row.name}
                                    className="border-border border-t"
                                >
                                    <td className="p-3 font-mono text-xs">
                                        {row.name}
                                    </td>
                                    <td className="p-3">{row.purpose}</td>
                                    <td className="p-3">{row.duration}</td>
                                    <td className="p-3">
                                        {t("legal:cookie_policy_type_essential")}
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
                <p className="text-muted-foreground mt-3 text-sm">
                    {t("legal:cookie_policy_essential_note")}
                </p>
            </section>

            <section className="mb-8">
                <h2 className="mb-2 text-xl font-semibold">
                    {t("legal:cookie_policy_control_title")}
                </h2>
                <p className="text-muted-foreground">
                    {t("legal:cookie_policy_control_body")}
                </p>
            </section>

            <section>
                <h2 className="mb-2 text-xl font-semibold">
                    {t("legal:cookie_policy_contact_title")}
                </h2>
                <p className="text-muted-foreground">
                    {t("legal:cookie_policy_contact_body")}
                </p>
            </section>
        </div>
    );
}
