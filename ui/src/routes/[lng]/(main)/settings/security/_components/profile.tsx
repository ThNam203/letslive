import { cn } from "@/utils/cn";
import { Button } from "@/components/ui/button";
import useT from "@/hooks/use-translation";
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog";
import { AuthProvider, User } from "@/types/user";
import Section from "@/routes/[lng]/(main)/settings/_components/section";
import ApiKeyTab from "@/routes/[lng]/(main)/settings/security/_components/api-key-tab";
import ChangePasswordTab from "@/routes/[lng]/(main)/settings/security/_components/change-password-tab";
import PhoneNumber from "@/routes/[lng]/(main)/settings/security/_components/phone-number";

export default function ContactSettings({ user }: { user: User }) {
    const { t } = useT("settings");

    return (
        <>
            <Section
                title={t("settings:security.contact.title")}
                description={t("settings:security.contact.description")}
                contentClassName="space-y-4 p-4"
            >
                <div className="flex flex-row items-center justify-between">
                    <h2 className="text-sm font-medium">
                        {t("settings:security.contact.email")}
                    </h2>
                    <p className="text-medium text-foreground font-semibold italic">
                        {user?.email}
                    </p>
                </div>
                <PhoneNumber />
            </Section>

            <Section
                title={t("settings:security.security.title")}
                description={t("settings:security.security.description")}
                contentClassName="space-y-4 rounded-lg border border-border p-4"
                className="mt-4"
            >
                <ApiKeyTab />
                <div className="border-border flex items-start justify-between border-t pt-4">
                    <h2 className="text-sm font-medium">
                        {t("settings:security.security.password.title")}
                    </h2>
                    <div className="flex flex-col items-end gap-2">
                        <Dialog>
                            <DialogTrigger asChild>
                                <Button
                                    disabled={
                                        user?.authProvider !==
                                        AuthProvider.LOCAL
                                    }
                                >
                                    {t(
                                        "settings:security.security.password.change",
                                    )}
                                </Button>
                            </DialogTrigger>
                            <DialogContent className="sm:max-w-[425px]">
                                <DialogHeader>
                                    <DialogTitle>
                                        {t(
                                            "settings:security.security.password.dialog_title",
                                        )}
                                    </DialogTitle>
                                </DialogHeader>
                                <ChangePasswordTab />
                            </DialogContent>
                        </Dialog>
                        <span className="text-foreground-muted text-sm">
                            {user?.authProvider === AuthProvider.LOCAL
                                ? t(
                                      "settings:security.security.password.local_description",
                                  )
                                : t(
                                      "settings:security.security.password.third_party_description",
                                  )}
                        </span>
                    </div>
                </div>

                <div className="border-border border-t pt-4">
                    <div className="flex flex-row justify-between space-y-2">
                        <div>
                            <h2 className="text-sm font-medium">
                                {t(
                                    "settings:security.security.two_factor.title",
                                )}
                            </h2>
                            <p
                                className={cn(
                                    "text-sm",
                                    user.authProvider === AuthProvider.LOCAL
                                        ? "text-foreground-muted"
                                        : "text-success",
                                )}
                            >
                                {user.authProvider !== AuthProvider.LOCAL
                                    ? t(
                                          "settings:security.security.two_factor.third_party_description",
                                      )
                                    : t(
                                          "settings:security.security.two_factor.local_description",
                                      )}
                            </p>
                        </div>

                        <Button disabled={true}>
                            {t("settings:security.security.two_factor.setup")}
                        </Button>
                    </div>
                </div>

                <div className="border-border border-t pt-4">
                    <div className="space-y-2">
                        <div className="flex items-center justify-between">
                            <h2 className="text-sm font-medium">
                                {t("settings:security.security.sign_out.title")}
                            </h2>
                        </div>
                        <p className="text-foreground-muted text-sm">
                            {t(
                                "settings:security.security.sign_out.description",
                            )}{" "}
                            <a
                                href="#"
                                className="text-purple-400 hover:text-purple-300"
                            >
                                {t(
                                    "settings:security.security.sign_out.change_password",
                                )}
                            </a>
                            .
                        </p>
                        <Button variant="destructive">
                            {t("settings:security.security.sign_out.button")}
                        </Button>
                    </div>
                </div>
            </Section>
        </>
    );
}
