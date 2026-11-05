/// The Messages tab: the user's conversations, newest activity first (web
/// `/messages`, the list half). Each row shows the other person or group, the
/// last message with its time, an online dot and the unread count.
library;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/router/app_routes.dart';
import 'package:letslive_mobile/features/home/presentation/feed_section.dart';
import 'package:letslive_mobile/features/messages/data/dm_models.dart';
import 'package:letslive_mobile/features/messages/messages_providers.dart';
import 'package:letslive_mobile/features/messages/presentation/conversation_display.dart';
import 'package:letslive_mobile/features/user/presentation/user_avatar.dart';

/// Unread counts above this show as "99+" (web).
const int _maxShownUnread = 99;

/// The Messages tab screen.
class ConversationsScreen extends ConsumerWidget {
  const ConversationsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final conversations = ref.watch(conversationsProvider);
    final feed = conversations.value;
    final status = feedStatusBody(
      value: conversations.whenData((feed) => feed.items),
      onRetry: () => ref.invalidate(conversationsProvider),
      empty: EmptyNotice(
        icon: Icons.chat_bubble_outline,
        title: l10n.messages_no_conversations_yet,
      ),
    );
    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.messages_title),
        actions: [
          IconButton(
            key: const ValueKey('new-conversation'),
            tooltip: l10n.messages_new_conversation,
            icon: const Icon(Icons.add),
            onPressed: () => context.push(AppRoutes.newConversation),
          ),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: () => ref.refresh(conversationsProvider.future),
        child: status != null
            ? ListView(children: [status])
            : ListView.builder(
                itemCount: feed!.items.length + 1,
                itemBuilder: (context, index) => index == feed.items.length
                    ? PagedFooter<Conversation>(
                        feed: feed,
                        onLoadMore: () =>
                            ref.read(conversationsProvider.notifier).loadMore(),
                      )
                    : ConversationTile(
                        key: ValueKey('conversation-${feed.items[index].id}'),
                        conversation: feed.items[index],
                      ),
              ),
      ),
    );
  }
}

/// One row of the list. Text is clipped with an ellipsis so a long name or
/// message never overflows, even on a narrow screen with large text.
class ConversationTile extends ConsumerWidget {
  const ConversationTile({required this.conversation, super.key});

  final Conversation conversation;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final theme = Theme.of(context);
    final display = conversationDisplay(
      conversation,
      ref.watch(currentUserIdProvider),
      l10n,
    );
    final otherId = display.otherUserId;
    final isOnline =
        otherId != null &&
        ref.watch(dmPresenceProvider.select((p) => p.online.contains(otherId)));
    final unread =
        ref.watch(dmUnreadProvider.select((u) => u.value?[conversation.id])) ??
        0;
    final last = conversation.lastMessage;
    final muted = theme.colorScheme.onSurfaceVariant;
    return InkWell(
      onTap: () => context.push(AppRoutes.conversation(conversation.id)),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        child: Row(
          children: [
            AvatarWithDot(
              pictureUrl: display.avatarUrl,
              name: display.name,
              isOnline: isOnline,
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Expanded(
                        child: Text(
                          display.name,
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: theme.textTheme.titleSmall,
                        ),
                      ),
                      if (last != null) ...[
                        const SizedBox(width: 8),
                        Flexible(
                          child: Text(
                            formatConversationTime(
                              last.createdAt,
                              DateTime.now(),
                              Localizations.localeOf(context).toString(),
                              l10n,
                            ),
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: theme.textTheme.bodySmall?.copyWith(
                              color: muted,
                            ),
                          ),
                        ),
                      ],
                    ],
                  ),
                  Row(
                    children: [
                      Expanded(
                        child: Text(
                          last?.text ?? l10n.messages_no_messages_yet,
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: theme.textTheme.bodySmall?.copyWith(
                            color: muted,
                          ),
                        ),
                      ),
                      if (unread > 0) ...[
                        const SizedBox(width: 8),
                        _UnreadPill(count: unread),
                      ],
                    ],
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

/// A round avatar with a green dot while the person is online.
class AvatarWithDot extends StatelessWidget {
  const AvatarWithDot({
    required this.name,
    required this.isOnline,
    this.pictureUrl,
    this.radius = 20,
    super.key,
  });

  final String name;
  final String? pictureUrl;
  final bool isOnline;
  final double radius;

  @override
  Widget build(BuildContext context) => Stack(
    children: [
      UserAvatar(pictureUrl: pictureUrl, name: name, radius: radius),
      if (isOnline)
        Positioned(
          right: 0,
          bottom: 0,
          child: Container(
            key: const ValueKey('online-dot'),
            width: 12,
            height: 12,
            decoration: BoxDecoration(
              color: Colors.green,
              shape: BoxShape.circle,
              border: Border.all(
                color: Theme.of(context).colorScheme.surface,
                width: 2,
              ),
            ),
          ),
        ),
    ],
  );
}

class _UnreadPill extends StatelessWidget {
  const _UnreadPill({required this.count});

  final int count;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    return Container(
      constraints: const BoxConstraints(minWidth: 20, minHeight: 20),
      padding: const EdgeInsets.symmetric(horizontal: 6),
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: scheme.primary,
        borderRadius: BorderRadius.circular(10),
      ),
      child: Text(
        count > _maxShownUnread ? '$_maxShownUnread+' : '$count',
        style: TextStyle(color: scheme.onPrimary, fontSize: 12),
      ),
    );
  }
}
