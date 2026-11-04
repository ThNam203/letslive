import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/features/home/presentation/home_screen.dart';

import 'home_fixtures.dart';
import 'home_test_app.dart';

/// Finder for something inside the section with [key].
Finder inSection(String key, Finder matching) =>
    find.descendant(of: find.byKey(ValueKey(key)), matching: matching);

void main() {
  testWidgets('home shows live and popular sections', (tester) async {
    final server = FakeHomeServer()
      ..live = [livestreamJson(title: 'Morning stream')]
      ..vods = [vodJson(title: 'Best of week')]
      ..following = [publicUserJson(id: 'u-2', username: 'followed_one')]
      ..recommended = [
        publicUserJson(id: 'u-2', username: 'followed_one'),
        publicUserJson(id: 'u-3', username: 'suggested_one'),
      ];
    await pumpHome(tester, server);

    expect(find.byType(HomeScreen), findsOneWidget);
    expect(inSection('section-live', find.text('Morning stream')), findsOne);
    expect(inSection('section-live', find.text('streamer')), findsOne);
    expect(
      inSection(
        'section-live',
        find.textContaining('1.5K views · Started 5 minutes ago'),
      ),
      findsOne,
    );
    expect(inSection('section-vods', find.text('Best of week')), findsOne);
    expect(inSection('section-vods', find.text('1:02:05')), findsOne);
    expect(inSection('section-following', find.text('followed_one')), findsOne);
    // Already followed channels are not recommended again.
    expect(
      inSection('section-recommended', find.text('followed_one')),
      findsNothing,
    );
    expect(
      inSection('section-recommended', find.text('suggested_one')),
      findsOne,
    );
  });

  testWidgets('empty feed shows empty state', (tester) async {
    await pumpHome(tester, FakeHomeServer());

    expect(find.text('No livestreams'), findsOneWidget);
    expect(find.text('No Videos Available'), findsOneWidget);
    // Channel lists are hidden when empty, like on web.
    expect(find.text('Following'), findsNothing);
    expect(find.text('Recommended'), findsNothing);
  });

  testWidgets('error shows retry that refetches', (tester) async {
    final server = FakeHomeServer()
      ..liveStatus = 503
      ..vods = [vodJson(title: 'Still here')];
    final app = await pumpHome(tester, server);

    // One failing section does not take the others down.
    expect(inSection('section-live', find.text('Try again')), findsOne);
    expect(inSection('section-vods', find.text('Still here')), findsOne);

    server
      ..liveStatus = 200
      ..live = [livestreamJson(title: 'Back online')];
    await tester.tap(inSection('section-live', find.text('Try again')));
    await tester.pumpAndSettle();

    expect(find.text('Back online'), findsOneWidget);
    expect(find.text('Try again'), findsNothing);
    expect(app.requestsTo(ApiEndpoints.popularLivestreams), hasLength(2));
    expect(app.requestsTo(ApiEndpoints.popularVods), hasLength(1));
  });

  testWidgets('pull to refresh reloads', (tester) async {
    final server = FakeHomeServer()..live = [livestreamJson(title: 'Old')];
    final app = await pumpHome(tester, server);
    expect(find.text('Old'), findsOneWidget);

    server.live = [livestreamJson(title: 'New')];
    // The indicator triggers after a drag of a quarter of the (tall) view.
    await tester.drag(find.byType(ListView).first, const Offset(0, 4000));
    await tester.pumpAndSettle();

    expect(find.text('New'), findsOneWidget);
    for (final path in [
      ApiEndpoints.popularLivestreams,
      ApiEndpoints.popularVods,
      ApiEndpoints.following,
      ApiEndpoints.recommendedChannels,
    ]) {
      expect(app.requestsTo(path), hasLength(2), reason: path);
    }
    // Owners are looked up again too, once per id.
    expect(app.requestsTo('${ApiEndpoints.user}/u-1'), hasLength(2));
  });

  testWidgets('coming back from the background refetches the feed', (
    tester,
  ) async {
    final server = FakeHomeServer()..live = [livestreamJson(title: 'Before')];
    final app = await pumpHome(tester, server);

    server.live = [livestreamJson(title: 'After')];
    for (final state in [
      AppLifecycleState.inactive,
      AppLifecycleState.hidden,
      AppLifecycleState.paused,
      AppLifecycleState.hidden,
      AppLifecycleState.inactive,
      AppLifecycleState.resumed,
    ]) {
      tester.binding.handleAppLifecycleStateChanged(state);
    }
    await tester.pumpAndSettle();

    expect(find.text('After'), findsOneWidget);
    expect(app.requestsTo(ApiEndpoints.popularLivestreams), hasLength(2));
  });

  testWidgets('cards of the same owner share one lookup', (tester) async {
    final server = FakeHomeServer()
      ..live = [livestreamJson(id: 'a'), livestreamJson(id: 'b')]
      ..vods = [vodJson(id: 'c'), vodJson(id: 'd')];
    final app = await pumpHome(tester, server);

    expect(find.text('streamer'), findsNWidgets(4));
    expect(app.requestsTo('${ApiEndpoints.user}/u-1'), hasLength(1));
  });

  testWidgets('a failed owner lookup still shows the card', (tester) async {
    final server = FakeHomeServer()
      ..users = {}
      ..live = [livestreamJson(title: 'Orphan stream')];
    await pumpHome(tester, server);

    expect(find.text('Orphan stream'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('very long Vietnamese and emoji texts do not overflow', (
    tester,
  ) async {
    final long = 'Tiếng Việt rất dài 🎉🔥 ' * 30;
    final server = FakeHomeServer()
      ..users = {'u-1': publicUserJson(username: 'Người_dùng_🎉${'x' * 200}')}
      ..live = [livestreamJson(title: long)]
      ..vods = [vodJson(title: long)]
      ..following = [publicUserJson(id: 'u-2', username: 'y' * 200)];
    await pumpHome(tester, server);

    expect(tester.takeException(), isNull);
    expect(find.textContaining('Tiếng Việt rất dài'), findsNWidgets(2));
  });
}
