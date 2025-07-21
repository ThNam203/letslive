"use client";
import { cn } from "@/utils/cn";
import React, { useEffect } from "react";
import { Popover, PopoverContent, PopoverTrigger } from "../ui/popover";
import IconLiveStream from "../icons/live-stream";
import { Button } from "../ui/button";

export default function StreamGuiding() {
  const [open, setOpen] = React.useState(false);
  const [isGotIt, setIsGotIt] = React.useState(true);

  useEffect(() => {
    const readStreamGuiding = localStorage.getItem("readStreamGuiding");
    setIsGotIt(readStreamGuiding === "true");
  }, []);

  const handleGotIt = () => {
    setIsGotIt(true);
    setOpen(false);
    localStorage.setItem("readStreamGuiding", JSON.stringify(true)); // keep it true forever
  };

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger>
        <div
          className={cn(
            "bg-whitecursor-pointer transition-all duration-700 animate-bounce",
            open && "scale-100 animate-none",
            isGotIt && "animate-none"
          )}
        >
          <IconLiveStream className="text-primary" />
        </div>
      </PopoverTrigger>
      <PopoverContent className="w-100 mr-4 text-sm bg-muted text-foreground border border-border">
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
        <div className="w-full flex flex-row justify-center">
          <Button onClick={handleGotIt}>
            Got it
          </Button>
        </div>
      </PopoverContent>
    </Popover>
  );
}
