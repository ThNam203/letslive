/// The banner and profile picture at the top of the edit-profile screen,
/// laid out like web's `ProfileBanner`: the avatar overlaps the bottom edge
/// of the banner, and each picture is itself the button that changes it.
///
/// Web shows its "change" overlay on hover; a phone has no hover, so each
/// picture carries a small camera badge instead, and a long press shows what
/// the tap does.
library;

import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/features/user/data/picked_image.dart';
import 'package:letslive_mobile/features/user/presentation/user_avatar.dart';

/// Radius of the profile picture.
const double _avatarRadius = 44;

/// Width of the ring around the profile picture (web's `border-4`).
const double _avatarRing = 4;

/// Width : height of the banner.
const double _bannerAspectRatio = 3;

/// Banner plus overlapping avatar; tapping either picks a replacement.
class ProfileImagesHeader extends StatelessWidget {
  const ProfileImagesHeader({
    required this.pendingBackground,
    required this.backgroundUrl,
    required this.pendingAvatar,
    required this.avatarUrl,
    required this.name,
    required this.onPickBackground,
    required this.onPickAvatar,
    super.key,
  });

  /// The chosen, not yet saved banner; shown instead of [backgroundUrl].
  final PickedImage? pendingBackground;

  /// The saved banner, or null for a flat placeholder.
  final String? backgroundUrl;

  /// The chosen, not yet saved picture; shown instead of [avatarUrl].
  final PickedImage? pendingAvatar;

  /// The saved picture, or null for the initial.
  final String? avatarUrl;

  /// Source of the initial shown when there is no picture.
  final String name;

  /// Called on a banner tap; null disables it (e.g. while saving).
  final VoidCallback? onPickBackground;

  /// Called on an avatar tap; null disables it (e.g. while saving).
  final VoidCallback? onPickAvatar;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Stack(
      children: [
        // Room below the banner for the lower half of the avatar.
        Padding(
          padding: const EdgeInsets.only(bottom: _avatarRadius + _avatarRing),
          child: _Banner(
            pending: pendingBackground,
            url: backgroundUrl,
            label: l10n.settings_profile_update_background,
            onTap: onPickBackground,
          ),
        ),
        Positioned.fill(
          child: Align(
            alignment: Alignment.bottomCenter,
            child: _Avatar(
              pending: pendingAvatar,
              url: avatarUrl,
              name: name,
              label: l10n.settings_profile_update_profile_picture,
              onTap: onPickAvatar,
            ),
          ),
        ),
      ],
    );
  }
}

/// The banner: the chosen image, else the saved one, else a flat
/// placeholder; the whole area is the change button.
class _Banner extends StatelessWidget {
  const _Banner({
    required this.pending,
    required this.url,
    required this.label,
    required this.onTap,
  });

  final PickedImage? pending;
  final String? url;
  final String label;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final image = pending;
    final saved = url;
    return AspectRatio(
      aspectRatio: _bannerAspectRatio,
      child: ClipRRect(
        borderRadius: BorderRadius.circular(8),
        child: Stack(
          fit: StackFit.expand,
          children: [
            ColoredBox(
              color: Theme.of(context).colorScheme.surfaceContainerHighest,
              child: image != null
                  ? Image.memory(image.bytes, fit: BoxFit.cover)
                  : saved == null
                  ? null
                  : CachedNetworkImage(
                      imageUrl: saved,
                      fit: BoxFit.cover,
                      errorWidget: (_, _, _) => const SizedBox.shrink(),
                    ),
            ),
            const Positioned(top: 8, right: 8, child: _CameraBadge()),
            Material(
              type: MaterialType.transparency,
              child: Tooltip(
                message: label,
                child: InkWell(
                  key: const ValueKey('pick-background'),
                  onTap: onTap,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

/// The round profile picture with a ring, overlapping the banner; the whole
/// circle is the change button.
class _Avatar extends StatelessWidget {
  const _Avatar({
    required this.pending,
    required this.url,
    required this.name,
    required this.label,
    required this.onTap,
  });

  final PickedImage? pending;
  final String? url;
  final String name;
  final String label;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final image = pending;
    final picture = image == null
        ? UserAvatar(pictureUrl: url, name: name, radius: _avatarRadius)
        : CircleAvatar(
            radius: _avatarRadius,
            backgroundImage: MemoryImage(image.bytes),
          );
    return Stack(
      children: [
        DecoratedBox(
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            border: Border.all(
              color: Theme.of(context).colorScheme.surface,
              width: _avatarRing,
            ),
          ),
          child: picture,
        ),
        const Positioned(right: 0, bottom: 0, child: _CameraBadge()),
        Positioned.fill(
          child: Material(
            type: MaterialType.transparency,
            shape: const CircleBorder(),
            clipBehavior: Clip.antiAlias,
            child: Tooltip(
              message: label,
              child: InkWell(key: const ValueKey('pick-avatar'), onTap: onTap),
            ),
          ),
        ),
      ],
    );
  }
}

/// The small camera mark that says "tap to change this picture".
class _CameraBadge extends StatelessWidget {
  const _CameraBadge();

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    return DecoratedBox(
      decoration: BoxDecoration(
        color: scheme.primary,
        shape: BoxShape.circle,
        border: Border.all(color: scheme.surface, width: 2),
      ),
      child: Padding(
        padding: const EdgeInsets.all(6),
        child: Icon(
          Icons.photo_camera_outlined,
          size: 16,
          color: scheme.onPrimary,
        ),
      ),
    );
  }
}
