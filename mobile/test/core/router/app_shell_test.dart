import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/core/realtime/realtime_manager.dart';
import 'package:letslive_mobile/core/router/app_chrome.dart';
import 'package:letslive_mobile/core/router/app_shell.dart';
import 'package:letslive_mobile/core/router/app_tabs.dart';
import 'package:mocktail/mocktail.dart';

class _MockRealtime extends Mock implements RealtimeManager {}

AppTab _tab(String path, String label) => AppTab(
  path: path,
  icon: Icons.circle,
  label: (_) => label,
  builder: (_) => Center(child: Text('screen $label')),
);

/// A tab whose root is a long list under its own top bar.
AppTab _listTab(String path, String label) => AppTab(
  path: path,
  icon: Icons.list,
  label: (_) => label,
  builder: (_) => TabScaffold(
    appBar: AppBar(title: Text('bar $label')),
    body: ListView.builder(
      itemCount: 100,
      itemBuilder: (_, i) => SizedBox(height: 60, child: Text('$label $i')),
    ),
  ),
);

/// Visible height of the bar around [finder] (zero once it slid away).
double _shown(WidgetTester tester, Finder finder) => tester
    .getSize(find.ancestor(of: finder, matching: find.byType(CollapsibleBar)))
    .height;

Future<_MockRealtime> _pump(WidgetTester tester, List<AppTab> tabs) async {
  final realtime = _MockRealtime();
  final router = GoRouter(routes: [AppShell.route(tabs)]);
  addTearDown(router.dispose);
  await tester.pumpWidget(
    ProviderScope(
      overrides: [realtimeManagerProvider.overrideWithValue(realtime)],
      child: MaterialApp.router(
        routerConfig: router,
        locale: const Locale('en'),
        supportedLocales: AppLocalizations.supportedLocales,
        localizationsDelegates: const [
          AppLocalizations.delegate,
          GlobalMaterialLocalizations.delegate,
          GlobalWidgetsLocalizations.delegate,
        ],
      ),
    ),
  );
  return realtime;
}

void main() {
  testWidgets('the navigation bar is hidden with fewer than two tabs', (
    tester,
  ) async {
    await _pump(tester, [_tab('/', 'One')]);

    expect(find.text('screen One'), findsOneWidget);
    expect(find.byType(NavigationBar), findsNothing);
  });

  testWidgets('with two tabs the bar switches between them', (tester) async {
    await _pump(tester, [_tab('/', 'One'), _tab('/two', 'Two')]);

    expect(find.byType(NavigationBar), findsOneWidget);
    expect(find.text('screen One'), findsOneWidget);

    await tester.tap(find.byTooltip('Two'));
    await tester.pumpAndSettle();

    expect(find.text('screen Two'), findsOneWidget);
  });

  testWidgets('the bar shows icons only, the label as tooltip', (tester) async {
    await _pump(tester, [_tab('/', 'One'), _tab('/two', 'Two')]);

    final bar = tester.widget<NavigationBar>(find.byType(NavigationBar));
    expect(bar.labelBehavior, NavigationDestinationLabelBehavior.alwaysHide);
    expect(find.byTooltip('Two'), findsOneWidget);
  });

  testWidgets('scrolling down hides both bars, scrolling up brings them '
      'back, like YouTube', (tester) async {
    await _pump(tester, [_listTab('/', 'One'), _tab('/two', 'Two')]);
    final topBar = find.text('bar One');
    final bottomBar = find.byType(NavigationBar);
    expect(_shown(tester, topBar), greaterThan(0));
    expect(_shown(tester, bottomBar), greaterThan(0));

    await tester.drag(find.text('One 3'), const Offset(0, -300));
    await tester.pumpAndSettle();
    expect(_shown(tester, topBar), 0);
    expect(_shown(tester, bottomBar), 0);

    await tester.drag(find.text('One 8'), const Offset(0, 100));
    await tester.pumpAndSettle();
    expect(_shown(tester, topBar), greaterThan(0));
    expect(_shown(tester, bottomBar), greaterThan(0));
  });

  test('the registry starts with Home and has unique tab paths', () {
    final paths = appTabs.map((t) => t.path).toList();
    expect(paths.first, '/');
    expect(paths.toSet(), hasLength(paths.length));
  });

  testWidgets('realtime pauses in the background and resumes on return', (
    tester,
  ) async {
    final realtime = await _pump(tester, [_tab('/', 'One')]);

    for (final state in [
      AppLifecycleState.inactive,
      AppLifecycleState.hidden,
      AppLifecycleState.paused,
    ]) {
      tester.binding.handleAppLifecycleStateChanged(state);
    }
    verify(realtime.pause).called(1);
    verifyNever(realtime.resume);

    for (final state in [
      AppLifecycleState.hidden,
      AppLifecycleState.inactive,
      AppLifecycleState.resumed,
    ]) {
      tester.binding.handleAppLifecycleStateChanged(state);
    }
    verify(realtime.resume).called(1);
  });
}
