import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:forui/forui.dart';

import '../../../l10n/app_localizations.dart';
import '../data/upload_queue_notifier.dart';

class UploadManagerOverlay extends ConsumerWidget {
  const UploadManagerOverlay({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final queueState = ref.watch(uploadQueueProvider);
    if (queueState.items.isEmpty) return const SizedBox.shrink();

    final colors = context.theme.colors;
    final typography = context.theme.typography;
    final l10n = AppLocalizations.of(context);
    final notifier = ref.read(uploadQueueProvider.notifier);

    final activeCount = queueState.activeCount;
    final completedCount = queueState.completedCount;
    final failedCount = queueState.failedCount;

    String headerText;
    if (activeCount > 0) {
      headerText = l10n.uploadManagerUploading(activeCount);
    } else if (failedCount > 0) {
      headerText = l10n.uploadManagerFailed(failedCount);
    } else {
      headerText = l10n.uploadManagerComplete(completedCount);
    }

    return Positioned(
      left: 8,
      right: 8,
      bottom: 8,
      child: SafeArea(
        child: Material(
          elevation: 8,
          borderRadius: BorderRadius.circular(12),
          color: colors.card,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              // Header
              InkWell(
                borderRadius: const BorderRadius.vertical(
                  top: Radius.circular(12),
                ),
                onTap: notifier.toggleCollapsed,
                child: Padding(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 16,
                    vertical: 12,
                  ),
                  child: Row(
                    children: [
                      if (activeCount > 0)
                        Padding(
                          padding: const EdgeInsets.only(right: 8),
                          child: SizedBox(
                            width: 16,
                            height: 16,
                            child: CircularProgressIndicator(
                              strokeWidth: 2,
                              color: colors.primary,
                            ),
                          ),
                        ),
                      Expanded(
                        child: Text(
                          headerText,
                          style: typography.sm.copyWith(
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ),
                      if (completedCount > 0 && activeCount == 0)
                        GestureDetector(
                          onTap: notifier.dismissCompleted,
                          child: Padding(
                            padding: const EdgeInsets.only(right: 8),
                            child: Text(
                              l10n.uploadManagerClear,
                              style: typography.xs.copyWith(
                                color: colors.mutedForeground,
                              ),
                            ),
                          ),
                        ),
                      Icon(
                        queueState.isCollapsed
                            ? FIcons.chevronUp
                            : FIcons.chevronDown,
                        size: 18,
                        color: colors.mutedForeground,
                      ),
                    ],
                  ),
                ),
              ),

              // Body
              if (!queueState.isCollapsed)
                ConstrainedBox(
                  constraints: const BoxConstraints(maxHeight: 240),
                  child: ListView.separated(
                    shrinkWrap: true,
                    padding: EdgeInsets.zero,
                    itemCount: queueState.items.length,
                    separatorBuilder: (_, _) =>
                        Divider(height: 1, color: colors.border),
                    itemBuilder: (context, index) {
                      final item = queueState.items[index];
                      return _UploadItemTile(item: item, notifier: notifier);
                    },
                  ),
                ),
            ],
          ),
        ),
      ),
    );
  }
}

class _UploadItemTile extends StatelessWidget {
  final UploadItem item;
  final UploadQueueNotifier notifier;

  const _UploadItemTile({required this.item, required this.notifier});

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;
    final l10n = AppLocalizations.of(context);

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
      child: Row(
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  item.title,
                  style: typography.sm,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
                const SizedBox(height: 4),
                if (item.status == UploadItemStatus.queued)
                  Text(
                    l10n.uploadManagerWaiting,
                    style: typography.xs.copyWith(
                      color: colors.mutedForeground,
                    ),
                  ),
                if (item.status == UploadItemStatus.uploading) ...[
                  ClipRRect(
                    borderRadius: BorderRadius.circular(2),
                    child: LinearProgressIndicator(
                      value: item.progress,
                      minHeight: 4,
                      backgroundColor: colors.muted,
                      color: colors.primary,
                    ),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    '${_formatBytes(item.loaded)} / ${_formatBytes(item.total)} (${(item.progress * 100).toStringAsFixed(0)}%)',
                    style: typography.xs.copyWith(
                      color: colors.mutedForeground,
                    ),
                  ),
                ],
                if (item.status == UploadItemStatus.processing)
                  Text(
                    l10n.uploadProcessingOnServer,
                    style: typography.xs.copyWith(
                      color: colors.mutedForeground,
                    ),
                  ),
                if (item.status == UploadItemStatus.completed)
                  Text(
                    l10n.uploadComplete,
                    style: typography.xs.copyWith(color: colors.primary),
                  ),
                if (item.status == UploadItemStatus.failed)
                  Text(
                    item.error ?? l10n.uploadFailedGeneric,
                    style: typography.xs.copyWith(color: colors.destructive),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                if (item.status == UploadItemStatus.cancelled)
                  Text(
                    l10n.cancel,
                    style: typography.xs.copyWith(
                      color: colors.mutedForeground,
                    ),
                  ),
              ],
            ),
          ),
          const SizedBox(width: 8),
          if (item.status == UploadItemStatus.uploading ||
              item.status == UploadItemStatus.queued)
            IconButton(
              icon: Icon(FIcons.x, size: 16, color: colors.mutedForeground),
              onPressed: () => notifier.cancel(item.id),
              visualDensity: VisualDensity.compact,
              padding: EdgeInsets.zero,
              constraints: const BoxConstraints(minWidth: 32, minHeight: 32),
            ),
          if (item.status == UploadItemStatus.failed ||
              item.status == UploadItemStatus.cancelled) ...[
            IconButton(
              icon: Icon(FIcons.refreshCw, size: 16, color: colors.primary),
              onPressed: () => notifier.retry(item.id),
              visualDensity: VisualDensity.compact,
              padding: EdgeInsets.zero,
              constraints: const BoxConstraints(minWidth: 32, minHeight: 32),
            ),
            IconButton(
              icon: Icon(FIcons.x, size: 16, color: colors.mutedForeground),
              onPressed: () => notifier.dismiss(item.id),
              visualDensity: VisualDensity.compact,
              padding: EdgeInsets.zero,
              constraints: const BoxConstraints(minWidth: 32, minHeight: 32),
            ),
          ],
          if (item.status == UploadItemStatus.completed)
            IconButton(
              icon: Icon(FIcons.x, size: 16, color: colors.mutedForeground),
              onPressed: () => notifier.dismiss(item.id),
              visualDensity: VisualDensity.compact,
              padding: EdgeInsets.zero,
              constraints: const BoxConstraints(minWidth: 32, minHeight: 32),
            ),
        ],
      ),
    );
  }

  String _formatBytes(int bytes) {
    if (bytes < 1024) return '$bytes B';
    if (bytes < 1048576) return '${(bytes / 1024).toStringAsFixed(1)} KB';
    if (bytes < 1073741824) return '${(bytes / 1048576).toStringAsFixed(1)} MB';
    return '${(bytes / 1073741824).toStringAsFixed(1)} GB';
  }
}
