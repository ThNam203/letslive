"use client";

import {
    Tabs,
    TabsContent,
    TabsList,
    TabsTrigger,
    UNDERLINE_TABS_LIST_CLASS,
    UNDERLINE_TABS_TRIGGER_CLASS,
} from "@/components/ui/tabs";
import SearchInput from "@/components/search/search-input";
import useT from "@/hooks/use-translation";
import useSearchUrl from "@/hooks/use-search-url";
import { RAW_INTERPOLATION, SEARCH_TABS, parseSearchTab } from "@/lib/search";
import SearchAllTab from "./search-all-tab";
import SearchTypeTab from "./search-type-tab";
import { SearchEmpty } from "./search-states";

export default function SearchView() {
    const { t } = useT("common");
    const { query, tab, draft, changeDraft, submit, selectTab } =
        useSearchUrl();

    return (
        <div className="h-full overflow-y-auto">
            <div className="mx-auto w-full max-w-4xl px-3 pt-3 pb-16 sm:px-6 md:pt-6">
                {/* The header search box is hidden on phones, so the page has its own. */}
                <SearchInput
                    value={draft}
                    onChange={changeDraft}
                    onSubmit={submit}
                    autoFocus={!query}
                    className="mb-3 md:hidden"
                    inputClassName="h-11 text-base"
                />
                <h1 className="sr-only md:not-sr-only md:mb-3 md:block md:truncate md:text-xl md:font-semibold">
                    {query
                        ? t("common:search.results_for", {
                              query,
                              ...RAW_INTERPOLATION,
                          })
                        : t("common:search.title")}
                </h1>

                <Tabs
                    value={tab}
                    onValueChange={(value) => selectTab(parseSearchTab(value))}
                >
                    <TabsList className={UNDERLINE_TABS_LIST_CLASS}>
                        {SEARCH_TABS.map((key) => (
                            <TabsTrigger
                                key={key}
                                value={key}
                                className={UNDERLINE_TABS_TRIGGER_CLASS}
                            >
                                {t(`common:search.tabs.${key}`)}
                            </TabsTrigger>
                        ))}
                    </TabsList>

                    {!query ? (
                        <SearchEmpty title={t("common:search.empty_prompt")} />
                    ) : (
                        <>
                            <TabsContent value="all" className="mt-4">
                                <SearchAllTab
                                    query={query}
                                    onSeeAll={selectTab}
                                />
                            </TabsContent>
                            <TabsContent value="users" className="mt-2">
                                <SearchTypeTab type="users" query={query} />
                            </TabsContent>
                            <TabsContent value="lives" className="mt-2">
                                <SearchTypeTab type="lives" query={query} />
                            </TabsContent>
                            <TabsContent value="vods" className="mt-2">
                                <SearchTypeTab type="vods" query={query} />
                            </TabsContent>
                        </>
                    )}
                </Tabs>
            </div>
        </div>
    );
}
