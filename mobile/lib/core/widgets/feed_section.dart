import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/l10n/api_message.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/network/api_load.dart';
import 'package:letslive_mobile/core/pagination/paged_feed.dart';

/// The body to show instead of content while [value] is loading or failed,
/// or when its list is empty: a spinner, an error with a retry button,
/// [empty] (an empty box when null). Null when there are items to show.
///
/// While reloading with data on screen the old content stays (no flicker on
/// pull-to-refresh); after an error the spinner shows so a retry gives
/// feedback.
Widget? feedStatusBody<T>({
  required AsyncValue<List<T>> value,
  required VoidCallback onRetry,
  Widget? empty,
}) {
  if (value.hasError && !value.isLoading) {
    return FeedError(error: value.error!, onRetry: onRetry);
  }
  final items = value.value;
  if (items == null) {
    return const Padding(
      padding: EdgeInsets.all(24),
      child: Center(child: CircularProgressIndicator()),
    );
  }
  if (items.isEmpty) return empty ?? const SizedBox.shrink();
  return null;
}

/// Title of a feed section.
class SectionHeader extends StatelessWidget {
  const SectionHeader(this.title, {super.key});

  final String title;

  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
    child: Text(
      title,
      maxLines: 1,
      overflow: TextOverflow.ellipsis,
      style: Theme.of(context).textTheme.titleLarge,
    ),
  );
}

/// One titled block of the home feed that loads on its own: [feedStatusBody]
/// while loading, failed or empty, else [builder]'s content.
class FeedSection<T> extends StatelessWidget {
  const FeedSection({
    required this.title,
    required this.value,
    required this.onRetry,
    required this.builder,
    this.empty,
    super.key,
  });

  final String title;
  final AsyncValue<List<T>> value;
  final VoidCallback onRetry;
  final Widget Function(BuildContext context, List<T> items) builder;

  /// Shown for an empty list; null hides the whole section (like web's
  /// channel lists).
  final Widget? empty;

  @override
  Widget build(BuildContext context) {
    final items = value.value;
    final hasError = value.hasError && !value.isLoading;
    if (!hasError && items != null && items.isEmpty && empty == null) {
      return const SizedBox.shrink();
    }
    final status = feedStatusBody(value: value, onRetry: onRetry, empty: empty);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [SectionHeader(title), status ?? builder(context, items!)],
    );
  }
}

/// Title and optional description of an empty section (web's empty states).
class EmptyNotice extends StatelessWidget {
  const EmptyNotice({
    required this.icon,
    required this.title,
    this.description,
    super.key,
  });

  final IconData icon;
  final String title;
  final String? description;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Padding(
      padding: const EdgeInsets.all(24),
      child: Column(
        children: [
          Icon(icon, size: 48, color: theme.colorScheme.onSurfaceVariant),
          const SizedBox(height: 12),
          Text(
            title,
            textAlign: TextAlign.center,
            style: theme.textTheme.titleMedium,
          ),
          if (description != null) ...[
            const SizedBox(height: 4),
            Text(description!, textAlign: TextAlign.center),
          ],
        ],
      ),
    );
  }
}

/// The localized reason a section (or page) failed to load, with a retry
/// button.
class FeedError extends StatelessWidget {
  const FeedError({required this.error, required this.onRetry, super.key});

  /// An [ApiLoadException] for a server or network error; anything else
  /// shows the generic message.
  final Object error;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final failed = error;
    return Padding(
      padding: const EdgeInsets.all(16),
      child: Column(
        children: [
          Text(
            failed is ApiLoadException
                ? apiErrorMessage(l10n, failed.response)
                : l10n.error_general_description,
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 8),
          FilledButton(onPressed: onRetry, child: Text(l10n.auth_retry)),
        ],
      ),
    );
  }
}

/// End of a paged list. Because a lazy list builds it only when it is within
/// reach of the viewport, building it is the "near the end" signal: it asks
/// for the next page via [onLoadMore] (a no-op while one is loading or after
/// the last page). A failed page shows an inline retry instead of looping.
class PagedFooter<T> extends StatelessWidget {
  const PagedFooter({required this.feed, required this.onLoadMore, super.key});

  /// The list so far; null before the first page arrived (shows nothing).
  final PagedFeed<T>? feed;

  /// Loads the next page; also the retry after an error.
  final VoidCallback onLoadMore;

  @override
  Widget build(BuildContext context) {
    final current = feed;
    if (current == null) return const SizedBox.shrink();
    final error = current.loadMoreError;
    if (error != null) return FeedError(error: error, onRetry: onLoadMore);
    if (!current.hasMore) return const SizedBox(height: 16);
    // Not during build: loading a page changes provider state.
    Future.microtask(onLoadMore);
    return const Padding(
      padding: EdgeInsets.all(16),
      child: Center(child: CircularProgressIndicator()),
    );
  }
}
