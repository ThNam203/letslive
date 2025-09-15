"use client";

import useUser from "@/hooks/user";
import ContactSettings from "./_components/profile";

export default function SecurityPage() {
  const user = useUser((state) => state.user);

  if (!user) return null;
  return <ContactSettings user={user} />;
}
