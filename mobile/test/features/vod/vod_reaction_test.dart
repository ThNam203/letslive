import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/features/vod/data/vod_reaction.dart';

import '../watch/fake_player.dart';
import 'vod_test_app.dart';

final _like = find.byKey(const ValueKey('vod-like'));
final _dislike = find.byKey(const ValueKey('vod-dislike'));

/// Like count shown on the like button.
Finder _count(String text) =>
    find.descendant(of: _like, matching: find.text(text));

/// Like and dislike on the video itself, and the model behind them.
void main() {
  group('VodReactionState.reactedWith', () {
    const none = VodReactionState(likeCount: 3);

    test('liking adds one, disliking after a like removes it', () {
      final liked = none.reactedWith(VodReaction.like);
      expect(liked.likeCount, 4);
      expect(liked.reaction, VodReaction.like);
      expect(liked.reactedWith(VodReaction.dislike).likeCount, 3);
      expect(liked.reactedWith(null).likeCount, 3);
    });

    test('a dislike never changes the public count, nor goes below zero', () {
      expect(none.reactedWith(VodReaction.dislike).likeCount, 3);
      const liked = VodReactionState(likeCount: 0, reaction: VodReaction.like);
      expect(liked.reactedWith(null).likeCount, 0);
    });
  });

  testWidgets('shows the viewer\'s reaction and the like count', (
    tester,
  ) async {
    final server = FakeVodServer()..reaction = 'like';
    await pumpVod(tester, server, FakePlayers());

    // 3 likes from the video + the viewer's own, as the server counts them.
    expect(find.byIcon(Icons.thumb_up), findsOneWidget);
    expect(_count('4'), findsOneWidget);
  });

  testWidgets('liking is optimistic and the server answer replaces it', (
    tester,
  ) async {
    final server = FakeVodServer()..holdReaction = Completer<void>();
    final app = await pumpVod(tester, server, FakePlayers());
    expect(_count('3'), findsOneWidget);

    await tester.tap(_like);
    await tester.pump();
    expect(find.byIcon(Icons.thumb_up), findsOneWidget);
    expect(_count('4'), findsOneWidget);

    server.holdReaction!.complete();
    await tester.pumpAndSettle();

    final put = app
        .requestsTo('${ApiEndpoints.vods}/vod-1/reaction')
        .where((r) => r.method == 'PUT')
        .single;
    expect(put.data, {'reaction': 'like'});
    expect(_count('4'), findsOneWidget);
  });

  testWidgets('tapping twice while a reaction is in flight sends one request', (
    tester,
  ) async {
    final server = FakeVodServer()..holdReaction = Completer<void>();
    final app = await pumpVod(tester, server, FakePlayers());

    await tester.tap(_like);
    await tester.pumpAndSettle();
    await tester.tap(_like);
    await tester.pumpAndSettle();
    server.holdReaction!.complete();
    await tester.pumpAndSettle();

    final writes = app
        .requestsTo('${ApiEndpoints.vods}/vod-1/reaction')
        .where((r) => r.method != 'GET');
    expect(writes.map((r) => r.method), ['PUT']);
    expect(find.byIcon(Icons.thumb_up), findsOneWidget);
  });

  testWidgets('a refused reaction is undone with a message', (tester) async {
    final server = FakeVodServer()
      ..holdReaction = Completer<void>()
      ..writeStatus = 503;
    await pumpVod(tester, server, FakePlayers());

    await tester.tap(_like);
    await tester.pump();
    expect(_count('4'), findsOneWidget);

    server.holdReaction!.complete();
    await tester.pumpAndSettle();

    expect(find.byIcon(Icons.thumb_up_outlined), findsOneWidget);
    expect(_count('3'), findsOneWidget);
    expect(find.text('Something went wrong.'), findsOneWidget);
  });

  testWidgets('tapping the same reaction again removes it', (tester) async {
    final server = FakeVodServer();
    final app = await pumpVod(tester, server, FakePlayers());

    await tester.tap(_dislike);
    await tester.pumpAndSettle();
    expect(find.byIcon(Icons.thumb_down), findsOneWidget);

    await tester.tap(_dislike);
    await tester.pumpAndSettle();
    expect(find.byIcon(Icons.thumb_down_outlined), findsOneWidget);
    final methods = app
        .requestsTo('${ApiEndpoints.vods}/vod-1/reaction')
        .map((r) => r.method);
    expect(methods, ['GET', 'PUT', 'DELETE']);
  });

  testWidgets('when the reaction cannot load the buttons use the video count', (
    tester,
  ) async {
    final server = FakeVodServer()..reactionStatus = 503;
    await pumpVod(tester, server, FakePlayers());

    expect(_count('3'), findsOneWidget);
    expect(find.byIcon(Icons.thumb_up_outlined), findsOneWidget);
  });

  testWidgets('reopening the video loads everything again, no stale state', (
    tester,
  ) async {
    final server = FakeVodServer()..reaction = 'like';
    final app = await pumpVod(tester, server, FakePlayers());
    expect(app.requestsTo('${ApiEndpoints.vods}/vod-1/reaction'), hasLength(1));

    await tester.pageBack();
    await tester.pumpAndSettle();
    server.reaction = null;
    // Open it again from the router.
    await pumpVodAgain(tester, app);

    expect(app.requestsTo('${ApiEndpoints.vods}/vod-1/reaction'), hasLength(2));
    expect(find.byIcon(Icons.thumb_up_outlined), findsOneWidget);
  });
}
