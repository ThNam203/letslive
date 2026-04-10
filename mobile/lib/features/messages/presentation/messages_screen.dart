import 'package:cached_network_image/cached_network_image.dart';
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:forui/forui.dart';
import 'package:go_router/go_router.dart';

import '../../../core/config/app_config.dart';
import '../../../core/router/app_router.dart';
import '../../../l10n/app_localizations.dart';
import '../../../models/conversation.dart';
import '../../../providers.dart';
import '../../../shared/widgets/empty_state_view.dart';
import '../../../shared/widgets/error_display.dart';
import '../../../shared/widgets/loading_indicator.dart';
import 'new_conversation_dialog.dart';

class MessagesScreen extends ConsumerStatefulWidget {
  const MessagesScreen({super.key});

  @override
  ConsumerState<MessagesScreen> createState() => _MessagesScreenState();
}

class _MessagesScreenState extends ConsumerState<MessagesScreen> {
  List<Conversation> _conversations = [];
  bool _isLoading = true;
  bool _isLoadingMore = false;
  String? _error;
  int _currentPage = 0;
  bool _hasMore = true;

  @override
  void initState() {
    super.initState();
    _fetchConversations();
  }

  Future<void> _fetchConversations() async {
    final currentUser = ref.read(currentUserProvider);
    if (currentUser == null) {
      setState(() => _isLoading = false);
      return;
    }

    setState(() {
      _isLoading = true;
      _error = null;
      _currentPage = 0;
    });

    try {
      final repo = ref.read(messageRepositoryProvider);
      final response = await repo.getConversations(page: 0);
      if (!mounted) return;

      if (response.success) {
        setState(() {
          _conversations = response.data ?? [];
          _isLoading = false;
          _hasMore = (_conversations.length >= 20);
        });
      } else {
        setState(() {
          _error = response.message;
          _isLoading = false;
        });
      }
    } on DioException catch (_) {
      if (mounted) {
        setState(() {
          _error = AppLocalizations.of(context).fetchError;
          _isLoading = false;
        });
      }
    }
  }

  Future<void> _loadMore() async {
    if (_isLoadingMore || !_hasMore) return;
    setState(() => _isLoadingMore = true);

    try {
      final repo = ref.read(messageRepositoryProvider);
      final nextPage = _currentPage + 1;
      final response = await repo.getConversations(page: nextPage);
      if (!mounted) return;

      if (response.success) {
        final newConversations = response.data ?? [];
        setState(() {
          _currentPage = nextPage;
          _conversations.addAll(newConversations);
          _isLoadingMore = false;
          _hasMore = newConversations.length >= 20;
        });
      } else {
        setState(() => _isLoadingMore = false);
      }
    } on DioException catch (_) {
      if (mounted) setState(() => _isLoadingMore = false);
    }
  }

  void _openNewConversation() {
    showDialog(
      context: context,
      builder: (context) => NewConversationDialog(
        onCreated: (conversationId) {
          _fetchConversations();
          context.push(AppRoutes.conversation(conversationId));
        },
      ),
    );
  }

  String _formatTimeAgo(String createdAt, AppLocalizations l10n) {
    final date = DateTime.tryParse(createdAt);
    if (date == null) return '';
    final now = DateTime.now();
    final diff = now.difference(date);
    if (diff.inSeconds < 60) return l10n.timeSecondsAgo(diff.inSeconds);
    if (diff.inMinutes < 60) return l10n.timeMinutesAgo(diff.inMinutes);
    if (diff.inHours < 24) return l10n.timeHoursAgo(diff.inHours);
    if (diff.inDays < 7) return l10n.timeDaysAgo(diff.inDays);
    if (diff.inDays < 30) return l10n.timeWeeksAgo(diff.inDays ~/ 7);
    if (diff.inDays < 365) return l10n.timeMonthsAgo(diff.inDays ~/ 30);
    return l10n.timeYearsAgo(diff.inDays ~/ 365);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final currentUser = ref.watch(currentUserProvider);

    return FScaffold(
      header: FHeader(
        title: Text(l10n.messagesTitle),
        suffixes: [
          if (currentUser != null)
            FButton.icon(
              onPress: _openNewConversation,
              child: const Icon(FIcons.squarePen),
            ),
        ],
      ),
      child: currentUser == null
          ? _buildLoginPrompt(context, l10n)
          : _buildContent(context, l10n, currentUser.id),
    );
  }

  Widget _buildLoginPrompt(BuildContext context, AppLocalizations l10n) {
    return EmptyStateView(
      icon: FIcons.messageCircle,
      title: l10n.messagesLoginRequired,
    );
  }

  Widget _buildContent(
    BuildContext context,
    AppLocalizations l10n,
    String currentUserId,
  ) {
    if (_isLoading) {
      return LoadingIndicator(message: l10n.loading);
    }

    if (_error != null) {
      return ErrorDisplay(
        title: l10n.errorGeneralTitle,
        message: _error,
        onRetry: _fetchConversations,
      );
    }

    if (_conversations.isEmpty) {
      return EmptyStateView(
        icon: FIcons.messageCircle,
        title: l10n.messagesNoConversationsYet,
      );
    }

    return RefreshIndicator(
      onRefresh: _fetchConversations,
      child: ListView.builder(
        padding: const EdgeInsets.symmetric(vertical: 4),
        itemCount: _conversations.length + (_hasMore ? 1 : 0),
        itemBuilder: (context, index) {
          if (index == _conversations.length) {
            return Padding(
              padding: const EdgeInsets.all(16),
              child: FButton(
                variant: FButtonVariant.outline,
                onPress: _isLoadingMore ? null : _loadMore,
                child: _isLoadingMore
                    ? const SizedBox(
                        height: 16,
                        width: 16,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : Text(l10n.messagesLoadMore),
              ),
            );
          }

          final conv = _conversations[index];
          return _ConversationTile(
            conversation: conv,
            currentUserId: currentUserId,
            timeAgo: conv.lastMessage != null
                ? _formatTimeAgo(conv.lastMessage!.createdAt, l10n)
                : '',
            onTap: () => context.push(AppRoutes.conversation(conv.id)),
          );
        },
      ),
    );
  }
}

class _ConversationTile extends StatelessWidget {
  final Conversation conversation;
  final String currentUserId;
  final String timeAgo;
  final VoidCallback onTap;

  const _ConversationTile({
    required this.conversation,
    required this.currentUserId,
    required this.timeAgo,
    required this.onTap,
  });

  String _displayName(AppLocalizations l10n) {
    if (conversation.name != null && conversation.name!.isNotEmpty) {
      return conversation.name!;
    }
    final others = conversation.participants
        .where((p) => p.userId != currentUserId)
        .toList();
    if (others.isEmpty) return l10n.messagesUnknown;
    return others.map((p) => p.displayName ?? p.username).join(', ');
  }

  String? _avatarUrl() {
    if (conversation.avatarUrl != null) return conversation.avatarUrl;
    final others = conversation.participants
        .where((p) => p.userId != currentUserId)
        .toList();
    if (others.length == 1 && others.first.profilePicture != null) {
      return others.first.profilePicture;
    }
    return null;
  }

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;
    final l10n = AppLocalizations.of(context);
    final avatar = _avatarUrl();

    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        decoration: BoxDecoration(
          border: Border(bottom: BorderSide(color: colors.border, width: 0.5)),
        ),
        child: Row(
          children: [
            CircleAvatar(
              radius: 24,
              backgroundImage: avatar != null
                  ? CachedNetworkImageProvider('${AppConfig.apiUrl}/$avatar')
                  : null,
              child: avatar == null
                  ? Icon(
                      conversation.type == ConversationType.group
                          ? FIcons.users
                          : FIcons.user,
                      size: 20,
                    )
                  : null,
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
                          _displayName(l10n),
                          style: typography.sm.copyWith(
                            fontWeight: FontWeight.w600,
                          ),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
                      if (timeAgo.isNotEmpty)
                        Text(
                          timeAgo,
                          style: typography.xs.copyWith(
                            color: colors.mutedForeground,
                          ),
                        ),
                    ],
                  ),
                  if (conversation.lastMessage != null) ...[
                    const SizedBox(height: 2),
                    Text(
                      '${conversation.lastMessage!.senderUsername}: ${conversation.lastMessage!.text}',
                      style: typography.xs.copyWith(
                        color: colors.mutedForeground,
                      ),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ],
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
