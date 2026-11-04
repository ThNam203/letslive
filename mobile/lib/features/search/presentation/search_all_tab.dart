import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/features/home/presentation/feed_section.dart';
import 'package:letslive_mobile/features/search/domain/search_type.dart';
import 'package:letslive_mobile/features/search/presentation/search_result_rows.dart';
import 'package:letslive_mobile/features/search/search_providers.dart';

/// The All tab (web `search-all-tab.tsx`): up to [searchPreviewLimit] channels,
/// live streams and videos for [query]. The groups load and fail on their own
/// (an error with a retry in that group only); a group with no results is left
/// out, and when all three are empty a no-results message replaces them. A
/// group that filled its preview offers "See all", which calls [onSeeAll].
class SearchAllTab extends ConsumerWidget {
  const SearchAllTab({required this.query, required this.onSeeAll, super.key});

  final String query;
  final void Function(SearchType type) onSeeAll;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final users = ref.watch(usersPreviewProvider(query));
    final lives = ref.watch(livesPreviewProvider(query));
    final vods = ref.watch(vodsPreviewProvider(query));

    final nothingFound = [
      users.value?.isEmpty,
      lives.value?.isEmpty,
      vods.value?.isEmpty,
    ].every((empty) => empty == true);
    if (nothingFound) {
      return ListView(
        children: [
          EmptyNotice(
            icon: Icons.search_off,
            title: l10n.common_search_no_results(query),
            description: l10n.common_search_no_results_hint,
          ),
        ],
      );
    }

    return ListView(
      children: [
        FeedSection(
          title: l10n.common_search_sections_users,
          value: users,
          onRetry: () => ref.invalidate(usersPreviewProvider(query)),
          builder: (context, items) => _Group(
            full: items.length == searchPreviewLimit,
            onSeeAll: () => onSeeAll(SearchType.users),
            rows: [
              for (final user in items)
                UserResultRow(key: ValueKey(user.id), user: user, query: query),
            ],
          ),
        ),
        FeedSection(
          title: l10n.common_search_sections_lives,
          value: lives,
          onRetry: () => ref.invalidate(livesPreviewProvider(query)),
          builder: (context, items) => _Group(
            full: items.length == searchPreviewLimit,
            onSeeAll: () => onSeeAll(SearchType.lives),
            rows: [
              for (final stream in items)
                LiveResultRow(
                  key: ValueKey(stream.id),
                  stream: stream,
                  query: query,
                ),
            ],
          ),
        ),
        FeedSection(
          title: l10n.common_search_sections_vods,
          value: vods,
          onRetry: () => ref.invalidate(vodsPreviewProvider(query)),
          builder: (context, items) => _Group(
            full: items.length == searchPreviewLimit,
            onSeeAll: () => onSeeAll(SearchType.vods),
            rows: [
              for (final vod in items)
                VodResultRow(key: ValueKey(vod.id), vod: vod, query: query),
            ],
          ),
        ),
      ],
    );
  }
}

/// The rows of one group, then "See all" when the group may have more.
class _Group extends StatelessWidget {
  const _Group({
    required this.rows,
    required this.full,
    required this.onSeeAll,
  });

  final List<Widget> rows;
  final bool full;
  final VoidCallback onSeeAll;

  @override
  Widget build(BuildContext context) => Column(
    children: [
      ...rows,
      if (full)
        Align(
          alignment: AlignmentDirectional.centerEnd,
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 8),
            child: TextButton(
              onPressed: onSeeAll,
              child: Text(AppLocalizations.of(context).common_search_see_all),
            ),
          ),
        ),
    ],
  );
}
