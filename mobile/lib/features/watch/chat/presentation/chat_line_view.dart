import 'dart:math';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/chat_parser/emote.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/features/user/presentation/user_avatar.dart';
import 'package:letslive_mobile/features/user/user_providers.dart';
import 'package:letslive_mobile/features/watch/chat/domain/chat_line.dart';

/// Name colour of a sender, derived from the user id (web
/// `uuidToReadableHexColor`): the id's first three bytes as RGB with the
/// lightness capped so it stays readable. Null when the id is not hex.
Color? senderColor(String userId) {
  final hex = userId.replaceAll('-', '');
  final rgb = hex.length < 6
      ? null
      : int.tryParse(hex.substring(0, 6), radix: 16);
  if (rgb == null) return null;
  final hsl = HSLColor.fromColor(Color(0xFF000000 | rgb));
  return hsl.withLightness(min(hsl.lightness, 0.7)).toColor();
}

/// One row of the chat list: a message, a join/leave note or a local notice.
class ChatLineView extends StatelessWidget {
  const ChatLineView({required this.line, super.key});

  final ChatLine line;

  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.only(bottom: 12),
    child: switch (line) {
      ChatMessageLine(:final message) => _MessageRow(
        userId: message.userId,
        username: message.username,
        picture: message.profilePicture,
        text: message.text,
      ),
      ChatMemberLine(:final userId, :final joined) => _MemberRow(
        userId: userId,
        joined: joined,
      ),
      ChatNoticeLine(:final text) => Text(
        text,
        style: Theme.of(context).textTheme.bodySmall?.copyWith(
          fontStyle: FontStyle.italic,
          color: Theme.of(context).colorScheme.onSurfaceVariant,
        ),
      ),
    },
  );
}

/// Avatar, coloured name and text. The text wraps anywhere, so a very long
/// unbroken string never overflows; `:shortcodes:` show as emoji.
class _MessageRow extends StatelessWidget {
  const _MessageRow({
    required this.userId,
    required this.username,
    required this.picture,
    required this.text,
  });

  final String userId;
  final String username;
  final String? picture;
  final String text;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        UserAvatar(pictureUrl: picture, name: username, radius: 12),
        const SizedBox(width: 8),
        Expanded(
          child: Text.rich(
            TextSpan(
              children: [
                TextSpan(
                  text: '$username: ',
                  style: TextStyle(
                    fontWeight: FontWeight.w600,
                    color: senderColor(userId) ?? theme.colorScheme.primary,
                  ),
                ),
                TextSpan(text: expandEmotes(text)),
              ],
            ),
          ),
        ),
      ],
    );
  }
}

/// "NAME joined the chat". The event carries only the id, so the name
/// comes from the cached public profile; nothing shows until it loads.
class _MemberRow extends ConsumerWidget {
  const _MemberRow({required this.userId, required this.joined});

  final String userId;
  final bool joined;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final profile = ref.watch(publicUserProvider(userId)).value;
    if (profile == null) return const SizedBox.shrink();
    final l10n = AppLocalizations.of(context);
    return _MessageRow(
      userId: userId,
      username: profile.username,
      picture: profile.profilePicture,
      text: joined ? l10n.chat_joined : l10n.chat_left,
    );
  }
}
