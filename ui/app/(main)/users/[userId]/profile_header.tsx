import { Button } from "@/components/ui/button";
import useUser from "@/hooks/user";
import { User } from "@/types/user";
import { ShieldCheck } from "lucide-react";

export default function ProfileHeader({ user }: { user: User }) {
    const me = useUser((state) => state.user);

    return (
        <div className="flex items-start gap-8">
            <div>
                <h1 className="text-3xl font-bold text-gray-900">
                    {user.displayName ?? user.username}
                    {user?.isVerified && <ShieldCheck color="#10b981" className="p-[1px]" />}
                </h1>
                <p className="text-gray-500">@{user.username}</p>
            </div>
            {me?.id !== user.id && (
                <Button className="bg-purple-600 hover:bg-purple-700 text-white border-none">
                    Follow
                </Button>
            )}
        </div>
    );
}
