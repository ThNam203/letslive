import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/router/app_router.dart';
import 'package:letslive_mobile/features/auth/presentation/login_screen.dart';

import 'profile_test_app.dart';

Finder _key(String key) => find.byKey(ValueKey(key));

/// Account settings ported from web: disable the account
/// (`disable-account-dialog.tsx`) and the phone number (`phone-number.tsx`).
void main() {
  group('disable account', () {
    Future<void> openDialog(WidgetTester tester) async {
      // The section is last; a small screen builds it only when scrolled to.
      await tester.scrollUntilVisible(
        _key('disable-account'),
        300,
        scrollable: find.byType(Scrollable).first,
      );
      await tester.ensureVisible(_key('disable-account'));
      await tester.pumpAndSettle();
      await tester.tap(_key('disable-account'));
      await tester.pumpAndSettle();
      expect(find.text('Disable your account!'), findsOneWidget);
    }

    testWidgets('confirming disables the account, then logs out', (
      tester,
    ) async {
      final server = FakeProfileServer();
      await pumpRoute(tester, server, AppRoutes.settingsProfile);

      await openDialog(tester);
      await tester.tap(find.text('I understand!'));
      await tester.pumpAndSettle();

      expect(server.profileUpdates.single, {'status': 'disabled'});
      expect(find.byType(LoginScreen), findsOneWidget);
    });

    testWidgets('cancel sends nothing', (tester) async {
      final server = FakeProfileServer();
      await pumpRoute(tester, server, AppRoutes.settingsProfile);

      await openDialog(tester);
      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();

      expect(server.profileUpdates, isEmpty);
      expect(find.byType(LoginScreen), findsNothing);
    });

    testWidgets('a server failure keeps the session and says why', (
      tester,
    ) async {
      final server = FakeProfileServer()..profileUpdateStatus = 500;
      await pumpRoute(tester, server, AppRoutes.settingsProfile);

      await openDialog(tester);
      await tester.tap(find.text('I understand!'));
      await tester.pumpAndSettle();

      expect(find.byType(LoginScreen), findsNothing);
      expect(find.byType(SnackBar), findsOneWidget);
    });

    testWidgets('fits 320px at text scale 2, dialog included', (tester) async {
      await pumpRoute(
        tester,
        FakeProfileServer(),
        AppRoutes.settingsProfile,
        width: 320,
        height: 568,
        textScale: 2,
      );
      await openDialog(tester);
      expect(tester.takeException(), isNull);
    });
  });

  group('phone number', () {
    Future<void> edit(WidgetTester tester, String text) async {
      await tester.tap(_key('phone-number'));
      await tester.pumpAndSettle();
      await tester.enterText(_key('phone-number-field'), text);
      await tester.tap(_key('phone-number-save'));
      await tester.pumpAndSettle();
    }

    testWidgets('a valid number is saved and shown', (tester) async {
      final server = FakeProfileServer();
      await pumpRoute(tester, server, AppRoutes.settingsSecurity);
      expect(find.text('Add a number'), findsOneWidget);

      await edit(tester, ' +84 912 345 678 ');

      expect(server.profileUpdates.single, {'phoneNumber': '+84 912 345 678'});
      expect(find.text('+84 912 345 678'), findsOneWidget);
    });

    testWidgets('an invalid number is refused before sending', (tester) async {
      final server = FakeProfileServer();
      await pumpRoute(tester, server, AppRoutes.settingsSecurity);

      await edit(tester, '12ab');

      expect(
        find.text('Invalid phone number, please try again'),
        findsOneWidget,
      );
      expect(server.profileUpdates, isEmpty);
    });

    testWidgets('an empty number removes it, like web', (tester) async {
      final server = FakeProfileServer()
        ..meExtra = {'phoneNumber': '+84912345678'};
      await pumpRoute(tester, server, AppRoutes.settingsSecurity);
      expect(find.text('+84912345678'), findsOneWidget);

      await edit(tester, '');

      expect(server.profileUpdates.single, {'phoneNumber': ''});
    });

    testWidgets('fits 320px at text scale 2 with the editor open', (
      tester,
    ) async {
      await pumpRoute(
        tester,
        FakeProfileServer(),
        AppRoutes.settingsSecurity,
        width: 320,
        height: 568,
        textScale: 2,
      );
      expect(tester.takeException(), isNull);
      await tester.tap(_key('phone-number'));
      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull);
    });
  });
}
