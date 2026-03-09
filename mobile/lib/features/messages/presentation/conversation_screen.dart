import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:forui/forui.dart';

import '../../../core/network/dm_websocket_service.dart';
import '../../../l10n/app_localizations.dart';
import '../../../models/conversation.dart';
import '../../../providers.dart';
import '../../../shared/widgets/error_display.dart';
import '../../../shared/widgets/loading_indicator.dart';

class ConversationScreen extends ConsumerStatefulWidget {
  final String conversationId;

  const ConversationScreen({super.key, required this.conversationId});

  @override
  ConsumerState<ConversationScreen> createState() => _ConversationScreenState();
}

class _ConversationScreenState extends ConsumerState<ConversationScreen> {
  final _messageController = TextEditingController();
  final _scrollController = ScrollController();

  Conversation? _conversation;
  List<DmMessage> _messages = [];
  bool _isLoading = true;
  bool _isLoadingMore = false;
  bool _isSending = false;
  bool _hasMore = true;
  String? _error;

  // For edit mode
  DmMessage? _editingMessage;

  // Typing indicators (remote users typing in this conversation)
  final Set<String> _typingUsernames = {};

  // WebSocket subscription
  StreamSubscription<DmServerEvent>? _wsSubscription;

  @override
  void initState() {
    super.initState();
    _fetchConversation();
    _fetchMessages();
    _markAsRead();
    _connectWebSocket();
  }

  @override
  void dispose() {
    _wsSubscription?.cancel();
    // Stop local typing indicator on leave
    final currentUser = ref.read(currentUserProvider);
    if (currentUser != null) {
      final dmWs = ref.read(dmWebSocketServiceProvider);
      dmWs.stopTyping(
        conversationId: widget.conversationId,
        username: currentUser.displayName ?? currentUser.username,
      );
    }
    _messageController.dispose();
    _scrollController.dispose();
    super.dispose();
  }

  void _connectWebSocket() {
    final dmWs = ref.read(dmWebSocketServiceProvider);
    if (!dmWs.isConnected) {
      dmWs.connect();
    }

    _wsSubscription = dmWs.events.listen(_handleWsEvent);
  }

  void _handleWsEvent(DmServerEvent event) {
    if (!mounted) return;

    switch (event) {
      case DmNewMessageEvent():
        if (event.conversationId == widget.conversationId) {
          // Avoid duplicates (message we just sent via REST)
          final exists = _messages.any((m) => m.id == event.message.id);
          if (!exists) {
            setState(() => _messages.add(event.message));
            _scrollToBottom();
            _markAsRead();
          }
        }
      case DmMessageEditedEvent():
        if (event.conversationId == widget.conversationId) {
          setState(() {
            final idx =
                _messages.indexWhere((m) => m.id == event.messageId);
            if (idx != -1) {
              final old = _messages[idx];
              _messages[idx] = DmMessage(
                id: old.id,
                conversationId: old.conversationId,
                senderId: old.senderId,
                senderUsername: old.senderUsername,
                type: old.type,
                text: event.newText,
                imageUrls: old.imageUrls,
                replyTo: old.replyTo,
                isDeleted: old.isDeleted,
                readBy: old.readBy,
                createdAt: old.createdAt,
                updatedAt: event.updatedAt,
              );
            }
          });
        }
      case DmMessageDeletedEvent():
        if (event.conversationId == widget.conversationId) {
          setState(
              () => _messages.removeWhere((m) => m.id == event.messageId));
        }
      case DmUserTypingEvent():
        if (event.conversationId == widget.conversationId) {
          final currentUser = ref.read(currentUserProvider);
          // Don't show our own typing indicator
          if (event.userId == currentUser?.id) break;

          setState(() {
            if (event.type == DmServerEventType.userTyping) {
              _typingUsernames.add(event.username);
            } else {
              _typingUsernames.remove(event.username);
            }
          });
        }
      case DmSendFailedEvent():
        _showError(event.message ?? event.key);
      default:
        break;
    }
  }

  Future<void> _fetchConversation() async {
    try {
      final repo = ref.read(messageRepositoryProvider);
      final response = await repo.getConversation(widget.conversationId);
      if (mounted && response.success && response.data != null) {
        setState(() => _conversation = response.data);
      }
    } catch (_) {}
  }

  Future<void> _fetchMessages() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final repo = ref.read(messageRepositoryProvider);
      final response =
          await repo.getConversationMessages(widget.conversationId);
      if (!mounted) return;

      if (response.success) {
        setState(() {
          _messages = response.data ?? [];
          _isLoading = false;
          _hasMore = (_messages.length >= 50);
        });
        _scrollToBottom();
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
    if (_isLoadingMore || !_hasMore || _messages.isEmpty) return;
    setState(() => _isLoadingMore = true);

    try {
      final repo = ref.read(messageRepositoryProvider);
      final oldestId = _messages.first.id;
      final response = await repo.getConversationMessages(
        widget.conversationId,
        before: oldestId,
      );
      if (!mounted) return;

      if (response.success) {
        final newMessages = response.data ?? [];
        setState(() {
          _messages.insertAll(0, newMessages);
          _isLoadingMore = false;
          _hasMore = newMessages.length >= 50;
        });
      } else {
        setState(() => _isLoadingMore = false);
      }
    } on DioException catch (_) {
      if (mounted) setState(() => _isLoadingMore = false);
    }
  }

  Future<void> _markAsRead() async {
    try {
      final repo = ref.read(messageRepositoryProvider);
      await repo.markConversationRead(widget.conversationId);
    } catch (_) {}
  }

  Future<void> _sendMessage() async {
    final text = _messageController.text.trim();
    if (text.isEmpty) return;

    final currentUser = ref.read(currentUserProvider);
    if (currentUser == null) return;

    // Stop typing indicator on send
    final dmWs = ref.read(dmWebSocketServiceProvider);
    dmWs.stopTyping(
      conversationId: widget.conversationId,
      username: currentUser.displayName ?? currentUser.username,
    );

    // If editing
    if (_editingMessage != null) {
      await _doEditMessage(_editingMessage!, text);
      return;
    }

    setState(() => _isSending = true);

    try {
      final repo = ref.read(messageRepositoryProvider);
      final response = await repo.sendMessage(
        widget.conversationId,
        text: text,
        senderUsername: currentUser.username,
      );
      if (!mounted) return;

      if (response.success && response.data != null) {
        setState(() {
          _messages.add(response.data!);
          _isSending = false;
        });
        _messageController.clear();
        _scrollToBottom();
      } else {
        setState(() => _isSending = false);
        _showError(response.message);
      }
    } on DioException catch (_) {
      if (mounted) {
        setState(() => _isSending = false);
        _showError(AppLocalizations.of(context).fetchError);
      }
    }
  }

  Future<void> _doEditMessage(DmMessage msg, String newText) async {
    setState(() => _isSending = true);

    try {
      final repo = ref.read(messageRepositoryProvider);
      final response = await repo.editMessage(
        widget.conversationId,
        msg.id,
        text: newText,
      );
      if (!mounted) return;

      if (response.success && response.data != null) {
        setState(() {
          final idx = _messages.indexWhere((m) => m.id == msg.id);
          if (idx != -1) _messages[idx] = response.data!;
          _editingMessage = null;
          _isSending = false;
        });
        _messageController.clear();
      } else {
        setState(() => _isSending = false);
        _showError(response.message);
      }
    } on DioException catch (_) {
      if (mounted) {
        setState(() => _isSending = false);
        _showError(AppLocalizations.of(context).fetchError);
      }
    }
  }

  Future<void> _deleteMessage(DmMessage msg) async {
    final l10n = AppLocalizations.of(context);
    final confirmed = await showAdaptiveDialog<bool>(
      context: context,
      builder: (context) => AlertDialog.adaptive(
        title: Text(l10n.delete),
        content: Text(l10n.commentsDeleteConfirmDescription),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(l10n.cancel),
          ),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            child: Text(l10n.delete,
                style: TextStyle(color: context.theme.colors.destructive)),
          ),
        ],
      ),
    );
    if (confirmed != true) return;

    try {
      final repo = ref.read(messageRepositoryProvider);
      final response =
          await repo.deleteMessage(widget.conversationId, msg.id);
      if (mounted && response.success) {
        setState(() => _messages.removeWhere((m) => m.id == msg.id));
      }
    } catch (_) {
      if (mounted) _showError(AppLocalizations.of(context).fetchError);
    }
  }

  void _startEdit(DmMessage msg) {
    setState(() => _editingMessage = msg);
    _messageController.text = msg.text;
  }

  void _cancelEdit() {
    setState(() => _editingMessage = null);
    _messageController.clear();
  }

  void _scrollToBottom() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (_scrollController.hasClients) {
        _scrollController.animateTo(
          _scrollController.position.maxScrollExtent,
          duration: const Duration(milliseconds: 200),
          curve: Curves.easeOut,
        );
      }
    });
  }

  void _showError(String message) {
    showFToast(
      context: context,
      title: Text(message),
      icon: const Icon(FIcons.circleAlert),
    );
  }

  void _handleInputChanged(String _) {
    final currentUser = ref.read(currentUserProvider);
    if (currentUser == null) return;

    final dmWs = ref.read(dmWebSocketServiceProvider);
    dmWs.handleTyping(
      conversationId: widget.conversationId,
      username: currentUser.displayName ?? currentUser.username,
    );
  }

  String _conversationTitle() {
    final conv = _conversation;
    if (conv == null) return '';
    if (conv.name != null && conv.name!.isNotEmpty) return conv.name!;

    final currentUser = ref.read(currentUserProvider);
    final others = conv.participants
        .where((p) => p.userId != currentUser?.id)
        .toList();
    if (others.isEmpty) return AppLocalizations.of(context).messagesUnknown;
    return others
        .map((p) => p.displayName ?? p.username)
        .join(', ');
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

    return FScaffold(
      header: FHeader.nested(
        title: Text(_conversationTitle()),
      ),
      child: _isLoading
          ? LoadingIndicator(message: l10n.loading)
          : _error != null
              ? ErrorDisplay(
                  title: l10n.errorGeneralTitle,
                  message: _error,
                  onRetry: _fetchMessages,
                )
              : _buildContent(context),
    );
  }

  Widget _buildContent(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;
    final l10n = AppLocalizations.of(context);
    final currentUser = ref.watch(currentUserProvider);

    return Column(
      children: [
        Expanded(
          child: _messages.isEmpty
              ? Center(
                  child: Text(
                    l10n.messagesNoMessagesYet,
                    style: typography.sm
                        .copyWith(color: colors.mutedForeground),
                  ),
                )
              : ListView.builder(
                  controller: _scrollController,
                  padding:
                      const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                  itemCount: _messages.length + (_hasMore ? 1 : 0),
                  itemBuilder: (context, index) {
                    if (_hasMore && index == 0) {
                      return Center(
                        child: Padding(
                          padding: const EdgeInsets.all(8),
                          child: FButton(
                            variant: FButtonVariant.ghost,
                            onPress: _isLoadingMore ? null : _loadMore,
                            child: _isLoadingMore
                                ? const SizedBox(
                                    height: 16,
                                    width: 16,
                                    child: CircularProgressIndicator(
                                        strokeWidth: 2),
                                  )
                                : Text(l10n.messagesLoadMore),
                          ),
                        ),
                      );
                    }

                    final msgIndex = _hasMore ? index - 1 : index;
                    final msg = _messages[msgIndex];
                    final isMe = msg.senderId == currentUser?.id;

                    return _MessageBubble(
                      message: msg,
                      isMe: isMe,
                      timeAgo: _formatTimeAgo(msg.createdAt, l10n),
                      onEdit: isMe && !msg.isDeleted
                          ? () => _startEdit(msg)
                          : null,
                      onDelete: isMe && !msg.isDeleted
                          ? () => _deleteMessage(msg)
                          : null,
                    );
                  },
                ),
        ),

        // Typing indicator
        if (_typingUsernames.isNotEmpty)
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
            alignment: Alignment.centerLeft,
            child: Text(
              _typingUsernames.length == 1
                  ? l10n.messagesTypingOne(_typingUsernames.first)
                  : '${_typingUsernames.join(', ')} ...',
              style: typography.xs.copyWith(
                color: colors.mutedForeground,
                fontStyle: FontStyle.italic,
              ),
            ),
          ),

        // Edit indicator
        if (_editingMessage != null)
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
            color: colors.muted,
            child: Row(
              children: [
                Icon(FIcons.pencil, size: 14, color: colors.primary),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    '${l10n.edit}: ${_editingMessage!.text}',
                    style: typography.xs.copyWith(color: colors.primary),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
                GestureDetector(
                  onTap: _cancelEdit,
                  child: Icon(FIcons.x, size: 16, color: colors.mutedForeground),
                ),
              ],
            ),
          ),

        // Message input
        Container(
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            border: Border(top: BorderSide(color: colors.border, width: 0.5)),
          ),
          child: Row(
            children: [
              Expanded(
                child: TextField(
                  controller: _messageController,
                  maxLines: 4,
                  minLines: 1,
                  onChanged: _handleInputChanged,
                  decoration: InputDecoration(
                    hintText: l10n.messagesPlaceholderTypeMessage,
                    hintStyle: typography.sm
                        .copyWith(color: colors.mutedForeground),
                    border: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(20),
                      borderSide: BorderSide(color: colors.border),
                    ),
                    enabledBorder: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(20),
                      borderSide: BorderSide(color: colors.border),
                    ),
                    contentPadding: const EdgeInsets.symmetric(
                        horizontal: 16, vertical: 10),
                    isDense: true,
                  ),
                  style: typography.sm,
                  onSubmitted: (_) => _sendMessage(),
                ),
              ),
              const SizedBox(width: 8),
              FButton.icon(
                onPress: _isSending ? null : _sendMessage,
                child: _isSending
                    ? const SizedBox(
                        height: 16,
                        width: 16,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : const Icon(FIcons.send),
              ),
            ],
          ),
        ),
      ],
    );
  }
}

class _MessageBubble extends StatelessWidget {
  final DmMessage message;
  final bool isMe;
  final String timeAgo;
  final VoidCallback? onEdit;
  final VoidCallback? onDelete;

  const _MessageBubble({
    required this.message,
    required this.isMe,
    required this.timeAgo,
    this.onEdit,
    this.onDelete,
  });

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;
    final l10n = AppLocalizations.of(context);

    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        mainAxisAlignment:
            isMe ? MainAxisAlignment.end : MainAxisAlignment.start,
        crossAxisAlignment: CrossAxisAlignment.end,
        children: [
          if (!isMe) ...[
            CircleAvatar(
              radius: 14,
              child: Text(
                (message.senderUsername.isNotEmpty
                        ? message.senderUsername[0]
                        : '?')
                    .toUpperCase(),
                style: typography.xs,
              ),
            ),
            const SizedBox(width: 6),
          ],
          Flexible(
            child: GestureDetector(
              onLongPress: (onEdit != null || onDelete != null)
                  ? () => _showActions(context)
                  : null,
              child: Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
                decoration: BoxDecoration(
                  color: message.isDeleted
                      ? colors.muted
                      : isMe
                          ? colors.primary
                          : colors.card,
                  borderRadius: BorderRadius.circular(16),
                  border: isMe
                      ? null
                      : Border.all(color: colors.border, width: 0.5),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    if (!isMe && !message.isDeleted)
                      Padding(
                        padding: const EdgeInsets.only(bottom: 2),
                        child: Text(
                          message.senderUsername,
                          style: typography.xs.copyWith(
                            fontWeight: FontWeight.w600,
                            color: colors.primary,
                          ),
                        ),
                      ),
                    Text(
                      message.isDeleted
                          ? l10n.messagesMessageDeleted
                          : message.text,
                      style: typography.sm.copyWith(
                        color: message.isDeleted
                            ? colors.mutedForeground
                            : isMe
                                ? colors.primaryForeground
                                : colors.foreground,
                        fontStyle: message.isDeleted
                            ? FontStyle.italic
                            : FontStyle.normal,
                      ),
                    ),
                    const SizedBox(height: 2),
                    Text(
                      timeAgo,
                      style: typography.xs.copyWith(
                        color: message.isDeleted
                            ? colors.mutedForeground
                            : isMe
                                ? colors.primaryForeground
                                    .withValues(alpha: 0.7)
                                : colors.mutedForeground,
                        fontSize: 10,
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  void _showActions(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final colors = context.theme.colors;

    showModalBottomSheet(
      context: context,
      builder: (context) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            if (onEdit != null)
              ListTile(
                leading: const Icon(FIcons.pencil),
                title: Text(l10n.edit),
                onTap: () {
                  Navigator.pop(context);
                  onEdit!();
                },
              ),
            if (onDelete != null)
              ListTile(
                leading: Icon(FIcons.trash, color: colors.destructive),
                title: Text(l10n.delete,
                    style: TextStyle(color: colors.destructive)),
                onTap: () {
                  Navigator.pop(context);
                  onDelete!();
                },
              ),
          ],
        ),
      ),
    );
  }
}
