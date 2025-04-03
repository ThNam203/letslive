"use client";
import { cn } from "@/utils/cn";
import { Tv } from "lucide-react";
import React from "react";
import { Popover, PopoverContent, PopoverTrigger } from "../ui/popover";

export default function StreamGuiding() {
  const [open, setOpen] = React.useState(false);
  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger>
        <Tv
          className={cn(
            "bg-white text-purple-600 cursor-pointer scale-80 hover:scale-100 transition-all duration-700 animate-bounce",
            open && "scale-100 animate-none"
          )}
        />
      </PopoverTrigger>
      <PopoverContent className="w-100 mr-4 text-sm">
        <h1 className="font-semibold text-xl">Livestreaming</h1>
        <p>How to start your livestream: </p>
        <p>Open OBS &rarr; Settings &rarr; Stream </p>
        <p>
          Enter: &quot;Server: rtmp://
          {process.env.NEXT_PUBLIC_ENVIRONMENT === "production"
            ? "sen1or-huly.com"
            : "localhost"}
          :1935, StreamKey: Your key in Security Setting&quot;
        </p>
        <p className="mb-2">Start your livestream</p>
      </PopoverContent>
    </Popover>
  );
}
