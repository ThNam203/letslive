import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/features/home/domain/media_format.dart';
import 'package:letslive_mobile/features/profile/domain/profile_format.dart';
import 'package:letslive_mobile/features/profile/profile_providers.dart';
import 'package:letslive_mobile/features/user/data/public_user.dart';

/// The About tab of a profile (web `ProfileAboutTab`): the bio, channel
/// stats and the social links. Only http(s) links are shown (they are user
/// supplied); a tap opens the browser.
class ProfileAboutTab extends ConsumerWidget {
  const ProfileAboutTab({required this.user, super.key});

  /// The profile owner.
  final PublicUser user;

  Future<void> _open(BuildContext context, WidgetRef ref, String url) async {
    final l10n = AppLocalizations.of(context);
    final messenger = ScaffoldMessenger.of(context);
    final uri = Uri.tryParse(url);
    final opened = uri != null && await ref.read(urlOpenerProvider)(uri);
    if (!opened) {
      messenger.showSnackBar(
        SnackBar(content: Text(l10n.api_response_default_error)),
      );
    }
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final theme = Theme.of(context);
    final bio = user.bio?.trim() ?? '';
    final links = [
      for (final MapEntry(:value) in user.socialMediaLinks.entries)
        if (socialLinkLabel(value) case final label?)
          (url: value, label: label),
    ];
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        Text(
          '${l10n.users_profile_about} ${user.username}',
          style: theme.textTheme.titleMedium,
        ),
        const SizedBox(height: 8),
        Text(
          bio.isEmpty ? l10n.users_channel_no_bio(user.username) : bio,
          style: bio.isEmpty
              ? theme.textTheme.bodyMedium?.copyWith(
                  color: theme.colorScheme.onSurfaceVariant,
                )
              : theme.textTheme.bodyMedium,
        ),
        const SizedBox(height: 24),
        Text(l10n.users_channel_stats, style: theme.textTheme.titleSmall),
        ListTile(
          contentPadding: EdgeInsets.zero,
          leading: const Icon(Icons.people_outline),
          title: Text(
            l10n.common_followers_count(
              compactCount(l10n, user.followerCount),
              user.followerCount,
            ),
          ),
        ),
        ListTile(
          contentPadding: EdgeInsets.zero,
          leading: const Icon(Icons.calendar_today_outlined),
          title: Text(joinedLabel(l10n, user.createdAt)),
        ),
        if (links.isNotEmpty) ...[
          const SizedBox(height: 16),
          Text(l10n.users_channel_links, style: theme.textTheme.titleSmall),
          for (final link in links)
            ListTile(
              key: ValueKey('social-${link.url}'),
              contentPadding: EdgeInsets.zero,
              leading: const Icon(Icons.link),
              title: Text(
                link.label,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
              onTap: () => _open(context, ref, link.url),
            ),
        ],
      ],
    );
  }
}
