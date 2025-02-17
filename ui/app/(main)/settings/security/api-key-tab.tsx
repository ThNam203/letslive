"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Copy, RefreshCw } from "lucide-react";
import { User } from "@/types/user";
import { toast } from "react-toastify";

export default function ApiKeyTab({ user }: { user: User | undefined }) {
    const [apiKey, setApiKey] = useState(user ? user.streamAPIKey : "");

    const generateNewApiKey = () => {
        if (!user) return;

        const newApiKey = "yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy";
        setApiKey(newApiKey);
    };

    const copyApiKey = () => {
        navigator.clipboard.writeText(apiKey);
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
                    value={apiKey}
                    readOnly={true}
                    className="flex-grow"
                />
            </div>
            <div className="flex gap-4">
                <div className="flex-grow" />
                <Button
                    className="bg-purple-600 hover:bg-purple-700"
                    onClick={copyApiKey}
                >
                    <Copy className="h-4 w-4" />
                </Button>

                <Button
                    className="bg-purple-600 hover:bg-purple-700"
                    onClick={generateNewApiKey}
                >
                    <RefreshCw className="mr-2 h-4 w-4" /> Generate New API Key
                </Button>
            </div>
        </div>
    );
}
