/// The creator's own videos (web `/settings/vods`): every video of the user
/// with its status, loaded page by page, with edit and delete (after a
/// confirmation). A ready video opens its player page.
library;

import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:letslive_mobile/core/l10n/api_message.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/core/router/app_routes.dart';
import 'package:letslive_mobile/features/creator/creator_providers.dart';
import 'package:letslive_mobile/features/creator/presentation/vod_edit_dialog.dart';
import 'package:letslive_mobile/features/home/data/vod.dart';
import 'package:letslive_mobile/features/home/domain/media_format.dart';
import 'package:letslive_mobile/features/home/presentation/feed_section.dart';
import 'package:letslive_mobile/features/home/presentation/media_card.dart';

/// How close to the end of the list (logical pixels) the next page starts
/// loading (see `PagedFooter`).
const double _loadMoreDistance = 300;

/// The "my videos" screen.
class MyVodsScreen extends ConsumerWidget {
  const MyVodsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final feed = ref.watch(myVodsProvider);
    final status = feedStatusBody(
      value: feed.whenData((f) => f.items),
      onRetry: () => ref.invalidate(myVodsProvider),
      empty: EmptyNotice(
        icon: Icons.movie_outlined,
        title: l10n.settings_vods_empty,
      ),
    );
    final items = feed.value?.items ?? const <Vod>[];
    return Scaffold(
      appBar: AppBar(title: Text(l10n.settings_vods_title)),
      body: status != null
          ? SingleChildScrollView(child: status)
          : RefreshIndicator(
              onRefresh: () async => ref.refresh(myVodsProvider.future),
              child: ListView.builder(
                padding: const EdgeInsets.all(16),
                scrollCacheExtent: const ScrollCacheExtent.pixels(
                  _loadMoreDistance,
                ),
                itemCount: items.length + 1,
                itemBuilder: (context, index) => index < items.length
                    ? _VodCard(vod: items[index])
                    : PagedFooter(
                        feed: feed.value,
                        onLoadMore: ref.read(myVodsProvider.notifier).loadMore,
                      ),
              ),
            ),
    );
  }
}

class _VodCard extends ConsumerStatefulWidget {
  const _VodCard({required this.vod});

  final Vod vod;

  @override
  ConsumerState<_VodCard> createState() => _VodCardState();
}

class _VodCardState extends ConsumerState<_VodCard> {
  /// A delete request is running; the button is off so a double tap cannot
  /// send it twice.
  bool _deleting = false;

  Vod get vod => widget.vod;

  Future<void> _delete() async {
    final l10n = AppLocalizations.of(context);
    final messenger = ScaffoldMessenger.of(context);
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(l10n.settings_vods_delete_dialog_title),
        content: Text(l10n.settings_vods_delete_dialog_description),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(l10n.settings_vods_delete_dialog_cancel),
          ),
          FilledButton(
            key: const ValueKey('confirm-delete-vod'),
            onPressed: () => Navigator.pop(context, true),
            child: Text(l10n.settings_vods_delete_dialog_delete),
          ),
        ],
      ),
    );
    if (confirmed != true || !mounted) return;
    setState(() => _deleting = true);
    final response = await ref.read(myVodsRepositoryProvider).delete(vod.id);
    if (!mounted) return;
    setState(() => _deleting = false);
    if (response.success) {
      ref.invalidate(myVodsProvider);
    } else {
      messenger.showSnackBar(
        SnackBar(content: Text(apiErrorMessage(l10n, response))),
      );
    }
  }

  String _statusLabel(AppLocalizations l10n) => switch (vod.status) {
    'ready' =>
      vod.visibility == 'public' ? l10n.settings_vods_edit_dialog_public : '',
    'failed' => l10n.settings_vods_status_failed,
    _ => l10n.settings_vods_status_processing,
  };

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final theme = Theme.of(context);
    final apiBaseUrl = ref.watch(appConfigProvider).apiBaseUrl;
    final isReady = vod.status == 'ready';
    final label = _statusLabel(l10n);
    return Card(
      margin: const EdgeInsets.only(bottom: 16),
      clipBehavior: Clip.antiAlias,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          InkWell(
            onTap: isReady
                ? () => context.push(AppRoutes.vod(vod.userId, vod.id))
                : null,
            child: MediaThumbnail(
              url: mediaThumbnailUrl(apiBaseUrl, vod.id, vod.thumbnailUrl),
              durationLabel: isReady ? formatDuration(vod.duration) : null,
            ),
          ),
          Padding(
            padding: const EdgeInsets.fromLTRB(12, 8, 4, 4),
            child: Row(
              children: [
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        vod.title,
                        maxLines: 2,
                        overflow: TextOverflow.ellipsis,
                        style: theme.textTheme.titleSmall,
                      ),
                      Text(
                        [
                          viewsLabel(l10n, vod.viewCount),
                          if (label.isNotEmpty) label,
                        ].join(' · '),
                        maxLines: 2,
                        overflow: TextOverflow.ellipsis,
                        style: vod.status == 'failed'
                            ? theme.textTheme.bodySmall?.copyWith(
                                color: theme.colorScheme.error,
                              )
                            : theme.textTheme.bodySmall,
                      ),
                    ],
                  ),
                ),
                IconButton(
                  key: ValueKey('edit-vod-${vod.id}'),
                  tooltip: l10n.settings_vods_actions_edit,
                  icon: const Icon(Icons.edit_outlined),
                  onPressed: () => showVodEditDialog(context, vod),
                ),
                IconButton(
                  key: ValueKey('delete-vod-${vod.id}'),
                  tooltip: l10n.settings_vods_actions_delete,
                  icon: const Icon(Icons.delete_outline),
                  onPressed: _deleting ? null : _delete,
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
