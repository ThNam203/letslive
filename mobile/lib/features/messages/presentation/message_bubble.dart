import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:letslive_mobile/core/chat_parser/emote.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/features/messages/data/dm_models.dart';

/// Widest a bubble gets, as a share of the screen width (web `max-w-[70%]`).
const double _maxBubbleShare = 0.7;

/// Captions the server writes for a message that is only pictures; they are
/// shown translated (web `message-bubble.tsx`).
final RegExp _imageCaption = RegExp(r'^Sent (an image|\d+ images)$');

/// One message of a conversation (web `MessageBubble`): own messages on the
/// right, the others on the left with the sender's name at the start of a
/// run, deleted messages as a muted notice and system messages centered.
/// Text wraps anywhere, so a very long unbroken string never overflows;
/// `:shortcodes:` show as emoji.
class MessageBubble extends StatelessWidget {
  const MessageBubble({
    required this.message,
    required this.isOwn,
    required this.showSender,
    this.onLongPress,
    this.onOpenImage,
    super.key,
  });

  final DmMessage message;
  final bool isOwn;

  /// First message of a run by one sender: adds space above and, for others'
  /// messages, the sender's name.
  final bool showSender;

  /// Opens the actions of an own message (edit, delete); null when there
  /// are none.
  final VoidCallback? onLongPress;

  /// Opens the picture at a URL outside the app.
  final ValueChanged<String>? onOpenImage;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final theme = Theme.of(context);
    final scheme = theme.colorScheme;
    final top = showSender ? 12.0 : 2.0;

    if (message.type == DmMessageType.system && !message.isDeleted) {
      return Padding(
        padding: const EdgeInsets.symmetric(vertical: 8),
        child: Text(
          message.text,
          textAlign: TextAlign.center,
          style: theme.textTheme.bodySmall?.copyWith(
            color: scheme.onSurfaceVariant,
          ),
        ),
      );
    }

    final Widget content;
    final Color background;
    if (message.isDeleted) {
      background = scheme.surfaceContainerHighest.withValues(alpha: 0.6);
      content = Text(
        l10n.messages_message_deleted,
        style: theme.textTheme.bodyMedium?.copyWith(
          fontStyle: FontStyle.italic,
          color: scheme.onSurfaceVariant,
        ),
      );
    } else {
      background = isOwn ? scheme.primary : scheme.surfaceContainerHighest;
      final foreground = isOwn ? scheme.onPrimary : scheme.onSurface;
      content = Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (showSender && !isOwn)
            Text(
              message.senderUsername,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: theme.textTheme.labelSmall?.copyWith(
                color: foreground.withValues(alpha: 0.7),
                fontWeight: FontWeight.w600,
              ),
            ),
          if (message.type == DmMessageType.image &&
              message.imageUrls.isNotEmpty)
            _Images(urls: message.imageUrls, onOpen: onOpenImage),
          Text(
            _text(l10n),
            style: theme.textTheme.bodyMedium?.copyWith(color: foreground),
          ),
          Align(
            alignment: Alignment.centerRight,
            child: Wrap(
              spacing: 4,
              children: [
                Text(
                  DateFormat.jm(Localizations.localeOf(context).toString())
                      .format(message.createdAt.toLocal()),
                  style: theme.textTheme.labelSmall?.copyWith(
                    color: foreground.withValues(alpha: 0.6),
                  ),
                ),
                if (message.isEdited)
                  Text(
                    l10n.messages_edited,
                    style: theme.textTheme.labelSmall?.copyWith(
                      color: foreground.withValues(alpha: 0.5),
                    ),
                  ),
              ],
            ),
          ),
        ],
      );
    }

    return Padding(
      padding: EdgeInsets.only(top: top),
      child: Align(
        alignment: isOwn ? Alignment.centerRight : Alignment.centerLeft,
        child: LayoutBuilder(
          builder: (context, constraints) => ConstrainedBox(
            constraints: BoxConstraints(
              maxWidth: constraints.maxWidth * _maxBubbleShare,
            ),
            child: Material(
              color: background,
              borderRadius: BorderRadius.circular(12),
              child: InkWell(
                borderRadius: BorderRadius.circular(12),
                onLongPress: onLongPress,
                child: Padding(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 12,
                    vertical: 6,
                  ),
                  child: content,
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  String _text(AppLocalizations l10n) {
    final urls = message.imageUrls.length;
    if (message.type == DmMessageType.image &&
        urls > 0 &&
        _imageCaption.hasMatch(message.text)) {
      return urls == 1
          ? l10n.messages_sent_an_image
          : l10n.messages_sent_images_count(urls);
    }
    return expandEmotes(message.text);
  }
}

/// The pictures of a message, one or two per row; a tap opens one.
class _Images extends StatelessWidget {
  const _Images({required this.urls, required this.onOpen});

  final List<String> urls;
  final ValueChanged<String>? onOpen;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Padding(
      padding: const EdgeInsets.only(bottom: 4),
      child: Wrap(
        spacing: 4,
        runSpacing: 4,
        children: [
          for (var i = 0; i < urls.length; i++)
            Semantics(
              label: l10n.messages_image_alt(i + 1),
              image: true,
              child: GestureDetector(
                onTap: onOpen == null ? null : () => onOpen!(urls[i]),
                child: ClipRRect(
                  borderRadius: BorderRadius.circular(6),
                  child: CachedNetworkImage(
                    imageUrl: urls[i],
                    // Small enough for a bubble on a 320 px screen.
                    width: urls.length == 1 ? 160 : 72,
                    height: urls.length == 1 ? 128 : 72,
                    fit: BoxFit.cover,
                    errorWidget: (_, _, _) => const SizedBox.square(
                      dimension: 48,
                      child: Icon(Icons.broken_image_outlined),
                    ),
                  ),
                ),
              ),
            ),
        ],
      ),
    );
  }
}
