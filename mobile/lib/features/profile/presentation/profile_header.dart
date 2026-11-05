import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/features/home/domain/media_format.dart';
import 'package:letslive_mobile/features/profile/domain/profile_format.dart';
import 'package:letslive_mobile/features/user/data/public_user.dart';
import 'package:letslive_mobile/features/user/presentation/follow_button.dart';
import 'package:letslive_mobile/features/user/presentation/user_avatar.dart';

/// Height of the profile banner.
const double _bannerHeight = 96;

/// Avatar radius in the header.
const double _avatarRadius = 28;

/// Upper bound of the header's height without the banner at the current text
/// scale, for layouts that must decide whether everything fits.
double profileHeaderHeight(BuildContext context) {
  final scaler = MediaQuery.textScalerOf(context);
  // Padding, name, two small lines, then the follow button (label plus its
  // own padding) under them.
  return 24 +
      scaler.scale(22) * 1.4 +
      scaler.scale(12) * 1.5 * 2 +
      8 +
      scaler.scale(14) * 1.5 +
      24;
}

/// The top of a profile (web `ProfileHeader`): the banner when
/// [showBanner], then the avatar, username, follower count, join date and
/// the follow button. The banner is a gradient until the user uploaded
/// one. Long names are cut with an ellipsis.
class ProfileHeader extends StatelessWidget {
  const ProfileHeader({required this.user, this.showBanner = true, super.key});

  /// The user shown.
  final PublicUser user;

  /// Live channels already show the player, so they leave the banner out
  /// (web does the same).
  final bool showBanner;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final theme = Theme.of(context);
    final banner = user.backgroundPicture;
    final muted = theme.textTheme.bodySmall?.copyWith(
      color: theme.colorScheme.onSurfaceVariant,
    );
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        if (showBanner)
          SizedBox(
            height: _bannerHeight,
            child: DecoratedBox(
              decoration: BoxDecoration(
                gradient: LinearGradient(
                  colors: [
                    theme.colorScheme.primary.withValues(alpha: 0.3),
                    theme.colorScheme.surfaceContainerHighest,
                  ],
                ),
              ),
              child: banner == null
                  ? null
                  : CachedNetworkImage(
                      key: const ValueKey('profile-banner'),
                      imageUrl: banner,
                      fit: BoxFit.cover,
                      errorWidget: (_, _, _) => const SizedBox.shrink(),
                    ),
            ),
          ),
        Padding(
          padding: const EdgeInsets.all(12),
          child: Row(
            children: [
              UserAvatar(
                pictureUrl: user.profilePicture,
                name: user.username,
                radius: _avatarRadius,
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      user.username,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: theme.textTheme.titleLarge,
                    ),
                    Text(
                      l10n.common_followers_count(
                        compactCount(l10n, user.followerCount),
                        user.followerCount,
                      ),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: muted,
                    ),
                    Text(
                      joinedLabel(l10n, user.createdAt),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: muted,
                    ),
                    // Under the text, not beside it: a long label (Vietnamese
                    // "Bỏ theo dõi" at large text) would squeeze the name away.
                    Align(
                      alignment: Alignment.centerLeft,
                      child: FollowButton(userId: user.id),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }
}
