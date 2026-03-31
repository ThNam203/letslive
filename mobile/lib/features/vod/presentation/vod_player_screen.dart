import 'dart:async';

import 'package:cached_network_image/cached_network_image.dart';
import 'package:chewie/chewie.dart';
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:forui/forui.dart';
import 'package:go_router/go_router.dart';
import 'package:video_player/video_player.dart';

import '../../../core/config/app_config.dart';
import '../../../core/router/app_router.dart';
import '../../../l10n/app_localizations.dart';
import '../../../models/user.dart';
import '../../../models/vod.dart';
import '../../../providers.dart';
import '../../../shared/widgets/error_display.dart';
import '../../../shared/widgets/loading_indicator.dart';
import 'vod_comment_section.dart';

class VodPlayerScreen extends ConsumerStatefulWidget {
  final String vodId;

  const VodPlayerScreen({super.key, required this.vodId});

  @override
  ConsumerState<VodPlayerScreen> createState() => _VodPlayerScreenState();
}

class _VodPlayerScreenState extends ConsumerState<VodPlayerScreen> {
  VideoPlayerController? _videoController;
  ChewieController? _chewieController;

  Vod? _vod;
  User? _vodOwner;
  bool _isLoading = true;
  bool _isVideoLoading = true;
  String? _error;

  bool _viewRegistered = false;
  Timer? _watchTimer;

  @override
  void initState() {
    super.initState();
    _fetchVod();
  }

  @override
  void dispose() {
    _watchTimer?.cancel();
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

  Future<void> _fetchVod() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final repo = ref.read(vodRepositoryProvider);
      final response = await repo.getVod(widget.vodId);
      if (!mounted) return;

      if (response.success && response.data != null) {
        setState(() {
          _vod = response.data;
          _isLoading = false;
        });
        _initVideoPlayer(response.data!);
        _fetchVodOwner(response.data!.userId);
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

  Future<void> _fetchVodOwner(String userId) async {
    try {
      final userRepo = ref.read(userRepositoryProvider);
      final response = await userRepo.getUser(userId);
      if (mounted && response.success && response.data != null) {
        setState(() => _vodOwner = response.data);
      }
    } catch (_) {}
  }

  void _initVideoPlayer(Vod vod) {
    final videoUrl = '${AppConfig.apiUrl}/${vod.playbackUrl}';

    _videoController = VideoPlayerController.networkUrl(Uri.parse(videoUrl))
      ..initialize()
          .then((_) {
            if (!mounted) return;
            _chewieController = ChewieController(
              videoPlayerController: _videoController!,
              autoPlay: true,
              allowFullScreen: true,
              allowMuting: true,
              allowPlaybackSpeedChanging: true,
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
            _startWatchTracking();
          })
          .catchError((_) {
            if (mounted) {
              setState(() => _isVideoLoading = false);
            }
          });
  }

  void _startWatchTracking() {
    _watchTimer = Timer.periodic(const Duration(seconds: 1), (_) {
      if (_viewRegistered || !mounted) return;

      final controller = _videoController;
      if (controller == null || !controller.value.isPlaying) return;

      final position = controller.value.position.inSeconds;
      final duration = _vod?.duration ?? 0;

      // Threshold: at least 15 seconds OR 10% of video duration (whichever is smaller)
      int threshold = 15;
      if (duration > 0) {
        final tenPercent = (duration * 0.1).ceil();
        if (tenPercent < threshold) threshold = tenPercent;
      }
      if (threshold < 1) threshold = 1;

      if (position >= threshold) {
        _registerView(position);
      }
    });
  }

  Future<void> _registerView(int watchedSeconds) async {
    if (_viewRegistered) return;
    _viewRegistered = true;
    _watchTimer?.cancel();

    try {
      final repo = ref.read(vodRepositoryProvider);
      await repo.registerView(
        vodId: widget.vodId,
        watchedSeconds: watchedSeconds,
      );
    } catch (_) {
      // View registration is best-effort; don't disrupt playback
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);

    return FScaffold(
      header: FHeader.nested(title: Text(_vod?.title ?? l10n.videos)),
      child: _isLoading
          ? LoadingIndicator(message: l10n.loading)
          : _error != null
          ? ErrorDisplay(
              title: l10n.errorGeneralTitle,
              message: _error,
              onRetry: _fetchVod,
            )
          : _buildContent(context),
    );
  }

  Widget _buildContent(BuildContext context) {
    final vod = _vod;
    if (vod == null) return const SizedBox.shrink();

    return SingleChildScrollView(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _buildVideoPlayer(context),
          _buildVodInfo(context, vod),
          VodCommentSection(vodId: vod.id, vodOwnerId: vod.userId),
        ],
      ),
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

  Widget _buildVodInfo(BuildContext context, Vod vod) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;
    final l10n = AppLocalizations.of(context);
    final owner = _vodOwner;
    final ownerName = owner?.displayName ?? owner?.username ?? '';

    return Padding(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            vod.title,
            style: typography.lg.copyWith(fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 8),
          Row(
            children: [
              Icon(FIcons.eye, size: 14, color: colors.mutedForeground),
              const SizedBox(width: 4),
              Text(
                l10n.homeViewCount(vod.viewCount),
                style: typography.xs.copyWith(color: colors.mutedForeground),
              ),
            ],
          ),
          if (vod.description != null && vod.description!.isNotEmpty) ...[
            const SizedBox(height: 12),
            Text(
              vod.description!,
              style: typography.sm.copyWith(color: colors.foreground),
            ),
          ],
          const SizedBox(height: 16),
          // Streamer info
          GestureDetector(
            onTap: () => context.push(AppRoutes.userProfile(vod.userId)),
            child: Row(
              children: [
                CircleAvatar(
                  radius: 16,
                  backgroundImage: owner?.profilePicture != null
                      ? CachedNetworkImageProvider(
                          '${AppConfig.apiUrl}/${owner!.profilePicture}',
                        )
                      : null,
                  child: owner?.profilePicture == null
                      ? const Icon(FIcons.user, size: 16)
                      : null,
                ),
                const SizedBox(width: 8),
                Text(
                  ownerName,
                  style: typography.sm.copyWith(fontWeight: FontWeight.w500),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
