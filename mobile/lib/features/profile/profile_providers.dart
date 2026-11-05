/// Riverpod providers for the profile screens (a channel `/users/:userId`
/// and the own profile tab): the user's videos, and the hook that opens
/// links. Data providers are `autoDispose`, so logout drops them.
library;

import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/network/api_load.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/core/pagination/paged_feed.dart';
import 'package:letslive_mobile/features/home/data/vod.dart';
import 'package:letslive_mobile/features/vod/vod_providers.dart';
import 'package:url_launcher/url_launcher.dart';

/// Videos per page on a channel (web `CHANNEL_VODS_PAGE_SIZE` in
/// `hooks/queries/use-vods.ts`).
const int channelVodsPageSize = 12;

/// The public videos of one user, newest first, paged. Shared by the Home
/// and Videos tabs, so they cost one request.
final channelVodsProvider = AsyncNotifierProvider.autoDispose
    .family<ChannelVodsNotifier, PagedFeed<Vod>, String>(
      ChannelVodsNotifier.new,
      retry: noRetry,
    );

/// Loads the videos of [userId] page by page.
class ChannelVodsNotifier extends PagedFeedNotifier<Vod> {
  ChannelVodsNotifier(this.userId);

  /// The channel owner.
  final String userId;

  @override
  int get pageSize => channelVodsPageSize;

  @override
  String idOf(Vod item) => item.id;

  @override
  Future<ApiResponse<List<Vod>>> fetchPage(int page) => ref
      .read(vodRepositoryProvider)
      .ofUser(userId, page: page, limit: channelVodsPageSize);
}

/// Opens a link outside the app; true when something handled it. Widget
/// tests override it with a fake.
typedef UrlOpener = Future<bool> Function(Uri url);

/// Opens links in the browser (the About tab's social links).
final urlOpenerProvider = Provider<UrlOpener>(
  (ref) => (url) async {
    try {
      return await launchUrl(url, mode: LaunchMode.externalApplication);
    } on PlatformException {
      return false;
    }
  },
);
