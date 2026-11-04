import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/misc.dart' show Override;
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';

import '../../core/network/fake_adapter.dart';
import '../auth/auth_test_app.dart';
import '../home/home_test_app.dart';

/// [FakeHomeServer] plus the three search endpoints. Each endpoint answers the
/// page of its list that `page`/`limit` ask for, whatever the query; the
/// requests are recorded by the app's adapter. A `*Status` other than 200
/// makes that endpoint fail.
class FakeSearchServer extends FakeHomeServer {
  List<Map<String, Object?>> searchUsers = [];
  List<Map<String, Object?>> searchLives = [];
  List<Map<String, Object?>> searchVods = [];
  int searchUsersStatus = 200;
  int searchLivesStatus = 200;
  int searchVodsStatus = 200;

  /// Page (0-based) of every search endpoint that answers 503, or null.
  int? failSearchPage;

  /// When set, search requests wait for it before answering.
  Future<void>? holdSearch;

  @override
  Future<ResponseBody> handle(RequestOptions o) async {
    final (list, status) = switch (o.path) {
      ApiEndpoints.searchUsers => (searchUsers, searchUsersStatus),
      ApiEndpoints.searchLivestreams => (searchLives, searchLivesStatus),
      ApiEndpoints.searchVods => (searchVods, searchVodsStatus),
      _ => (null, 200),
    };
    if (list == null) return super.handle(o);
    await holdSearch;
    final page = o.queryParameters['page'] as int;
    final limit = o.queryParameters['limit'] as int;
    if (status != 200 || page == failSearchPage) {
      return jsonBody({
        'requestId': 'req-err',
        'success': false,
        'code': 1,
        'key': 'res_err_internal_server',
        'message': 'Server down',
      }, status: 503);
    }
    return okEnvelope(list.skip(page * limit).take(limit).toList());
  }
}

/// Pumps the signed-in app against [server] and opens the Search tab.
Future<TestApp> pumpSearch(
  WidgetTester tester,
  FakeSearchServer server, {
  double width = 800,
  double height = 6000,
  double textScale = 1,
  List<Override> overrides = const [],
}) async {
  final app = await pumpHome(
    tester,
    server,
    width: width,
    height: height,
    textScale: textScale,
    overrides: overrides,
  );
  await tester.tap(
    find.descendant(
      of: find.byType(NavigationBar),
      matching: find.text('Search'),
    ),
  );
  await tester.pumpAndSettle();
  return app;
}

/// Types [text] into the search box and waits for the debounce and the
/// answers.
Future<void> search(WidgetTester tester, String text) async {
  await tester.enterText(find.byKey(const ValueKey('search-input')), text);
  await tester.pump(const Duration(milliseconds: 300));
  await tester.pumpAndSettle();
}

/// Opens the tab called [label] (the tab bar scrolls on a small phone).
Future<void> openTab(WidgetTester tester, String label) async {
  final tab = find.descendant(
    of: find.byType(TabBar),
    matching: find.text(label),
  );
  await tester.ensureVisible(tab);
  await tester.tap(tab);
  await tester.pumpAndSettle();
}

/// The parts of the texts on screen that are highlighted.
List<String> highlightedTexts(WidgetTester tester) => [
  for (final rich in tester.widgetList<RichText>(find.byType(RichText)))
    ..._marked(rich.text),
];

Iterable<String> _marked(InlineSpan span) sync* {
  if (span is! TextSpan) return;
  final text = span.text;
  if (text != null && span.style?.backgroundColor != null) yield text;
  for (final child in span.children ?? const <InlineSpan>[]) {
    yield* _marked(child);
  }
}
