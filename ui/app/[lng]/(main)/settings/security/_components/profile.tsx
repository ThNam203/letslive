import { cn } from "@/utils/cn";
import { Button } from "@/components/ui/button";
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog";
import { AuthProvider, User } from "@/types/user";
import Section from "../../_components/section";
import ApiKeyTab from "./api-key-tab";
import ChangePasswordTab from "./change-password-tab";

export default function ContactSettings({ user }: { user: User }) {
    return (
        <>
            <Section
                title="Contact"
                description="Where we send important messages about your account"
                hasBorder
                contentClassName="space-y-4"
            >
                <div className="flex flex-row items-center justify-between">
                    <h2 className="text-sm font-medium">Email</h2>
                    <p className="text-medium font-semibold italic text-foreground">
                        {user?.email}
                    </p>
                </div>
                <div className="flex items-center justify-between border-t border-border pt-4">
                    <h2 className="text-sm font-medium">Phone Number</h2>
                    <a
                        href="#"
                        className="text-sm text-primary hover:text-primary-hover"
                    >
                        Add a number
                    </a>
                </div>
            </Section>

            <Section
                title="Security"
                description="Keep your account safe and sound"
                contentClassName="space-y-4 rounded-lg border border-border p-4"
                className="mt-4"
            >
                <ApiKeyTab />
                <div className="flex items-start justify-between border-t border-border pt-4">
                    <h2 className="text-sm font-medium">Password</h2>
                    <div className="flex flex-col items-end gap-2">
                        <Dialog>
                            <DialogTrigger asChild>
                                <Button
                                    disabled={
                                        user?.authProvider !==
                                        AuthProvider.LOCAL
                                    }
                                >
                                    Change password
                                </Button>
                            </DialogTrigger>
                            <DialogContent className="sm:max-w-[425px]">
                                <DialogHeader>
                                    <DialogTitle>Change Password</DialogTitle>
                                </DialogHeader>
                                <ChangePasswordTab />
                            </DialogContent>
                        </Dialog>
                        <span className="text-sm text-foreground-muted">
                            {user?.authProvider === AuthProvider.LOCAL
                                ? "Improve your security with a strong password."
                                : "Your account is secured with a password from a third-party provider."}
                        </span>
                    </div>
                </div>

                <div className="border-t border-border pt-4">
                    <div className="flex flex-row justify-between space-y-2">
                        <div>
                            <h2 className="text-sm font-medium">
                                Two-Factor Authentication
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
                                    ? `Your account is secured with third-party provider.`
                                    : "Add an extra layer of security to your Let&apos;s Live account by enabling 2FA before logging in."}
                            </p>
                        </div>

                        <Button disabled={true}>
                            Set Up Two-Factor Authentication
                        </Button>
                    </div>
                </div>

                <div className="border-t border-border pt-4">
                    <div className="space-y-2">
                        <div className="flex items-center justify-between">
                            <h2 className="text-sm font-medium">
                                Sign Out Everywhere
                            </h2>
                        </div>
                        <p className="text-sm text-foreground-muted">
                            This will log out you of Let&apos;s Live everywhere
                            you&apos;re logged in. If you believe your account
                            has been compromised, we recommend you{" "}
                            <a
                                href="#"
                                className="text-purple-400 hover:text-purple-300"
                            >
                                change your password
                            </a>
                            .
                        </p>
                        <Button variant="destructive">
                            Sign Out Everywhere
                        </Button>
                    </div>
                </div>
            </Section>
        </>
    );
}
