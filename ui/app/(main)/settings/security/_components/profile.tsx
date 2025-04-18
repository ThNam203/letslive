import { Button } from "../../../../../components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../../../../../components/ui/dialog";
import { AuthProvider, User } from "../../../../../types/user";
import Section from "../../_components/section";
import ApiKeyTab from "./api-key-tab";
import ChangePasswordTab from "./change-password-tab";

export default function ContactSettings({ user }: { user: User }) {
  return (
    <div className="min-h-screen text-gray-900 p-6">
      <div className="max-w-4xl">
        <Section
          title="Contact"
          description="Where we send important messages about your account"
          hasBorder
          contentClassName="space-y-4"
        >
          <div className="flex flex-row items-center justify-between">
            <h2 className="text-sm font-medium">Email</h2>
            <p className="text-gray-900 font-semibold text-medium italic">
              {user?.email}
            </p>
          </div>
          <div className="flex items-center justify-between border-t border-zinc-800 pt-4">
            <h2 className="text-sm font-medium">Phone Number</h2>
            <a
              href="#"
              className="text-purple-600 hover:text-purple-700 text-sm"
            >
              Add a number
            </a>
          </div>
        </Section>

        <Section
          title="Security"
          description="Keep your account safe and sound"
          contentClassName="space-y-4 rounded-lg border-1 border-gray-900 p-4"
          className="mt-4"
        >
          <ApiKeyTab />
          <div className="flex items-start justify-between border-t border-zinc-800 pt-4">
            <h2 className="text-sm font-medium">Password</h2>
            <div className="flex flex-col items-end gap-2">
              <Dialog>
                <DialogTrigger asChild>
                  <Button
                    variant="outline"
                    disabled={user?.authProvider !== AuthProvider.LOCAL}
                    className="bg-purple-600 hover:bg-purple-700 text-white text-sm hover:text-white font-normal"
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
              <span className="text-gray-400 text-sm">
                {user?.authProvider === AuthProvider.LOCAL
                  ? "Improve your security with a strong password"
                  : "Your account is secured with a password from a third-party provider."}
                .
              </span>
            </div>
          </div>

          <div className="border-t border-zinc-800 pt-4">
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <h2 className="text-sm font-medium">
                  Two-Factor Authentication
                </h2>
              </div>
              <button
                disabled={true}
                className="bg-purple-600 hover:bg-purple-700 text-white px-4 py-2 rounded text-sm"
              >
                Set Up Two-Factor Authentication
              </button>
              <p className="text-sm text-gray-400">
                {user.authProvider !== AuthProvider.LOCAL
                  ? `Your account is secured with third-party provider.`
                  : "Add an extra layer of security to your Let&apos;s Live account by enabling 2FA before logging in."}
              </p>
            </div>
          </div>

          <div className="border-t border-zinc-800 pt-4">
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <h2 className="text-sm font-medium">Sign Out Everywhere</h2>
              </div>
              <p className="text-sm text-gray-400">
                This will log out you of Let&apos;s Live everywhere you&apos;re
                logged in. If you believe your account has been compromised, we
                recommend you{" "}
                <a href="#" className="text-purple-400 hover:text-purple-300">
                  change your password
                </a>
                .
              </p>
              <button className="bg-zinc-700 hover:bg-zinc-600 text-white px-4 py-2 rounded text-sm">
                Sign Out Everywhere
              </button>
            </div>
          </div>
        </Section>
      </div>
    </div>
  );
}
