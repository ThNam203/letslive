import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/features/user/presentation/user_avatar.dart';
import 'package:letslive_mobile/features/user/user_providers.dart';

/// A live stream or video in the feed: 16:9 thumbnail, title, owner and a
/// meta line. Tappable when [onTap] is set (live streams open the channel;
/// videos wait for the VOD screen).
///
/// The owner is looked up by [userId] ([publicUserProvider], one request per
/// id); until it arrives, or if it fails, the card shows a neutral avatar and
/// no name. Texts are clipped so very long titles and names never overflow.
class MediaCard extends ConsumerWidget {
  const MediaCard({
    required this.thumbnailUrl,
    required this.title,
    required this.userId,
    required this.meta,
    this.durationLabel,
    this.onTap,
    super.key,
  });

  /// Absolute URL of the 16:9 thumbnail.
  final String thumbnailUrl;

  /// Shown on at most two lines.
  final String title;

  /// Id of the owner to look up for the avatar and name.
  final String userId;

  /// Under the owner name, e.g. "1.2K views · 5 minutes ago".
  final String meta;

  /// Length badge on the thumbnail (videos only).
  final String? durationLabel;

  /// Called when the card is tapped; null leaves it inert.
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final theme = Theme.of(context);
    final owner = ref.watch(publicUserProvider(userId)).value;
    final card = Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        MediaThumbnail(url: thumbnailUrl, durationLabel: durationLabel),
        Padding(
          padding: const EdgeInsets.symmetric(vertical: 8),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              UserAvatar(
                pictureUrl: owner?.profilePicture,
                name: owner?.username,
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      title,
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                      style: theme.textTheme.titleMedium,
                    ),
                    if (owner != null)
                      Text(
                        owner.username,
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: theme.textTheme.bodyMedium?.copyWith(
                          color: theme.colorScheme.onSurfaceVariant,
                        ),
                      ),
                    Text(
                      meta,
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
      ],
    );
    return onTap == null ? card : InkWell(onTap: onTap, child: card);
  }
}

/// A 16:9 thumbnail with an optional length badge (bottom right) and an
/// optional "live" badge (top left). Shows a plain icon while the image loads
/// and when it cannot (web falls back to a bundled image; an icon keeps the
/// app free of extra assets).
class MediaThumbnail extends StatelessWidget {
  const MediaThumbnail({
    required this.url,
    this.durationLabel,
    this.liveLabel,
    super.key,
  });

  /// Absolute URL of the image.
  final String url;

  /// Length badge (videos only).
  final String? durationLabel;

  /// Red badge for streams that are live now.
  final String? liveLabel;

  @override
  Widget build(BuildContext context) {
    final labelStyle = Theme.of(context).textTheme.labelSmall
        ?.copyWith(color: Colors.white);
    final duration = durationLabel;
    final live = liveLabel;
    return AspectRatio(
      aspectRatio: 16 / 9,
      child: Stack(
        fit: StackFit.expand,
        children: [
          CachedNetworkImage(
            imageUrl: url,
            fit: BoxFit.cover,
            placeholder: (_, _) => const _ThumbnailPlaceholder(),
            errorWidget: (_, _, _) => const _ThumbnailPlaceholder(),
          ),
          if (duration != null)
            Positioned(
              right: 8,
              bottom: 8,
              child: _Badge(duration, Colors.black87, labelStyle),
            ),
          if (live != null)
            Positioned(
              left: 8,
              top: 8,
              child: _Badge(live, Colors.red.shade600, labelStyle),
            ),
        ],
      ),
    );
  }
}

class _Badge extends StatelessWidget {
  const _Badge(this.text, this.color, this.style);

  final String text;
  final Color color;
  final TextStyle? style;

  @override
  Widget build(BuildContext context) => DecoratedBox(
    decoration: BoxDecoration(
      color: color,
      borderRadius: BorderRadius.circular(4),
    ),
    child: Padding(
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
      child: Text(text, style: style),
    ),
  );
}

class _ThumbnailPlaceholder extends StatelessWidget {
  const _ThumbnailPlaceholder();

  @override
  Widget build(BuildContext context) => ColoredBox(
    color: Theme.of(context).colorScheme.surfaceContainerHighest,
    child: const Center(child: Icon(Icons.play_circle_outline, size: 40)),
  );
}
