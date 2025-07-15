"use client";

import { useState } from "react";
import { Copy, RefreshCw } from "lucide-react";
import { toast } from "react-toastify";
import useUser from "../../../../../hooks/user";
import { RequestToGenerateNewAPIKey } from "../../../../../lib/api/user";
import { Label } from "../../../../../components/ui/label";
import { Input } from "../../../../../components/ui/input";
import { Button } from "../../../../../components/ui/button";

export default function ApiKeyTab() {
  const user = useUser((state) => state.user);
  const updateUser = useUser((state) => state.updateUser);
  const [isGenerating, setIsGenerating] = useState(false);

  const generateNewApiKey = async () => {
    if (!user) return;

    setIsGenerating(true);
    const { newKey, fetchError } = await RequestToGenerateNewAPIKey();

    if (fetchError) toast(fetchError.message, { type: "error" });
    else
      updateUser({
        ...user,
        streamAPIKey: newKey!,
      });
    setIsGenerating(false);
  };

  const copyApiKey = () => {
    if (!user) return;
    navigator.clipboard.writeText(user?.streamAPIKey);
    toast.success("API Key copied to clipboard");
  };

  return (
    <div>
      <div className="flex flex-row justify-between items-center mb-4">
        <Label className="min-w-[200px]" htmlFor="api-key">
          Your API Key
        </Label>
        <Input
          id="api-key"
          value={user?.streamAPIKey}
          readOnly={true}
          className="flex-grow text-right border border-border"
        />
      </div>
      <div className="flex gap-4">
        <div className="flex-grow" />
        <Button onClick={copyApiKey}>
          <Copy className="h-4 w-4 text-primary-foreground" />
        </Button>

        <Button onClick={generateNewApiKey}>
          <RefreshCw className="mr-2 h-4 w-4" /> Generate New API Key
        </Button>
      </div>
    </div>
  );
}
