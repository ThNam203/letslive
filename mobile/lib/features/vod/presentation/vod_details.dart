import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/router/app_routes.dart';
import 'package:letslive_mobile/features/home/data/vod.dart';
import 'package:letslive_mobile/features/home/domain/media_format.dart';
import 'package:letslive_mobile/features/user/presentation/follow_button.dart';
import 'package:letslive_mobile/features/user/presentation/user_avatar.dart';
import 'package:letslive_mobile/features/user/user_providers.dart';
import 'package:letslive_mobile/features/vod/presentation/timestamped_text.dart';
import 'package:letslive_mobile/features/vod/presentation/vod_reaction_buttons.dart';

/// Lines of the description shown before "...more" (web `line-clamp-3`).
const int _collapsedDescriptionLines = 3;

/// Title, uploader, like/dislike, a copy-link button and the description box
/// under the player (web `VODDetails`). The uploader is [ownerId] from the
/// route, as on web, with a follow button. The link
/// is the web page of the video (`WEB_BASE_URL`); without that setting (dev
/// only) the button is hidden. Web's "start at the current time" switch is
/// not ported.
class VodDetails extends ConsumerWidget {
  const VodDetails({required this.vod, required this.ownerId, super.key});

  /// The video to describe.
  final Vod vod;

  /// The user of the route `/users/:userId/vods/:vodId`.
  final String ownerId;

  /// Copies the video's web link and says whether that worked.
  Future<void> _copyLink(BuildContext context, String webBaseUrl) async {
    final l10n = AppLocalizations.of(context);
    final messenger = ScaffoldMessenger.of(context);
    String message;
    try {
      await Clipboard.setData(
        ClipboardData(text: '$webBaseUrl${AppRoutes.vod(vod.userId, vod.id)}'),
      );
      message = l10n.common_vod_link_copied;
    } on PlatformException {
      message = l10n.common_vod_copy_failed;
    }
    messenger.showSnackBar(SnackBar(content: Text(message)));
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final theme = Theme.of(context);
    final owner = ref.watch(publicUserProvider(ownerId)).value;
    final webBaseUrl = ref.watch(appConfigProvider).webBaseUrl;
    final description = vod.description?.trim() ?? '';
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(vod.title, style: theme.textTheme.titleLarge),
          const SizedBox(height: 12),
          Wrap(
            spacing: 12,
            runSpacing: 8,
            crossAxisAlignment: WrapCrossAlignment.center,
            alignment: WrapAlignment.spaceBetween,
            children: [
              InkWell(
                onTap: () => context.push(AppRoutes.channel(ownerId)),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    UserAvatar(
                      pictureUrl: owner?.profilePicture,
                      name: owner?.username,
                    ),
                    const SizedBox(width: 12),
                    if (owner != null)
                      ConstrainedBox(
                        constraints: BoxConstraints(
                          maxWidth: MediaQuery.sizeOf(context).width * 0.5,
                        ),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              owner.username,
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                              style: theme.textTheme.titleSmall,
                            ),
                            Text(
                              l10n.common_followers_count(
                                compactCount(l10n, owner.followerCount),
                                owner.followerCount,
                              ),
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                              style: theme.textTheme.bodySmall?.copyWith(
                                color: theme.colorScheme.onSurfaceVariant,
                              ),
                            ),
                          ],
                        ),
                      ),
                  ],
                ),
              ),
              FollowButton(userId: ownerId),
              Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  VodReactionButtons(vod: vod),
                  if (webBaseUrl != null)
                    IconButton(
                      key: const ValueKey('vod-copy-link'),
                      tooltip: l10n.common_vod_share,
                      icon: const Icon(Icons.share_outlined),
                      onPressed: () => _copyLink(context, webBaseUrl),
                    ),
                ],
              ),
            ],
          ),
          const SizedBox(height: 12),
          DecoratedBox(
            decoration: BoxDecoration(
              color: theme.colorScheme.surfaceContainerHighest,
              borderRadius: BorderRadius.circular(8),
            ),
            child: Padding(
              padding: const EdgeInsets.all(12),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    '${l10n.common_vod_views(vod.viewCount)} · '
                    '${timeAgo(l10n, vod.createdAt, DateTime.now())}',
                    style: theme.textTheme.titleSmall,
                  ),
                  if (description.isNotEmpty)
                    _Description(description: description),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

/// The description cut at three lines with a "...more" / "Show less"
/// toggle that only shows when text is really cut off.
class _Description extends StatefulWidget {
  const _Description({required this.description});

  final String description;

  @override
  State<_Description> createState() => _DescriptionState();
}

class _DescriptionState extends State<_Description> {
  bool _expanded = false;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final style = Theme.of(context).textTheme.bodyMedium;
    return LayoutBuilder(
      builder: (context, box) {
        final painter = TextPainter(
          text: TextSpan(text: widget.description, style: style),
          maxLines: _collapsedDescriptionLines,
          textDirection: Directionality.of(context),
          textScaler: MediaQuery.textScalerOf(context),
        )..layout(maxWidth: box.maxWidth);
        final overflows = painter.didExceedMaxLines;
        painter.dispose();
        return Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const SizedBox(height: 4),
            TimestampedText(
              widget.description,
              style: style,
              maxLines: _expanded ? null : _collapsedDescriptionLines,
            ),
            if (overflows || _expanded)
              TextButton(
                onPressed: () => setState(() => _expanded = !_expanded),
                child: Text(
                  _expanded
                      ? l10n.common_vod_show_less
                      : l10n.common_vod_show_more,
                ),
              ),
          ],
        );
      },
    );
  }
}
