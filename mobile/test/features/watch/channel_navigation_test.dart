import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/features/home/presentation/home_screen.dart';
import 'package:letslive_mobile/features/watch/presentation/channel_screen.dart';
import 'package:letslive_mobile/features/watch/watch_providers.dart';

import '../home/home_fixtures.dart';
import '../search/search_test_app.dart';
import 'fake_live_player.dart';
import 'watch_test_app.dart';

/// Taps on Home and Search cards that lead to `/users/:userId`.
void main() {
  testWidgets('tapping a live card on home opens the channel', (tester) async {
    final server = FakeWatchServer()
      ..live = [livestreamJson(id: 'ls-1', title: 'Morning stream')]
      ..liveByUser = {
        'u-1': livestreamJson(id: 'ls-1', title: 'Morning stream'),
      };
    final players = FakeLivePlayers();
    await pumpWatchHome(tester, server, players);

    await tester.tap(find.text('Morning stream'));
    await tester.pumpAndSettle();

    expect(find.byType(ChannelScreen), findsOneWidget);
    expect(players.last.url, endsWith('/transcode/ls-1/index.m3u8'));

    // Back returns to the feed.
    await tester.pageBack();
    await tester.pumpAndSettle();
    expect(find.byType(HomeScreen), findsOneWidget);
  });

  testWidgets('tapping a followed channel on home opens it', (tester) async {
    final server = FakeWatchServer()
      ..users = {
        'u-1': publicUserJson(),
        'u-2': publicUserJson(id: 'u-2', username: 'followed_one'),
      }
      ..following = [publicUserJson(id: 'u-2', username: 'followed_one')];
    await pumpWatchHome(tester, server, FakeLivePlayers());

    await tester.tap(find.text('followed_one'));
    await tester.pumpAndSettle();

    expect(find.byType(ChannelScreen), findsOneWidget);
    expect(find.text("followed_one isn't live right now."), findsOneWidget);
  });

  testWidgets('video cards are not tappable yet', (tester) async {
    final server = FakeWatchServer()..vods = [vodJson(title: 'Best of week')];
    await pumpWatchHome(tester, server, FakeLivePlayers());

    await tester.tap(find.text('Best of week'));
    await tester.pumpAndSettle();

    expect(find.byType(ChannelScreen), findsNothing);
  });

  testWidgets('tapping a live result in search opens the channel', (
    tester,
  ) async {
    final server = FakeWatchServer()
      ..searchLives = [livestreamJson(id: 'ls-9', title: 'Trò chơi live')]
      ..liveByUser = {'u-1': livestreamJson(id: 'ls-9', title: 'Trò chơi')};
    final players = FakeLivePlayers();
    await pumpSearch(
      tester,
      server,
      overrides: [livePlayerFactoryProvider.overrideWithValue(players.create)],
    );
    await search(tester, 'tro choi');

    await tester.tap(find.byKey(const ValueKey('ls-9')));
    await tester.pumpAndSettle();

    expect(find.byType(ChannelScreen), findsOneWidget);
    expect(players.last.url, endsWith('/transcode/ls-9/index.m3u8'));
  });

  testWidgets('tapping a channel result in search opens it', (tester) async {
    final server = FakeWatchServer()
      ..users = {'u-1': publicUserJson(), 'u-3': publicUserJson(id: 'u-3')}
      ..searchUsers = [publicUserJson(id: 'u-3', username: 'found_one')];
    await pumpSearch(
      tester,
      server,
      overrides: [
        livePlayerFactoryProvider.overrideWithValue(FakeLivePlayers().create),
      ],
    );
    await search(tester, 'found');

    await tester.tap(find.byKey(const ValueKey('u-3')));
    await tester.pumpAndSettle();

    expect(find.byType(ChannelScreen), findsOneWidget);
  });
}
