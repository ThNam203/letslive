import 'dart:async';

import 'package:chewie/chewie.dart';
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:forui/forui.dart';
import 'package:go_router/go_router.dart';
import 'package:video_player/video_player.dart';

import '../../../core/config/app_config.dart';
import '../../../core/network/api_endpoints.dart';
import '../../../core/network/websocket_service.dart';
import '../../../core/router/app_router.dart';
import '../../../l10n/app_localizations.dart';
import '../../../models/chat_message.dart';
import '../../../models/livestream.dart';
import '../../../models/user.dart';
import '../../../providers.dart';
import '../../../core/emotes/emote_parser.dart';
import '../../../shared/widgets/emote_picker.dart';
import '../../../shared/widgets/error_display.dart';
import '../../../shared/widgets/loading_indicator.dart';
import '../../../shared/widgets/user_avatar.dart';

class LivestreamScreen extends ConsumerStatefulWidget {
  final String livestreamId;
  final String userId;

  const LivestreamScreen({
    super.key,
    required this.livestreamId,
    required this.userId,
  });

  @override
  ConsumerState<LivestreamScreen> createState() => _LivestreamScreenState();
}

class _LivestreamScreenState extends ConsumerState<LivestreamScreen> {
  VideoPlayerController? _videoController;
  ChewieController? _chewieController;
  LiveChatService? _chatService;
  final _chatController = TextEditingController();
  final _scrollController = ScrollController();

  Livestream? _livestream;
  User? _streamer;
  List<ChatMessage> _messages = [];
  bool _isLoading = true;
  bool _isVideoLoading = true;
  String? _error;
  bool _isChatVisible = true;
  StreamSubscription<ChatMessage>? _chatSubscription;

  @override
  void initState() {
    super.initState();
    _fetchLivestreamData();
  }

  @override
  void dispose() {
    _chatSubscription?.cancel();
    _chatService?.dispose();
    _chewieController?.dispose();
    _videoController?.dispose();
    _chatController.dispose();
    _scrollController.dispose();
    SystemChrome.setPreferredOrientations([
      DeviceOrientation.portraitUp,
      DeviceOrientation.portraitDown,
      DeviceOrientation.landscapeLeft,
      DeviceOrientation.landscapeRight,
    ]);
    super.dispose();
  }

  Future<void> _fetchLivestreamData() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final repo = ref.read(livestreamRepositoryProvider);
      final response = await repo.getLivestreamOfUser(widget.userId);
      if (!mounted) return;

      if (response.success &&
          response.data != null &&
          response.data!.isNotEmpty) {
        final livestream = response.data!.first;
        setState(() {
          _livestream = livestream;
          _isLoading = false;
        });
        _initVideoPlayer(livestream);
        _initChat(livestream);
        _fetchStreamer(livestream.userId);
      } else {
        setState(() {
          _error = AppLocalizations.of(context).usersOffline;
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

  Future<void> _fetchStreamer(String userId) async {
    try {
      final userRepo = ref.read(userRepositoryProvider);
      final response = await userRepo.getUser(userId);
      if (mounted && response.success && response.data != null) {
        setState(() => _streamer = response.data);
      }
    } catch (_) {}
  }

  void _initVideoPlayer(Livestream livestream) {
    final videoUrl =
        '${AppConfig.apiUrl}${ApiEndpoints.livestreamTranscode(livestream.id)}';

    _videoController = VideoPlayerController.networkUrl(Uri.parse(videoUrl))
      ..initialize()
          .then((_) {
            if (!mounted) return;
            _chewieController = ChewieController(
              videoPlayerController: _videoController!,
              autoPlay: true,
              isLive: true,
              allowFullScreen: true,
              allowMuting: true,
              showControlsOnInitialize: false,
              errorBuilder: (context, errorMessage) {
                return Center(
                  child: Text(
                    errorMessage,
                    style: const TextStyle(color: Colors.white),
                  ),
                );
              },
            );
            setState(() => _isVideoLoading = false);
          })
          .catchError((_) {
            if (mounted) {
              setState(() => _isVideoLoading = false);
            }
          });
  }

  void _initChat(Livestream livestream) {
    final currentUser = ref.read(currentUserProvider);
    if (currentUser == null) return;

    // Fetch chat history
    final chatRepo = ref.read(chatRepositoryProvider);
    chatRepo.getMessages(roomId: livestream.id).then((response) {
      if (mounted && response.success && response.data != null) {
        setState(() => _messages = response.data!);
        _scrollToBottom();
      }
    });

    // Connect WebSocket
    _chatService = LiveChatService();
    _chatSubscription = _chatService!.messages.listen((message) {
      if (mounted) {
        setState(() => _messages.add(message));
        _scrollToBottom();
      }
    });
    _chatService!.connect(
      roomId: livestream.id,
      userId: currentUser.id,
      username: currentUser.username,
    );
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

  void _sendMessage() {
    final text = _chatController.text.trim();
    if (text.isEmpty || _chatService == null) return;

    _chatService!.sendMessage(text);
    _chatController.clear();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);

    return FScaffold(
      header: FHeader.nested(
        title: Text(_livestream?.title ?? l10n.liveStreaming),
        suffixes: [
          FButton.icon(
            onPress: () => setState(() => _isChatVisible = !_isChatVisible),
            child: Icon(
              _isChatVisible ? FIcons.messageSquare : FIcons.messageSquareOff,
            ),
          ),
        ],
      ),
      child: _isLoading
          ? LoadingIndicator(message: l10n.loading)
          : _error != null
          ? ErrorDisplay(
              title: l10n.errorGeneralTitle,
              message: _error,
              onRetry: _fetchLivestreamData,
            )
          : _buildContent(context),
    );
  }

  Widget _buildContent(BuildContext context) {
    return Column(
      children: [
        _buildVideoPlayer(context),
        if (_livestream != null) _buildStreamInfo(context),
        if (_isChatVisible) Expanded(child: _buildChat(context)),
      ],
    );
  }

  Widget _buildVideoPlayer(BuildContext context) {
    final colors = context.theme.colors;

    return AspectRatio(
      aspectRatio: 16 / 9,
      child: _isVideoLoading || _chewieController == null
          ? ColoredBox(
              color: Colors.black,
              child: Center(
                child: CircularProgressIndicator(color: colors.primary),
              ),
            )
          : Chewie(controller: _chewieController!),
    );
  }

  Widget _buildStreamInfo(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;
    final l10n = AppLocalizations.of(context);
    final livestream = _livestream!;

    return GestureDetector(
      onTap: () => context.push(AppRoutes.userProfile(livestream.userId)),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Row(
          children: [
            UserAvatar(
              imageUrl: _streamer?.profilePicture != null
                  ? '${AppConfig.apiUrl}/${_streamer!.profilePicture}'
                  : null,
              size: 36,
              fallback: const Icon(FIcons.user, size: 18),
            ),
            const SizedBox(width: 10),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    livestream.title,
                    style: typography.sm.copyWith(fontWeight: FontWeight.w600),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                  Text(
                    _streamer?.displayName ?? _streamer?.username ?? '',
                    style: typography.xs.copyWith(
                      color: colors.mutedForeground,
                    ),
                  ),
                ],
              ),
            ),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
              decoration: BoxDecoration(
                color: colors.destructive,
                borderRadius: BorderRadius.circular(4),
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(
                    FIcons.radio,
                    size: 12,
                    color: colors.destructiveForeground,
                  ),
                  const SizedBox(width: 4),
                  Text(
                    l10n.live,
                    style: typography.xs.copyWith(
                      color: colors.destructiveForeground,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(width: 8),
            Text(
              l10n.homeViewerCount(livestream.viewCount),
              style: typography.xs.copyWith(color: colors.mutedForeground),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildChat(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;
    final l10n = AppLocalizations.of(context);
    final currentUser = ref.watch(currentUserProvider);

    return Column(
      children: [
        Container(height: 1, color: colors.border),
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
          child: Row(
            children: [
              Icon(
                FIcons.messageSquare,
                size: 16,
                color: colors.mutedForeground,
              ),
              const SizedBox(width: 6),
              Text(
                l10n.usersChatTitle,
                style: typography.sm.copyWith(fontWeight: FontWeight.w600),
              ),
            ],
          ),
        ),
        Expanded(
          child: ListView.builder(
            controller: _scrollController,
            padding: const EdgeInsets.symmetric(horizontal: 12),
            itemCount: _messages.length,
            itemBuilder: (context, index) {
              final message = _messages[index];
              return _ChatBubble(message: message);
            },
          ),
        ),
        if (currentUser != null)
          _buildChatInput(context)
        else
          Padding(
            padding: const EdgeInsets.all(12),
            child: Text(
              l10n.usersChatPlaceholderLogin,
              style: typography.sm.copyWith(color: colors.mutedForeground),
              textAlign: TextAlign.center,
            ),
          ),
      ],
    );
  }

  Widget _buildChatInput(BuildContext context) {
    final colors = context.theme.colors;
    final l10n = AppLocalizations.of(context);

    return Container(
      padding: const EdgeInsets.fromLTRB(12, 8, 12, 12),
      decoration: BoxDecoration(
        border: Border(top: BorderSide(color: colors.border)),
      ),
      child: Row(
        children: [
          FButton.icon(
            onPress: () => showEmotePicker(context, (code) {
              final ctrl = _chatController;
              final sel = ctrl.selection;
              final text = ctrl.text;
              final newText =
                  text.substring(0, sel.baseOffset) +
                  code +
                  text.substring(sel.extentOffset);
              ctrl.value = TextEditingValue(
                text: newText,
                selection: TextSelection.collapsed(
                  offset: sel.baseOffset + code.length,
                ),
              );
            }),
            child: const Text('😊', style: TextStyle(fontSize: 18)),
          ),
          const SizedBox(width: 4),
          Expanded(
            child: FTextField(
              control: FTextFieldControl.managed(controller: _chatController),
              hint: l10n.usersChatPlaceholderTyping,
            ),
          ),
          const SizedBox(width: 8),
          FButton.icon(onPress: _sendMessage, child: const Icon(FIcons.send)),
        ],
      ),
    );
  }
}

class _ChatBubble extends StatelessWidget {
  final ChatMessage message;

  const _ChatBubble({required this.message});

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

    final textStyle = typography.xs.copyWith(color: colors.foreground);

    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2),
      child: Text.rich(
        TextSpan(
          children: [
            TextSpan(
              text: '${message.username}: ',
              style: typography.xs.copyWith(
                color: usernameColor,
                fontWeight: FontWeight.w600,
              ),
            ),
            ...parseEmotes(message.text, textStyle),
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
