import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/router/app_routes.dart';
import 'package:letslive_mobile/core/widgets/feed_section.dart';
import 'package:letslive_mobile/features/home/domain/media_format.dart';
import 'package:letslive_mobile/features/user/data/public_user.dart';
import 'package:letslive_mobile/features/user/presentation/user_avatar.dart';
import 'package:letslive_mobile/features/user/user_providers.dart';

/// The People tab of the user's own profile: web's left-bar channel lists,
/// the channels the user follows and the recommended ones (those not already
/// followed, like web). Each list loads and fails on its own and is left out
/// when empty; pull-to-refresh reloads both. A row opens the channel.
class ProfilePeopleTab extends ConsumerWidget {
  const ProfilePeopleTab({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final following = ref.watch(followingChannelsProvider);
    final recommended = ref.watch(recommendedChannelsProvider);
    // If the following request failed, recommended may include channels the
    // user follows.
    final followedIds = {
      for (final user in following.value ?? const <PublicUser>[]) user.id,
    };
    final recommendedNotFollowed = recommended.whenData(
      (users) => [
        for (final user in users)
          if (!followedIds.contains(user.id)) user,
      ],
    );
    return RefreshIndicator(
      onRefresh: () => Future.wait([
        ref
            .refresh(followingChannelsProvider.future)
            .then<void>((_) {}, onError: (Object _) {}),
        ref
            .refresh(recommendedChannelsProvider.future)
            .then<void>((_) {}, onError: (Object _) {}),
      ]),
      child: ListView(
        // Lets pull-to-refresh work when the lists are short.
        physics: const AlwaysScrollableScrollPhysics(),
        children: [
          FeedSection(
            key: const ValueKey('section-following'),
            title: l10n.common_following,
            value: following,
            onRetry: () => ref.invalidate(followingChannelsProvider),
            builder: (context, users) => _PeopleList(users: users),
          ),
          FeedSection(
            key: const ValueKey('section-recommended'),
            title: l10n.common_recommended,
            value: recommendedNotFollowed,
            onRetry: () => ref.invalidate(recommendedChannelsProvider),
            builder: (context, users) => _PeopleList(users: users),
          ),
        ],
      ),
    );
  }
}

/// One row per channel: avatar, name and follower count.
class _PeopleList extends StatelessWidget {
  const _PeopleList({required this.users});

  final List<PublicUser> users;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Column(
      children: [
        for (final user in users)
          ListTile(
            key: ValueKey('person-${user.id}'),
            leading: UserAvatar(
              pictureUrl: user.profilePicture,
              name: user.username,
            ),
            title: Text(
              user.username,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
            subtitle: Text(
              l10n.common_followers_count(
                compactCount(l10n, user.followerCount),
                user.followerCount,
              ),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
            onTap: () => context.push(AppRoutes.channel(user.id)),
          ),
      ],
    );
  }
}
