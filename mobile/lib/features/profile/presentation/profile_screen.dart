import 'package:cached_network_image/cached_network_image.dart';
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:forui/forui.dart';

import '../../../core/config/app_config.dart';
import '../../../l10n/app_localizations.dart';
import '../../../models/user.dart';
import '../../../models/vod.dart';
import '../../../providers.dart';
import '../../../shared/widgets/error_display.dart';
import '../../../shared/widgets/loading_indicator.dart';

class ProfileScreen extends ConsumerStatefulWidget {
  final String userId;

  const ProfileScreen({super.key, required this.userId});

  @override
  ConsumerState<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends ConsumerState<ProfileScreen> {
  User? _user;
  List<Vod> _vods = [];
  bool _isLoading = true;
  bool _isFollowLoading = false;
  String? _error;

  @override
  void initState() {
    super.initState();
    _fetchUserData();
  }

  Future<void> _fetchUserData() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final userRepo = ref.read(userRepositoryProvider);
      final vodRepo = ref.read(vodRepositoryProvider);

      // Fire both requests in parallel.
      final userFuture = userRepo.getUser(widget.userId);
      final vodsFuture = vodRepo.getUserVods(widget.userId);

      final userResponse = await userFuture;
      if (!mounted) return;

      if (userResponse.success && userResponse.data != null) {
        final vodsResponse = await vodsFuture;
        if (!mounted) return;

        setState(() {
          _user = userResponse.data;
          _vods = vodsResponse.data ?? [];
          _isLoading = false;
        });
      } else {
        setState(() {
          _error = userResponse.message;
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

  Future<void> _toggleFollow() async {
    final user = _user;
    if (user == null) return;

    final wasFollowing = user.isFollowing == true;
    setState(() => _isFollowLoading = true);

    // Optimistic update.
    setState(() {
      _user = user.copyWith(
        isFollowing: !wasFollowing,
        followerCount: user.followerCount + (wasFollowing ? -1 : 1),
      );
    });

    try {
      final userRepo = ref.read(userRepositoryProvider);
      final response = wasFollowing
          ? await userRepo.unfollowUser(user.id)
          : await userRepo.followUser(user.id);

      if (!mounted) return;

      if (!response.success) {
        // Revert on failure.
        setState(() => _user = user);
      }
    } on DioException catch (_) {
      if (mounted) setState(() => _user = user);
    } finally {
      if (mounted) setState(() => _isFollowLoading = false);
    }
  }

  bool get _isOwnProfile {
    final currentUser = ref.read(currentUserProvider);
    return currentUser?.id == widget.userId;
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);

    return FScaffold(
      header: FHeader.nested(
        title: Text(l10n.settingsNavProfile),
      ),
      child: _isLoading
          ? LoadingIndicator(message: l10n.loading)
          : _error != null
              ? ErrorDisplay(
                  title: l10n.errorGeneralTitle,
                  message: _error,
                  onRetry: _fetchUserData,
                )
              : _buildContent(context),
    );
  }

  Widget _buildContent(BuildContext context) {
    final user = _user;
    if (user == null) return const SizedBox.shrink();

    return RefreshIndicator(
      onRefresh: _fetchUserData,
      child: SingleChildScrollView(
        physics: const AlwaysScrollableScrollPhysics(),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _buildHeader(context, user),
            _buildUserInfo(context, user),
            if (_vods.isNotEmpty) _buildRecentStreams(context),
          ],
        ),
      ),
    );
  }

  // ---------------------------------------------------------------------------
  // Header: background + profile picture
  // ---------------------------------------------------------------------------

  Widget _buildHeader(BuildContext context, User user) {
    final colors = context.theme.colors;

    return Stack(
      clipBehavior: Clip.none,
      children: [
        // Background picture
        SizedBox(
          height: 150,
          width: double.infinity,
          child: user.backgroundPicture != null
              ? CachedNetworkImage(
                  imageUrl:
                      '${AppConfig.apiUrl}/${user.backgroundPicture}',
                  fit: BoxFit.cover,
                  placeholder: (_, _) => ColoredBox(color: colors.muted),
                  errorWidget: (_, _, _) => ColoredBox(color: colors.muted),
                )
              : ColoredBox(color: colors.muted),
        ),
        // Profile picture (overlapping)
        Positioned(
          bottom: -40,
          left: 16,
          child: Container(
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              border: Border.all(color: colors.background, width: 4),
            ),
            child: CircleAvatar(
              radius: 40,
              backgroundColor: colors.muted,
              backgroundImage: user.profilePicture != null
                  ? CachedNetworkImageProvider(
                      '${AppConfig.apiUrl}/${user.profilePicture}')
                  : null,
              child: user.profilePicture == null
                  ? const Icon(FIcons.user, size: 32)
                  : null,
            ),
          ),
        ),
      ],
    );
  }

  // ---------------------------------------------------------------------------
  // User info: name, follow, followers, bio, socials, joined
  // ---------------------------------------------------------------------------

  Widget _buildUserInfo(BuildContext context, User user) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;
    final l10n = AppLocalizations.of(context);

    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 52, 16, 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Name + follow button
          Row(
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      user.displayName ?? user.username,
                      style: typography.xl2
                          .copyWith(fontWeight: FontWeight.bold),
                    ),
                    Text(
                      '@${user.username}',
                      style: typography.sm
                          .copyWith(color: colors.mutedForeground),
                    ),
                  ],
                ),
              ),
              if (!_isOwnProfile)
                FButton(
                  onPress: _isFollowLoading ? null : _toggleFollow,
                  variant: user.isFollowing == true
                      ? FButtonVariant.destructive
                      : null,
                  child: _isFollowLoading
                      ? const SizedBox(
                          height: 16,
                          width: 16,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : Text(user.isFollowing == true
                          ? l10n.unfollow
                          : l10n.follow),
                ),
            ],
          ),
          const SizedBox(height: 12),

          // Follower count
          Row(
            children: [
              Icon(FIcons.users, size: 16, color: colors.mutedForeground),
              const SizedBox(width: 6),
              Text(
                '${user.followerCount} ${l10n.usersProfileFollowers(user.followerCount)}',
                style:
                    typography.sm.copyWith(color: colors.mutedForeground),
              ),
            ],
          ),
          const SizedBox(height: 16),

          // About
          Text(
            l10n.usersProfileAbout,
            style: typography.base.copyWith(fontWeight: FontWeight.w600),
          ),
          const SizedBox(height: 4),
          Text(
            user.bio?.isNotEmpty == true
                ? user.bio!
                : l10n.noDescription,
            style: typography.sm.copyWith(
              color: user.bio?.isNotEmpty == true
                  ? colors.foreground
                  : colors.mutedForeground,
            ),
          ),

          // Social media links
          if (user.socialMediaLinks != null)
            _buildSocialMediaLinks(context, user.socialMediaLinks!),

          const SizedBox(height: 12),

          // Joined date
          Row(
            children: [
              Icon(FIcons.calendar, size: 16,
                  color: colors.mutedForeground),
              const SizedBox(width: 6),
              Text(
                '${l10n.usersProfileJoinedPrefix} ${_formatDate(user.createdAt)}',
                style:
                    typography.sm.copyWith(color: colors.mutedForeground),
              ),
            ],
          ),
        ],
      ),
    );
  }

  // ---------------------------------------------------------------------------
  // Social media links
  // ---------------------------------------------------------------------------

  Widget _buildSocialMediaLinks(
      BuildContext context, SocialMediaLinks links) {
    final colors = context.theme.colors;

    final entries = <MapEntry<IconData, String>>[
      if (links.website != null) MapEntry(FIcons.globe, links.website!),
      if (links.github != null) MapEntry(FIcons.github, links.github!),
      if (links.youtube != null) MapEntry(FIcons.youtube, links.youtube!),
      if (links.facebook != null)
        MapEntry(FIcons.facebook, links.facebook!),
      if (links.twitter != null) MapEntry(FIcons.twitter, links.twitter!),
      if (links.instagram != null)
        MapEntry(FIcons.instagram, links.instagram!),
      if (links.linkedin != null)
        MapEntry(FIcons.linkedin, links.linkedin!),
      if (links.tiktok != null) MapEntry(FIcons.music, links.tiktok!),
    ];

    if (entries.isEmpty) return const SizedBox.shrink();

    return Padding(
      padding: const EdgeInsets.only(top: 12),
      child: Wrap(
        spacing: 16,
        runSpacing: 8,
        children: entries.map((entry) {
          return Tooltip(
            message: entry.value,
            child: Icon(entry.key, size: 20,
                color: colors.mutedForeground),
          );
        }).toList(),
      ),
    );
  }

  // ---------------------------------------------------------------------------
  // Recent streams (VODs)
  // ---------------------------------------------------------------------------

  Widget _buildRecentStreams(BuildContext context) {
    final typography = context.theme.typography;
    final l10n = AppLocalizations.of(context);

    return Padding(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            l10n.usersProfileRecentStreams,
            style: typography.lg.copyWith(fontWeight: FontWeight.w600),
          ),
          const SizedBox(height: 12),
          GridView.builder(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            gridDelegate:
                const SliverGridDelegateWithFixedCrossAxisCount(
              crossAxisCount: 2,
              crossAxisSpacing: 10,
              mainAxisSpacing: 10,
              childAspectRatio: 0.75,
            ),
            itemCount: _vods.length,
            itemBuilder: (context, index) =>
                _ProfileVodCard(vod: _vods[index]),
          ),
        ],
      ),
    );
  }

  // ---------------------------------------------------------------------------
  // Helpers
  // ---------------------------------------------------------------------------

  String _formatDate(String dateStr) {
    try {
      final date = DateTime.parse(dateStr);
      const months = [
        'Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun',
        'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec',
      ];
      return '${months[date.month - 1]} ${date.year}';
    } catch (_) {
      return dateStr;
    }
  }
}

// -----------------------------------------------------------------------------
// VOD card (reused styling from HomeScreen)
// -----------------------------------------------------------------------------

class _ProfileVodCard extends StatelessWidget {
  final Vod vod;

  const _ProfileVodCard({required this.vod});

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;
    final l10n = AppLocalizations.of(context);

    return DecoratedBox(
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
            borderRadius:
                const BorderRadius.vertical(top: Radius.circular(12)),
            child: AspectRatio(
              aspectRatio: 16 / 9,
              child: Stack(
                fit: StackFit.expand,
                children: [
                  if (vod.thumbnailUrl != null)
                    CachedNetworkImage(
                      imageUrl:
                          '${AppConfig.apiUrl}/${vod.thumbnailUrl}',
                      fit: BoxFit.cover,
                      placeholder: (_, _) => ColoredBox(
                        color: colors.muted,
                        child: const Center(
                            child: Icon(FIcons.film, size: 24)),
                      ),
                      errorWidget: (_, _, _) => ColoredBox(
                        color: colors.muted,
                        child: const Center(
                            child: Icon(FIcons.film, size: 24)),
                      ),
                    )
                  else
                    ColoredBox(
                      color: colors.muted,
                      child: const Center(
                          child: Icon(FIcons.film, size: 24)),
                    ),
                  // Duration badge
                  Positioned(
                    bottom: 4,
                    right: 4,
                    child: Container(
                      padding: const EdgeInsets.symmetric(
                          horizontal: 6, vertical: 2),
                      decoration: BoxDecoration(
                        color: Colors.black.withValues(alpha: 0.7),
                        borderRadius: BorderRadius.circular(4),
                      ),
                      child: Text(
                        _formatDuration(vod.duration),
                        style: typography.xs
                            .copyWith(color: Colors.white),
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
                    style: typography.xs
                        .copyWith(fontWeight: FontWeight.w600),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                  const Spacer(),
                  Text(
                    l10n.homeViewCount(vod.viewCount),
                    style: typography.xs
                        .copyWith(color: colors.mutedForeground),
                  ),
                ],
              ),
            ),
          ),
        ],
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
