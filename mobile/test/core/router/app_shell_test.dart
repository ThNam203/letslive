import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/core/realtime/realtime_manager.dart';
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

    await tester.tap(find.text('Two'));
    await tester.pumpAndSettle();

    expect(find.text('screen Two'), findsOneWidget);
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
