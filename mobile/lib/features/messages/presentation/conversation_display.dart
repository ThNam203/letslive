/// How a conversation is named and dated on screen (web
/// `getConversationDisplay` and `formatTime` in `conversation-list-item.tsx`).
library;

import 'package:intl/intl.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/features/messages/data/dm_models.dart';
import 'package:letslive_mobile/features/messages/domain/dm_rules.dart';

/// Name and picture of a conversation as the signed-in user sees it.
class ConversationDisplay {
  const ConversationDisplay({
    required this.name,
    this.avatarUrl,
    this.otherUserId,
  });

  /// The other person (direct message) or the group name, with a localized
  /// fallback when there is none.
  final String name;
  final String? avatarUrl;

  /// The other member of a direct message, to show their online state.
  final String? otherUserId;
}

/// The display of [conversation] for user [myId].
ConversationDisplay conversationDisplay(
  Conversation conversation,
  String? myId,
  AppLocalizations l10n,
) {
  if (conversation.type == ConversationType.dm) {
    final other = otherParticipant(conversation, myId);
    final name = other?.username ?? '';
    return ConversationDisplay(
      name: name.isEmpty ? l10n.messages_unknown : name,
      avatarUrl: other?.profilePicture,
      otherUserId: other?.userId,
    );
  }
  final name = conversation.name ?? '';
  return ConversationDisplay(
    name: name.isEmpty ? l10n.messages_group : name,
    avatarUrl: conversation.avatarUrl,
  );
}

/// When the last message was sent, short: the time today, "<time> Yesterday",
/// the weekday within the last week, else the date.
String formatConversationTime(
  DateTime at,
  DateTime now,
  String locale,
  AppLocalizations l10n,
) {
  final date = at.toLocal();
  final today = DateTime(now.year, now.month, now.day);
  final day = DateTime(date.year, date.month, date.day);
  final time = DateFormat.jm(locale).format(date);
  if (day == today) return time;
  // Calendar days apart, not 24 h periods (a DST change makes a day 23 h).
  final daysAgo = today.difference(day).inDays;
  if (daysAgo == 1) return l10n.messages_yesterday_at(time);
  if (now.difference(date) < const Duration(days: 7)) {
    return DateFormat.E(locale).format(date);
  }
  return DateFormat.MMMd(locale).format(date);
}
