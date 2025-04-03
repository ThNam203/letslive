import { ShieldCheck, ShieldOff } from "lucide-react";
import React from "react";
interface Props {
  isVerified?: boolean;
}

export default function UserVerification({ isVerified = false }: Props) {
  return (
    <div className="flex flex-row gap-1 items-center">
      {isVerified ? (
        <ShieldCheck color="#10b981" />
      ) : (
        <ShieldOff color="#ef4444" />
      )}

      {isVerified ? (
        <p className="text-emerald-500 font-medium text-sm">Verified.</p>
      ) : (
        <p className="text-red-500 font-medium text-sm">Unverified.</p>
      )}
    </div>
  );
}
