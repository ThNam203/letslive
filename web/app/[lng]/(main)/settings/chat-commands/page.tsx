"use client";

import { useEffect, useState } from "react";
import { toast } from "@/components/utils/toast";
import { Button } from "@/components/ui/button";
import IconClose from "@/components/icons/close";
import IconLoader from "@/components/icons/loader";
import IconPencil from "@/components/icons/pencil";
import Section from "../_components/section";
import TextField from "../_components/text-field";
import {
    CreateChatCommand,
    DeleteChatCommand,
    GetMyChatCommands,
    UpdateChatCommand,
} from "@/lib/api/chat-command";
import {
    ChatCommand,
    ChatCommandScope,
    MyChatCommands,
} from "@/types/chat-command";
import { BUILTIN_CHAT_COMMANDS } from "@/utils/chat-parser";
import useT from "@/hooks/use-translation";

const NAME_PATTERN = /^[a-z0-9_-]{1,32}$/;
const MAX_RESPONSE = 500;
const MAX_DESCRIPTION = 120;

export default function ChatCommandsSettings() {
    const { t } = useT("chat-commands");
    const [data, setData] = useState<MyChatCommands>({ user: [], channel: [] });
    const [loading, setLoading] = useState(true);

    const refresh = async () => {
        const res = await GetMyChatCommands();
        if (res.success && res.data) setData(res.data);
        setLoading(false);
    };

    useEffect(() => {
        refresh();
    }, []);

    const handleDelete = async (id: string) => {
        const res = await DeleteChatCommand(id);
        if (res.success) {
            toast.success(t("chat-commands:page.removed_toast"));
            refresh();
        } else {
            toast(t("chat-commands:page.remove_failed_toast"), {
                type: "error",
            });
        }
    };

    return (
        <div className="space-y-8">
            <ChatCommandScopeSection
                title={t("chat-commands:page.personal_title")}
                description={t("chat-commands:page.personal_description")}
                scope="user"
                items={data.user}
                onChanged={refresh}
                onDelete={handleDelete}
            />
            <ChatCommandScopeSection
                title={t("chat-commands:page.channel_title")}
                description={t("chat-commands:page.channel_description")}
                scope="channel"
                items={data.channel}
                onChanged={refresh}
                onDelete={handleDelete}
            />
            <Section
                title={t("chat-commands:page.builtin_title")}
                description={t("chat-commands:page.builtin_description")}
                contentClassName="p-4"
            >
                <ul className="space-y-2 text-sm">
                    {BUILTIN_CHAT_COMMANDS.map((c) => (
                        <li key={c.name} className="flex justify-between gap-4">
                            <span className="font-mono">{c.usage}</span>
                            <span className="text-muted-foreground">
                                {t(c.descriptionKey)}
                            </span>
                        </li>
                    ))}
                </ul>
            </Section>
            {loading && (
                <div className="flex justify-center py-4">
                    <IconLoader />
                </div>
            )}
        </div>
    );
}

function ChatCommandScopeSection({
    title,
    description,
    scope,
    items,
    onChanged,
    onDelete,
}: {
    title: string;
    description: string;
    scope: ChatCommandScope;
    items: ChatCommand[];
    onChanged: () => void;
    onDelete: (id: string) => void;
}) {
    const { t } = useT("chat-commands");
    const [editing, setEditing] = useState<ChatCommand | null>(null);
    const [name, setName] = useState("");
    const [response, setResponse] = useState("");
    const [desc, setDesc] = useState("");
    const [submitting, setSubmitting] = useState(false);

    const reset = () => {
        setEditing(null);
        setName("");
        setResponse("");
        setDesc("");
    };

    const startEdit = (cmd: ChatCommand) => {
        setEditing(cmd);
        setName(cmd.name);
        setResponse(cmd.response);
        setDesc(cmd.description ?? "");
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        const trimmedName = name.trim().toLowerCase();
        if (!NAME_PATTERN.test(trimmedName)) {
            toast(t("chat-commands:page.name_invalid_toast"), {
                type: "error",
            });
            return;
        }
        if (!response.trim()) {
            toast(t("chat-commands:page.response_required_toast"), {
                type: "error",
            });
            return;
        }
        setSubmitting(true);
        try {
            if (editing) {
                const res = await UpdateChatCommand(editing.id, {
                    name: trimmedName,
                    response: response.trim(),
                    description: desc.trim(),
                });
                if (res.success) {
                    toast.success(
                        t("chat-commands:page.updated_toast", {
                            name: trimmedName,
                        }),
                    );
                    reset();
                    onChanged();
                } else {
                    toast(t("chat-commands:page.update_failed_toast"), {
                        type: "error",
                    });
                }
            } else {
                const res = await CreateChatCommand({
                    scope,
                    name: trimmedName,
                    response: response.trim(),
                    description: desc.trim(),
                });
                if (res.success) {
                    toast.success(
                        t("chat-commands:page.added_toast", {
                            name: trimmedName,
                        }),
                    );
                    reset();
                    onChanged();
                } else {
                    toast(t("chat-commands:page.create_failed_toast"), {
                        type: "error",
                    });
                }
            }
        } finally {
            setSubmitting(false);
        }
    };

    return (
        <Section
            title={title}
            description={description}
            contentClassName="p-4 space-y-4"
        >
            {items.length === 0 ? (
                <p className="text-muted-foreground text-sm">
                    {t("chat-commands:page.empty")}
                </p>
            ) : (
                <ul className="space-y-2">
                    {items.map((c) => (
                        <li
                            key={c.id}
                            className="border-border flex items-start gap-3 rounded-md border p-3"
                        >
                            <div className="min-w-0 flex-1">
                                <div className="font-mono font-medium">
                                    /{c.name}
                                </div>
                                <div className="text-muted-foreground text-sm break-words">
                                    {c.response}
                                </div>
                                {c.description && (
                                    <div className="text-muted-foreground mt-1 text-xs">
                                        {c.description}
                                    </div>
                                )}
                            </div>
                            <Button
                                variant="ghost"
                                size="icon"
                                onClick={() => startEdit(c)}
                                aria-label={t("chat-commands:page.edit_aria")}
                            >
                                <IconPencil className="h-4 w-4" />
                            </Button>
                            <Button
                                variant="ghost"
                                size="icon"
                                onClick={() => onDelete(c.id)}
                                aria-label={t("chat-commands:page.remove_aria")}
                            >
                                <IconClose className="h-4 w-4" />
                            </Button>
                        </li>
                    ))}
                </ul>
            )}

            <form
                onSubmit={handleSubmit}
                className="border-border space-y-3 border-t pt-4"
            >
                {editing && (
                    <div className="text-muted-foreground text-sm">
                        {t("chat-commands:page.edit_title", {
                            name: editing.name,
                        })}
                    </div>
                )}
                <TextField
                    label={t("chat-commands:page.name_label")}
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    placeholder={t("chat-commands:page.name_placeholder")}
                    maxLength={32}
                />
                <TextField
                    label={t("chat-commands:page.response_label")}
                    value={response}
                    onChange={(e) => setResponse(e.target.value)}
                    placeholder={t("chat-commands:page.response_placeholder")}
                    maxLength={MAX_RESPONSE}
                />
                <TextField
                    label={t("chat-commands:page.description_label")}
                    value={desc}
                    onChange={(e) => setDesc(e.target.value)}
                    placeholder={t("chat-commands:page.description_placeholder")}
                    maxLength={MAX_DESCRIPTION}
                />
                <div className="flex justify-end gap-2">
                    {editing && (
                        <Button
                            type="button"
                            variant="ghost"
                            onClick={reset}
                            disabled={submitting}
                        >
                            {t("chat-commands:page.cancel_edit")}
                        </Button>
                    )}
                    <Button type="submit" disabled={submitting}>
                        {submitting && <IconLoader />}
                        {editing
                            ? t("chat-commands:page.submit_edit")
                            : t("chat-commands:page.submit")}
                    </Button>
                </div>
            </form>
        </Section>
    );
}
