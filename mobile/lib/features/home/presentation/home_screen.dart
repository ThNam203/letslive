import 'package:cached_network_image/cached_network_image.dart';
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:forui/forui.dart';
import 'package:go_router/go_router.dart';

import '../../../core/config/app_config.dart';
import '../../../core/router/app_router.dart';
import '../../../l10n/app_localizations.dart';
import '../../../models/livestream.dart';
import '../../../models/user.dart';
import '../../../models/vod.dart';
import '../../../providers.dart';
import '../../../shared/widgets/empty_state_view.dart';
import '../../../shared/widgets/error_display.dart';
import '../../../shared/widgets/loading_indicator.dart';

class HomeScreen extends ConsumerStatefulWidget {
  const HomeScreen({super.key});

  @override
  ConsumerState<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends ConsumerState<HomeScreen>
    with SingleTickerProviderStateMixin {
  late final TabController _tabController;

  List<Livestream> _livestreams = [];
  List<Vod> _vods = [];
  final Map<String, User> _userCache = {};
  bool _isLoadingLivestreams = true;
  bool _isLoadingVods = true;
  String? _livestreamError;
  String? _vodError;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
    _fetchLivestreams();
    _fetchVods();
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  Future<void> _fetchLivestreams() async {
    setState(() {
      _isLoadingLivestreams = true;
      _livestreamError = null;
    });

    try {
      final repo = ref.read(livestreamRepositoryProvider);
      final response = await repo.getLivestreams();
      if (!mounted) return;

      if (response.success) {
        final streams = response.data ?? [];
        setState(() {
          _livestreams = streams;
          _isLoadingLivestreams = false;
        });
        _fetchUsersFor(streams.map((s) => s.userId).toSet());
      } else {
        setState(() {
          _livestreamError = response.message;
          _isLoadingLivestreams = false;
        });
      }
    } on DioException catch (_) {
      if (mounted) {
        setState(() {
          _livestreamError = AppLocalizations.of(context).fetchError;
          _isLoadingLivestreams = false;
        });
      }
    }
  }

  Future<void> _fetchVods() async {
    setState(() {
      _isLoadingVods = true;
      _vodError = null;
    });

    try {
      final repo = ref.read(vodRepositoryProvider);
      final response = await repo.getPopularVods();
      if (!mounted) return;

      if (response.success) {
        final vods = response.data ?? [];
        setState(() {
          _vods = vods;
          _isLoadingVods = false;
        });
        _fetchUsersFor(vods.map((v) => v.userId).toSet());
      } else {
        setState(() {
          _vodError = response.message;
          _isLoadingVods = false;
        });
      }
    } on DioException catch (_) {
      if (mounted) {
        setState(() {
          _vodError = AppLocalizations.of(context).fetchError;
          _isLoadingVods = false;
        });
      }
    }
  }

  Future<void> _fetchUsersFor(Set<String> userIds) async {
    final toFetch = userIds.where((id) => !_userCache.containsKey(id));
    if (toFetch.isEmpty) return;

    final userRepo = ref.read(userRepositoryProvider);
    for (final userId in toFetch) {
      try {
        final response = await userRepo.getUser(userId);
        if (mounted && response.success && response.data != null) {
          setState(() => _userCache[userId] = response.data!);
        }
      } catch (_) {}
    }
  }

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;
    final l10n = AppLocalizations.of(context);

    return FScaffold(
      header: FHeader(
        title: Text(l10n.appTitle),
        suffixes: [
          FButton.icon(
            onPress: () => context.push(AppRoutes.search),
            child: const Icon(FIcons.search),
          ),
        ],
      ),
      child: Column(
        children: [
          Material(
            color: colors.background,
            child: TabBar(
              controller: _tabController,
              labelColor: colors.primary,
              unselectedLabelColor: colors.mutedForeground,
              indicatorColor: colors.primary,
              labelStyle: typography.sm.copyWith(fontWeight: FontWeight.w600),
              tabs: [
                Tab(text: l10n.homeTabLivestreams),
                Tab(text: l10n.homeTabVods),
              ],
            ),
          ),
          Expanded(
            child: TabBarView(
              controller: _tabController,
              children: [
                _buildLivestreamsTab(colors, typography, l10n),
                _buildVodsTab(colors, typography, l10n),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildLivestreamsTab(
    FColors colors,
    FTypography typography,
    AppLocalizations l10n,
  ) {
    if (_isLoadingLivestreams) {
      return LoadingIndicator(message: l10n.loading);
    }

    if (_livestreamError != null) {
      return ErrorDisplay(
        title: l10n.errorGeneralTitle,
        message: _livestreamError,
        onRetry: _fetchLivestreams,
      );
    }

    if (_livestreams.isEmpty) {
      return EmptyStateView(
        icon: FIcons.video,
        title: l10n.noLivestreams,
        description: l10n.noLivestreamsDescription,
      );
    }

    return RefreshIndicator(
      onRefresh: _fetchLivestreams,
      child: ListView.builder(
        padding: const EdgeInsets.all(12),
        itemCount: _livestreams.length,
        itemBuilder: (context, index) {
          final stream = _livestreams[index];
          return _LivestreamCard(
            livestream: stream,
            user: _userCache[stream.userId],
            onTap: () => context.push(
              '${AppRoutes.livestream(stream.userId)}?livestreamId=${stream.id}',
            ),
          );
        },
      ),
    );
  }

  Widget _buildVodsTab(
    FColors colors,
    FTypography typography,
    AppLocalizations l10n,
  ) {
    if (_isLoadingVods) {
      return LoadingIndicator(message: l10n.loading);
    }

    if (_vodError != null) {
      return ErrorDisplay(
        title: l10n.errorGeneralTitle,
        message: _vodError,
        onRetry: _fetchVods,
      );
    }

    if (_vods.isEmpty) {
      return EmptyStateView(
        icon: FIcons.film,
        title: l10n.noVideos,
        description: l10n.noVideosDescription,
      );
    }

    return RefreshIndicator(
      onRefresh: _fetchVods,
      child: GridView.builder(
        padding: const EdgeInsets.all(12),
        gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
          crossAxisCount: 2,
          crossAxisSpacing: 10,
          mainAxisSpacing: 10,
          childAspectRatio: 0.75,
        ),
        itemCount: _vods.length,
        itemBuilder: (context, index) {
          final vod = _vods[index];
          return _VodCard(
            vod: vod,
            user: _userCache[vod.userId],
            onTap: () => context.push(AppRoutes.vodPlayer(vod.id)),
          );
        },
      ),
    );
  }
}

class _LivestreamCard extends StatelessWidget {
  final Livestream livestream;
  final User? user;
  final VoidCallback onTap;

  const _LivestreamCard({
    required this.livestream,
    this.user,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;
    final l10n = AppLocalizations.of(context);

    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: GestureDetector(
        onTap: onTap,
        child: DecoratedBox(
          decoration: BoxDecoration(
            color: colors.card,
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: colors.border, width: 0.5),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Thumbnail
              ClipRRect(
                borderRadius: const BorderRadius.vertical(
                  top: Radius.circular(12),
                ),
                child: AspectRatio(
                  aspectRatio: 16 / 9,
                  child: Stack(
                    fit: StackFit.expand,
                    children: [
                      if (livestream.thumbnailUrl != null)
                        CachedNetworkImage(
                          imageUrl:
                              '${AppConfig.apiUrl}/${livestream.thumbnailUrl}',
                          fit: BoxFit.cover,
                          placeholder: (_, _) => ColoredBox(
                            color: colors.muted,
                            child: const Center(
                              child: Icon(FIcons.video, size: 32),
                            ),
                          ),
                          errorWidget: (_, _, _) => ColoredBox(
                            color: colors.muted,
                            child: const Center(
                              child: Icon(FIcons.video, size: 32),
                            ),
                          ),
                        )
                      else
                        ColoredBox(
                          color: colors.muted,
                          child: const Center(
                            child: Icon(FIcons.video, size: 32),
                          ),
                        ),
                      // LIVE badge
                      if (livestream.isLive)
                        Positioned(
                          top: 8,
                          left: 8,
                          child: Container(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 8,
                              vertical: 4,
                            ),
                            decoration: BoxDecoration(
                              color: colors.destructive,
                              borderRadius: BorderRadius.circular(4),
                            ),
                            child: Text(
                              l10n.live,
                              style: typography.xs.copyWith(
                                color: colors.destructiveForeground,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                          ),
                        ),
                      // Viewer count
                      Positioned(
                        bottom: 8,
                        right: 8,
                        child: Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 8,
                            vertical: 4,
                          ),
                          decoration: BoxDecoration(
                            color: Colors.black.withValues(alpha: 0.7),
                            borderRadius: BorderRadius.circular(4),
                          ),
                          child: Text(
                            l10n.homeViewerCount(livestream.viewCount),
                            style: typography.xs.copyWith(color: Colors.white),
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
              // Info
              Padding(
                padding: const EdgeInsets.all(12),
                child: Row(
                  children: [
                    // Profile picture
                    CircleAvatar(
                      radius: 18,
                      backgroundImage: user?.profilePicture != null
                          ? CachedNetworkImageProvider(
                              '${AppConfig.apiUrl}/${user!.profilePicture}',
                            )
                          : null,
                      child: user?.profilePicture == null
                          ? const Icon(FIcons.user, size: 18)
                          : null,
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            livestream.title,
                            style: typography.sm.copyWith(
                              fontWeight: FontWeight.w600,
                            ),
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                          ),
                          const SizedBox(height: 2),
                          Text(
                            user?.displayName ?? user?.username ?? '',
                            style: typography.xs.copyWith(
                              color: colors.mutedForeground,
                            ),
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _VodCard extends StatelessWidget {
  final Vod vod;
  final User? user;
  final VoidCallback onTap;

  const _VodCard({required this.vod, this.user, required this.onTap});

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;
    final l10n = AppLocalizations.of(context);

    return GestureDetector(
      onTap: onTap,
      child: DecoratedBox(
        decoration: BoxDecoration(
          color: colors.card,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: colors.border, width: 0.5),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Thumbnail
            ClipRRect(
              borderRadius: const BorderRadius.vertical(
                top: Radius.circular(12),
              ),
              child: AspectRatio(
                aspectRatio: 16 / 9,
                child: Stack(
                  fit: StackFit.expand,
                  children: [
                    if (vod.thumbnailUrl != null)
                      CachedNetworkImage(
                        imageUrl: '${AppConfig.apiUrl}/${vod.thumbnailUrl}',
                        fit: BoxFit.cover,
                        placeholder: (_, _) => ColoredBox(
                          color: colors.muted,
                          child: const Center(
                            child: Icon(FIcons.film, size: 24),
                          ),
                        ),
                        errorWidget: (_, _, _) => ColoredBox(
                          color: colors.muted,
                          child: const Center(
                            child: Icon(FIcons.film, size: 24),
                          ),
                        ),
                      )
                    else
                      ColoredBox(
                        color: colors.muted,
                        child: const Center(child: Icon(FIcons.film, size: 24)),
                      ),
                    // Duration
                    Positioned(
                      bottom: 4,
                      right: 4,
                      child: Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 6,
                          vertical: 2,
                        ),
                        decoration: BoxDecoration(
                          color: Colors.black.withValues(alpha: 0.7),
                          borderRadius: BorderRadius.circular(4),
                        ),
                        child: Text(
                          _formatDuration(vod.duration),
                          style: typography.xs.copyWith(color: Colors.white),
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
            // Info
            Expanded(
              child: Padding(
                padding: const EdgeInsets.all(8),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      vod.title,
                      style: typography.xs.copyWith(
                        fontWeight: FontWeight.w600,
                      ),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                    const Spacer(),
                    Text(
                      user?.displayName ?? user?.username ?? '',
                      style: typography.xs.copyWith(
                        color: colors.mutedForeground,
                      ),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                    Text(
                      l10n.homeViewCount(vod.viewCount),
                      style: typography.xs.copyWith(
                        color: colors.mutedForeground,
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  String _formatDuration(int seconds) {
    final h = seconds ~/ 3600;
    final m = (seconds % 3600) ~/ 60;
    final s = seconds % 60;
    if (h > 0) {
      return '${h.toString().padLeft(2, '0')}:${m.toString().padLeft(2, '0')}:${s.toString().padLeft(2, '0')}';
    }
    return '${m.toString().padLeft(2, '0')}:${s.toString().padLeft(2, '0')}';
  }
}

