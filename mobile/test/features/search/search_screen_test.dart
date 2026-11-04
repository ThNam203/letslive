import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/features/search/presentation/search_result_rows.dart';

import '../home/home_fixtures.dart';
import 'search_test_app.dart';

void main() {
  List<Map<String, Object?>> manyLives(int n) => [
    for (var i = 0; i < n; i++)
      livestreamJson(id: 'ls-$i', title: 'Trò chơi live $i'),
  ];
  List<Map<String, Object?>> manyVods(int n) => [
    for (var i = 0; i < n; i++) vodJson(id: 'v-$i', title: 'Trò chơi vod $i'),
  ];
  List<Map<String, Object?>> manyUsers(int n) => [
    for (var i = 0; i < n; i++) publicUserJson(id: 'c-$i', username: 'chan$i'),
  ];

  testWidgets('an empty query shows the prompt and sends nothing', (
    tester,
  ) async {
    final app = await pumpSearch(tester, FakeSearchServer());

    expect(find.text('Search for channels, live streams and videos'), findsOne);
    await search(tester, '   ');
    expect(find.text('Search for channels, live streams and videos'), findsOne);
    expect(app.requestsTo(ApiEndpoints.searchVods), isEmpty);
  });

  testWidgets('All tab caps at 5 per type', (tester) async {
    final server = FakeSearchServer()
      ..searchUsers = manyUsers(8)
      ..searchLives = manyLives(8)
      ..searchVods = manyVods(8);
    final app = await pumpSearch(tester, server);

    await search(tester, 'tro');

    for (final path in [
      ApiEndpoints.searchUsers,
      ApiEndpoints.searchLivestreams,
      ApiEndpoints.searchVods,
    ]) {
      final requests = app.requestsTo(path);
      expect(requests, hasLength(1), reason: path);
      expect(requests.single.queryParameters['limit'], 5, reason: path);
      expect(requests.single.queryParameters['page'], 0, reason: path);
    }
    expect(find.byType(UserResultRow), findsNWidgets(5));
    expect(find.byType(LiveResultRow), findsNWidgets(5));
    expect(find.byType(VodResultRow), findsNWidgets(5));
    expect(find.text('See all'), findsNWidgets(3));
  });

  testWidgets('See all opens that type tab', (tester) async {
    final server = FakeSearchServer()..searchVods = manyVods(8);
    await pumpSearch(tester, server);
    await search(tester, 'tro');

    await tester.tap(find.text('See all').first);
    await tester.pumpAndSettle();

    expect(find.textContaining('vod 7'), findsOne);
  });

  testWidgets('a failing group shows its retry, the others still render', (
    tester,
  ) async {
    final server = FakeSearchServer()
      ..searchUsers = manyUsers(1)
      ..searchLivesStatus = 503
      ..searchVods = manyVods(1);
    final app = await pumpSearch(tester, server);

    await search(tester, 'tro');

    expect(find.text('chan0'), findsOne);
    expect(find.textContaining('vod 0'), findsOne);
    expect(find.text('Try again'), findsOne);

    server
      ..searchLivesStatus = 200
      ..searchLives = manyLives(1);
    await tester.tap(find.text('Try again'));
    await tester.pumpAndSettle();

    expect(find.textContaining('live 0'), findsOne);
    expect(find.text('Try again'), findsNothing);
    expect(app.requestsTo(ApiEndpoints.searchLivestreams), hasLength(2));
  });

  testWidgets('nothing found shows the no-results message', (tester) async {
    await pumpSearch(tester, FakeSearchServer());

    await search(tester, 'zzz');

    expect(find.text('No results for “zzz”'), findsOne);
    expect(find.text('Check the spelling or try a shorter search.'), findsOne);
  });

  testWidgets('type tab loads next page on scroll end', (tester) async {
    final server = FakeSearchServer()..searchVods = manyVods(25);
    // Short viewport: the list is lazy and only pages in near the end.
    final app = await pumpSearch(tester, server, height: 700);
    await search(tester, 'tro');
    await tester.tap(
      find.descendant(of: find.byType(TabBar), matching: find.text('Videos')),
    );
    await tester.pumpAndSettle();

    final vods = app.requestsTo(ApiEndpoints.searchVods);
    expect(vods.last.queryParameters, {'q': 'tro', 'page': 0, 'limit': 20});
    expect(find.textContaining('vod 24'), findsNothing);

    await tester.drag(find.byType(Scrollable).last, const Offset(0, -20000));
    await tester.pumpAndSettle();

    expect(
      app.requestsTo(ApiEndpoints.searchVods).last.queryParameters['page'],
      1,
    );
    // The page that arrived is below the old end of the list.
    await tester.drag(find.byType(Scrollable).last, const Offset(0, -20000));
    await tester.pumpAndSettle();
    expect(find.textContaining('vod 24'), findsOne);
  });

  testWidgets('a failed next page keeps the list and retries inline', (
    tester,
  ) async {
    final server = FakeSearchServer()
      ..searchLives = manyLives(25)
      ..failSearchPage = 1;
    await pumpSearch(tester, server, height: 700);
    await search(tester, 'tro');
    await tester.tap(
      find.descendant(of: find.byType(TabBar), matching: find.text('Live')),
    );
    await tester.pumpAndSettle();

    await tester.drag(find.byType(Scrollable).last, const Offset(0, -20000));
    await tester.pumpAndSettle();
    expect(find.text('Try again'), findsOne);

    server.failSearchPage = null;
    await tester.tap(find.text('Try again'));
    await tester.pumpAndSettle();
    await tester.drag(find.byType(Scrollable).last, const Offset(0, -20000));
    await tester.pumpAndSettle();

    expect(find.textContaining('live 24'), findsOne);
  });

  testWidgets('a type tab with no results says so', (tester) async {
    await pumpSearch(tester, FakeSearchServer());
    await search(tester, 'zzz');

    await tester.tap(
      find.descendant(of: find.byType(TabBar), matching: find.text('Channels')),
    );
    await tester.pumpAndSettle();

    expect(find.text('No channels found'), findsOne);
  });

  testWidgets('a type tab error offers retry', (tester) async {
    final server = FakeSearchServer()..searchUsersStatus = 503;
    await pumpSearch(tester, server);
    await search(tester, 'tro');
    await tester.tap(
      find.descendant(of: find.byType(TabBar), matching: find.text('Channels')),
    );
    await tester.pumpAndSettle();
    expect(find.text('Try again'), findsOne);

    server
      ..searchUsersStatus = 200
      ..searchUsers = manyUsers(1);
    await tester.tap(find.text('Try again'));
    await tester.pumpAndSettle();

    expect(find.text('chan0'), findsOne);
  });

  testWidgets(
    'accent-less query "tro choi" renders server results with highlight',
    (tester) async {
      final server = FakeSearchServer()
        ..searchVods = [vodJson(title: 'Trò chơi vui')];
      final app = await pumpSearch(tester, server);

      await search(tester, 'tro choi');

      expect(
        app.requestsTo(ApiEndpoints.searchVods).single.queryParameters['q'],
        'tro choi',
      );
      expect(highlightedTexts(tester), contains('Trò chơi'));
      expect(find.byType(Text), findsWidgets);
    },
  );

  testWidgets('the query is trimmed and sent once after the debounce', (
    tester,
  ) async {
    final server = FakeSearchServer();
    final app = await pumpSearch(tester, server);
    final input = find.byKey(const ValueKey('search-input'));

    await tester.enterText(input, 't');
    await tester.pump(const Duration(milliseconds: 100));
    await tester.enterText(input, '  tro  ');
    await tester.pump(const Duration(milliseconds: 299));
    expect(app.requestsTo(ApiEndpoints.searchVods), isEmpty);
    await tester.pump(const Duration(milliseconds: 1));
    await tester.pumpAndSettle();

    final requests = app.requestsTo(ApiEndpoints.searchVods);
    expect(requests, hasLength(1));
    expect(requests.single.queryParameters['q'], 'tro');
  });

  testWidgets('a slow answer to an old query never replaces the new one', (
    tester,
  ) async {
    final server = FakeSearchServer()..searchVods = [vodJson(title: 'answer')];
    await pumpSearch(tester, server);
    final slow = Completer<void>();
    server.holdSearch = slow.future;
    await tester.enterText(find.byKey(const ValueKey('search-input')), 'old');
    await tester.pump(const Duration(milliseconds: 300));
    await tester.pump();
    // 'old' is still waiting; 'new' is answered at once.
    server
      ..holdSearch = null
      ..searchVods = [vodJson(title: 'Newest result')];
    await search(tester, 'new');
    slow.complete();
    await tester.pumpAndSettle();

    expect(find.text('Newest result'), findsOne);
    expect(find.text('answer'), findsNothing);
  });

  testWidgets('clearing the box goes back to the prompt', (tester) async {
    final server = FakeSearchServer()..searchVods = manyVods(1);
    await pumpSearch(tester, server);
    await search(tester, 'tro');
    expect(find.textContaining('vod 0'), findsOne);

    await tester.tap(find.byTooltip('Clear search'));
    await tester.pumpAndSettle();

    expect(find.textContaining('vod 0'), findsNothing);
    expect(find.text('Search for channels, live streams and videos'), findsOne);
  });

  testWidgets('very long query and result texts do not overflow', (
    tester,
  ) async {
    final longQuery = 'Tiếng Việt rất dài 🎉 ' * 20;
    final longTitle = 'Tiếng Việt rất dài 🎉 ' * 30;
    final server = FakeSearchServer()
      ..users = {'u-1': publicUserJson(username: 'Người_dùng_🎉${'x' * 200}')}
      ..searchUsers = [
        publicUserJson(id: 'c-1', username: 'Tên_kênh_rất_dài_${'y' * 200}'),
      ]
      ..searchLives = [livestreamJson(title: longTitle)]
      ..searchVods = [vodJson(title: longTitle)];
    await pumpSearch(tester, server, width: 320, height: 3000, textScale: 2);

    await search(tester, longQuery);

    expect(find.textContaining('Tiếng Việt rất dài'), findsWidgets);
    expect(tester.takeException(), isNull);

    for (final tab in ['Channels', 'Live', 'Videos']) {
      await openTab(tester, tab);
      expect(find.byType(ListView), findsWidgets, reason: tab);
      expect(tester.takeException(), isNull, reason: tab);
    }
  });
}
