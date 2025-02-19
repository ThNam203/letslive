"use client";
import ContactSettings from "@/app/(main)/settings/security/profile";
import useUser from "@/hooks/user";

export default function SecurityPage() {
    const user = useUser((state) => state.user);

    if (!user) return null;
    return <ContactSettings user={user} />;
}
