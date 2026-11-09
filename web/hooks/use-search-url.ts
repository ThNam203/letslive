"use client";

import { useEffect, useRef, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import {
    SEARCH_TYPING_DEBOUNCE_MS,
    SearchTab,
    parseSearchTab,
    searchHref,
} from "@/lib/search";

/** Search text and tab on /search, kept in the URL so results can be shared. */
export default function useSearchUrl() {
    const router = useRouter();
    const params = useSearchParams();
    const query = params.get("q")?.trim() ?? "";
    const tab = parseSearchTab(params.get("type"));

    const [draft, setDraft] = useState(query);
    const [lastWritten, setLastWritten] = useState(query);
    const [syncedQuery, setSyncedQuery] = useState(query);
    // Back/forward changed the text in the URL: show it. The URL catching up
    // with what this box wrote must not overwrite what the user kept typing.
    if (query !== syncedQuery) {
        setSyncedQuery(query);
        if (query !== lastWritten) setDraft(query);
    }

    const timer = useRef<ReturnType<typeof setTimeout> | null>(null);
    const cancelPending = () => {
        if (timer.current) clearTimeout(timer.current);
        timer.current = null;
    };

    useEffect(() => {
        window.addEventListener("popstate", cancelPending);
        return () => {
            window.removeEventListener("popstate", cancelPending);
            cancelPending();
        };
    }, []);

    const write = (nextQuery: string, nextTab: SearchTab) => {
        setLastWritten(nextQuery);
        router.replace(searchHref(nextQuery, nextTab), { scroll: false });
    };

    const changeDraft = (value: string) => {
        setDraft(value);
        cancelPending();
        timer.current = setTimeout(() => {
            // the tab may have changed while the user was typing
            const currentTab = parseSearchTab(
                new URLSearchParams(window.location.search).get("type"),
            );
            write(value.trim(), currentTab);
        }, SEARCH_TYPING_DEBOUNCE_MS);
    };

    const submit = () => {
        cancelPending();
        write(draft.trim(), tab);
    };

    const selectTab = (nextTab: SearchTab) => write(query, nextTab);

    return { query, tab, draft, changeDraft, submit, selectTab };
}
