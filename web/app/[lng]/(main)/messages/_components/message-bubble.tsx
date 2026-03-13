"use client";

import { DmMessage, DmMessageType } from "@/types/dm";
import useT from "@/hooks/use-translation";

function formatMessageTime(dateStr: string) {
    return new Date(dateStr).toLocaleTimeString([], {
        hour: "2-digit",
        minute: "2-digit",
    });
}

export default function MessageBubble({
    message,
    isOwn,
    showSender,
}: {
    message: DmMessage;
    isOwn: boolean;
    showSender: boolean;
}) {
    const { t } = useT("messages");

    if (message.isDeleted) {
        return (
            <div
                className={`flex ${isOwn ? "justify-end" : "justify-start"} ${showSender ? "mt-3" : "mt-0.5"}`}
            >
                <div className="bg-muted max-w-[70%] rounded-lg px-3 py-1.5 opacity-60">
                    <p className="text-muted-foreground text-sm italic">
                        {t("message_deleted")}
                    </p>
                </div>
            </div>
        );
    }

    if (message.type === DmMessageType.SYSTEM) {
        return (
            <div className="my-2 flex justify-center">
                <span className="text-muted-foreground text-xs">
                    {message.text}
                </span>
            </div>
        );
    }

    return (
        <div
            className={`flex ${isOwn ? "justify-end" : "justify-start"} ${showSender ? "mt-3" : "mt-0.5"}`}
        >
            <div
                className={`max-w-[70%] rounded-lg px-3 py-1.5 ${
                    isOwn
                        ? "bg-primary text-primary-foreground"
                        : "bg-muted text-foreground"
                }`}
            >
                {showSender && !isOwn && (
                    <p className="mb-0.5 text-xs font-semibold opacity-70">
                        {message.senderUsername}
                    </p>
                )}

                {message.type === DmMessageType.IMAGE &&
                    message.imageUrls &&
                    message.imageUrls.length > 0 && (
                        <div
                            className={`mb-1 grid gap-1 ${
                                message.imageUrls.length === 1
                                    ? "grid-cols-1"
                                    : "grid-cols-2"
                            }`}
                        >
                            {message.imageUrls.map((url, i) => (
                                <a
                                    key={i}
                                    href={url}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                >
                                    <img
                                        src={url}
                                        alt={t("image_alt", { number: i + 1 })}
                                        className="max-h-60 w-full cursor-pointer rounded object-cover transition-opacity hover:opacity-90"
                                    />
                                </a>
                            ))}
                        </div>
                    )}

                {/* Show translated caption for image-only messages */}
                {message.type === DmMessageType.IMAGE &&
                message.imageUrls &&
                message.imageUrls.length > 0 &&
                (message.text === "Sent an image" ||
                    message.text.match(/^Sent \d+ images$/)) ? (
                    <p className="text-sm break-words">
                        {message.imageUrls.length === 1
                            ? t("sent_an_image")
                            : t("sent_images_count", {
                                  count: message.imageUrls.length,
                              })}
                    </p>
                ) : (
                    <p className="text-sm break-words">{message.text}</p>
                )}

                <div className="mt-0.5 flex items-center justify-end gap-1">
                    <span className="text-[10px] opacity-60">
                        {formatMessageTime(message.createdAt)}
                    </span>
                    {message.updatedAt !== message.createdAt && (
                        <span className="text-[10px] opacity-50">(edited)</span>
                    )}
                </div>
            </div>
        </div>
    );
}
