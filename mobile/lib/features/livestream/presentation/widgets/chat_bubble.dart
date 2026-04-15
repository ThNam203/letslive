import 'package:flutter/material.dart';
import 'package:forui/forui.dart';

import '../../../../core/chat_parser/emote.dart';
import '../../../../l10n/app_localizations.dart';
import '../../../../models/chat_message.dart';

/// Renders one regular / join / leave chat message. Italicizes `_text_`
/// action-style messages (sent by `/me`).
class ChatBubble extends StatelessWidget {
  final ChatMessage message;

  const ChatBubble({super.key, required this.message});

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;
    final l10n = AppLocalizations.of(context);

    if (message.isJoin || message.isLeave) {
      return Padding(
        padding: const EdgeInsets.symmetric(vertical: 4),
        child: Center(
          child: Text(
            '${message.username} ${message.isJoin ? l10n.usersChatJoined : l10n.usersChatLeft}',
            style: typography.xs.copyWith(
              color: colors.mutedForeground,
              fontStyle: FontStyle.italic,
            ),
          ),
        ),
      );
    }

    final usernameColor = _colorFromUserId(message.userId);

    final isAction =
        message.text.length >= 2 &&
        message.text.startsWith('_') &&
        message.text.endsWith('_');
    final displayText = isAction
        ? message.text.substring(1, message.text.length - 1)
        : message.text;

    final textStyle = typography.xs.copyWith(
      color: colors.foreground,
      fontStyle: isAction ? FontStyle.italic : FontStyle.normal,
    );

    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2),
      child: Text.rich(
        TextSpan(
          children: [
            TextSpan(
              text:
                  isAction ? '${message.username} ' : '${message.username}: ',
              style: typography.xs.copyWith(
                color: usernameColor,
                fontWeight: FontWeight.w600,
              ),
            ),
            ...parseEmotes(displayText, textStyle),
          ],
        ),
      ),
    );
  }

  Color _colorFromUserId(String userId) {
    final hash = userId.hashCode;
    final hue = (hash % 360).abs().toDouble();
    return HSLColor.fromAHSL(1.0, hue, 0.7, 0.5).toColor();
  }
}

/// A locally-rendered system line (e.g. `/help` output, error messages).
class SystemLineWidget extends StatelessWidget {
  final String text;

  const SystemLineWidget({super.key, required this.text});

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Text(
        text,
        style: typography.xs.copyWith(
          color: colors.mutedForeground,
          fontStyle: FontStyle.italic,
        ),
      ),
    );
  }
}
