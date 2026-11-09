/// The tabs of the search screen and the numbers that shape a search, all
/// taken from the web app (`web/lib/search.ts`, `web/constant/field-limits.ts`,
/// `web/hooks/queries/use-search.ts`) and `docs/SEARCH.md`.
library;

/// What the search screen shows: everything, or one kind of result.
enum SearchType { all, users, lives, vods }

/// Wait after the last keystroke before searching
/// (web `SEARCH_TYPING_DEBOUNCE_MS`).
const Duration searchDebounce = Duration(milliseconds: 300);

/// Results per group on the All tab (web `SEARCH_PREVIEW_LIMIT`).
const int searchPreviewLimit = 5;

/// Results per page on a type tab (web `SEARCH_PAGE_SIZE`; the server's
/// maximum).
const int searchPageSize = 20;
