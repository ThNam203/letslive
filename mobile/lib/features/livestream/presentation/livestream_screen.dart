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
import '../../../core/router/app_router.dart';
import '../../../l10n/app_localizations.dart';
import '../../../models/livestream.dart';
import '../../../models/user.dart';
import '../../../providers.dart';
import '../../../shared/widgets/error_display.dart';
import '../../../shared/widgets/loading_indicator.dart';
import '../../../shared/widgets/user_avatar.dart';
import 'widgets/live_chat_panel.dart';

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

  Livestream? _livestream;
  User? _streamer;
  bool _isLoading = true;
  bool _isVideoLoading = true;
  String? _error;
  bool _isChatVisible = true;

  @override
  void initState() {
    super.initState();
    _fetchLivestreamData();
  }

  @override
  void dispose() {
    _chewieController?.dispose();
    _videoController?.dispose();
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
    final livestream = _livestream;
    return Column(
      children: [
        _buildVideoPlayer(context),
        if (livestream != null)
          _StreamInfoBar(livestream: livestream, streamer: _streamer),
        if (_isChatVisible && livestream != null)
          Expanded(child: LiveChatPanel(roomId: livestream.id)),
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
}

class _StreamInfoBar extends StatelessWidget {
  final Livestream livestream;
  final User? streamer;

  const _StreamInfoBar({required this.livestream, required this.streamer});

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;
    final l10n = AppLocalizations.of(context);

    return GestureDetector(
      onTap: () => context.push(AppRoutes.userProfile(livestream.userId)),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Row(
          children: [
            UserAvatar(
              imageUrl: streamer?.profilePicture != null
                  ? '${AppConfig.apiUrl}/${streamer!.profilePicture}'
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
                    streamer?.displayName ?? streamer?.username ?? '',
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
}
