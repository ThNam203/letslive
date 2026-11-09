import { Suspense } from "react";
import SearchView from "./_components/search-view";
import { SearchRowsSkeleton } from "./_components/search-states";

export default function SearchPage() {
    return (
        <Suspense
            fallback={
                <div className="mx-auto w-full max-w-4xl px-3 pt-6 sm:px-6">
                    <SearchRowsSkeleton count={3} />
                </div>
            }
        >
            <SearchView />
        </Suspense>
    );
}
