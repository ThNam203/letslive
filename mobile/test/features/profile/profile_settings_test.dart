import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/router/app_router.dart';
import 'package:letslive_mobile/features/auth/auth_providers.dart';
import 'package:letslive_mobile/features/auth/presentation/login_screen.dart';

import 'profile_test_app.dart';

final _save = find.byKey(const ValueKey('profile-save'));
final _saveButton = find.descendant(
  of: _save,
  matching: find.byType(FilledButton),
);

/// The own profile: its tab, the settings hub and the edit-profile screen.
void main() {
  group('own profile tab', () {
    testWidgets('shows the signed-in user without a follow button', (
      tester,
    ) async {
      final server = FakeProfileServer()..channelVods = vodsOf(0);
      await pumpRoute(tester, server, AppRoutes.profile);

      expect(find.text('Profile'), findsWidgets);
      expect(find.text('abcdef'), findsOneWidget);
      expect(find.byKey(const ValueKey('follow-button')), findsNothing);
      expect(find.widgetWithText(Tab, 'Chat'), findsNothing);
      expect(find.widgetWithText(Tab, 'Videos'), findsOneWidget);
    });

    testWidgets('is the last tab and opens settings from its app bar', (
      tester,
    ) async {
      final app = await pumpRoute(tester, FakeProfileServer(), AppRoutes.home);
      expect(app.container.read(appRouterProvider), isNotNull);

      await tester.tap(
        find.descendant(
          of: find.byType(NavigationBar),
          matching: find.text('Profile'),
        ),
      );
      await tester.pumpAndSettle();
      expect(find.text('abcdef'), findsOneWidget);

      await tester.tap(find.byKey(const ValueKey('open-settings')));
      await tester.pumpAndSettle();
      expect(find.text('Settings'), findsOneWidget);
    });

    testWidgets('a failed load offers retry', (tester) async {
      final server = FakeProfileServer();
      server.users.remove('u-1');
      await pumpRoute(tester, server, AppRoutes.profile);
      expect(find.text('Try again'), findsOneWidget);

      server.users['u-1'] = {
        'id': 'u-1',
        'username': 'abcdef',
        'followerCount': 0,
        'createdAt': '2026-01-15T00:00:00Z',
      };
      await tester.tap(find.text('Try again'));
      await tester.pumpAndSettle();
      expect(find.text('abcdef'), findsOneWidget);
    });

    testWidgets('settings hub opens edit profile and security', (tester) async {
      await pumpRoute(tester, FakeProfileServer(), AppRoutes.settings);

      await tester.tap(find.byKey(const ValueKey('settings-profile')));
      await tester.pumpAndSettle();
      expect(find.text('Profile Settings'), findsWidgets);
      await tester.pageBack();
      await tester.pumpAndSettle();

      await tester.tap(find.byKey(const ValueKey('settings-security')));
      await tester.pumpAndSettle();
      expect(find.text('user@example.com'), findsOneWidget);
    });

    testWidgets('log out ends the session and shows login', (tester) async {
      await pumpRoute(tester, FakeProfileServer(), AppRoutes.settings);

      await tester.tap(find.byKey(const ValueKey('settings-logout')));
      await tester.pumpAndSettle();

      expect(find.byType(LoginScreen), findsOneWidget);
    });
  });

  group('edit profile', () {
    testWidgets(
      'username below 6 chars is rejected locally with web\'s message',
      (tester) async {
        final server = FakeProfileServer();
        await pumpRoute(tester, server, AppRoutes.settingsProfile);

        await tester.enterText(field('username'), 'short');
        await tester.pump();
        await tester.tap(_save);
        await tester.pumpAndSettle();

        expect(find.text('Username must be >= 6 characters'), findsOneWidget);
        expect(server.profileUpdates, isEmpty);
      },
    );

    testWidgets('saving updates the session user and sends only changes', (
      tester,
    ) async {
      final server = FakeProfileServer();
      final app = await pumpRoute(tester, server, AppRoutes.settingsProfile);
      expect(app.container.read(sessionProvider).value!.user.bio, isNull);

      await tester.enterText(field('bio'), 'New bio ✨');
      await tester.enterText(
        field('social-github'),
        'https://github.com/octocat',
      );
      await tester.pump();
      await tester.tap(_save);
      await tester.pumpAndSettle();

      expect(server.profileUpdates, [
        {
          'bio': 'New bio ✨',
          'socialMediaLinks': {'github': 'https://github.com/octocat'},
        },
      ]);
      final me = app.container.read(sessionProvider).value!.user;
      expect(me.bio, 'New bio ✨');
      expect(me.socialMediaLinks, {'github': 'https://github.com/octocat'});
      expect(
        find.text('Profile information updated successfully!'),
        findsOneWidget,
      );
    });

    testWidgets('a new username reaches the session and the own profile', (
      tester,
    ) async {
      final server = FakeProfileServer();
      final app = await pumpRoute(tester, server, AppRoutes.settingsProfile);

      await tester.enterText(field('username'), 'renamed_user');
      await tester.pump();
      await tester.tap(_save);
      await tester.pumpAndSettle();

      expect(
        app.container.read(sessionProvider).value!.user.username,
        'renamed_user',
      );
      expect(server.profileUpdates.single, {'username': 'renamed_user'});
    });

    testWidgets('a server refusal is shown and the session is unchanged', (
      tester,
    ) async {
      final server = FakeProfileServer();
      final app = await pumpRoute(tester, server, AppRoutes.settingsProfile);

      await tester.enterText(field('username'), 'taken_name');
      await tester.pump();
      await tester.tap(_save);
      await tester.pumpAndSettle();

      expect(find.text('Username already taken.'), findsOneWidget);
      expect(
        app.container.read(sessionProvider).value!.user.username,
        'abcdef',
      );
    });

    testWidgets('emptying a saved social link is refused, not ignored', (
      tester,
    ) async {
      final server = FakeProfileServer()
        ..meExtra = {
          'socialMediaLinks': {'github': 'https://github.com/octocat'},
        };
      await pumpRoute(tester, server, AppRoutes.settingsProfile);

      await tester.enterText(field('social-github'), '');
      await tester.pump();
      await tester.tap(_save);
      await tester.pumpAndSettle();

      expect(find.text("Link can't be empty"), findsOneWidget);
      expect(server.profileUpdates, isEmpty);
    });

    testWidgets('save is disabled until something changed', (tester) async {
      await pumpRoute(tester, FakeProfileServer(), AppRoutes.settingsProfile);
      expect(tester.widget<FilledButton>(_saveButton).onPressed, isNull);

      await tester.enterText(field('bio'), 'x');
      await tester.pump();
      expect(tester.widget<FilledButton>(_saveButton).onPressed, isNotNull);
    });

    testWidgets('an invalid social link is rejected locally', (tester) async {
      final server = FakeProfileServer();
      await pumpRoute(tester, server, AppRoutes.settingsProfile);

      await tester.enterText(field('social-website'), 'not a link');
      await tester.pump();
      await tester.tap(_save);
      await tester.pumpAndSettle();

      expect(find.text('Invalid link, please try again'), findsOneWidget);
      expect(server.profileUpdates, isEmpty);
    });

    testWidgets('long names and bios fit 320px at text scale 2', (
      tester,
    ) async {
      final server = FakeProfileServer()
        ..username = 'tên_rất_dài_${'x' * 15}'
        ..meExtra = {
          'bio': 'Xin chào 🎉 ${'tiểu sử dài ' * 20}',
          'socialMediaLinks': {'website': 'https://example.com/${'p' * 200}'},
        };
      await pumpRoute(
        tester,
        server,
        AppRoutes.settingsProfile,
        width: 320,
        height: 700,
        textScale: 2,
      );

      expect(tester.takeException(), isNull);
      await tester.drag(find.byType(ListView), const Offset(0, -2000));
      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull);
    });
  });

  group('pictures', () {
    testWidgets('the pictures themselves are the change buttons, like web', (
      tester,
    ) async {
      await pumpRoute(tester, FakeProfileServer(), AppRoutes.settingsProfile);

      // No separate text buttons: each picture says what a tap does.
      expect(find.byType(OutlinedButton), findsNothing);
      expect(find.byTooltip('Update background'), findsOneWidget);
      expect(find.byTooltip('Update profile picture'), findsOneWidget);
      expect(
        find.descendant(
          of: field('pick-avatar'),
          matching: find.byType(InkWell),
          matchRoot: true,
        ),
        findsOneWidget,
      );
      // Both pictures carry the camera badge.
      expect(find.byIcon(Icons.photo_camera_outlined), findsNWidgets(2));
    });

    testWidgets('image upload sends multipart and shows progress', (
      tester,
    ) async {
      final server = FakeProfileServer()..holdUpload = Completer<void>();
      final device = FakeDevice()..next = pickedImage(name: 'me.jpg');
      final app = await pumpRoute(
        tester,
        server,
        AppRoutes.settingsProfile,
        device: device,
      );

      await tester.tap(field('pick-avatar'));
      await tester.pumpAndSettle();
      await tester.tap(_save);
      // The busy button spins, so pumpAndSettle would never return.
      for (var i = 0; i < 5; i++) {
        await tester.pump(const Duration(milliseconds: 50));
      }

      // Half sent, per the progress the fake server reports.
      final bar = tester.widget<LinearProgressIndicator>(
        field('upload-progress'),
      );
      expect(bar.value, 0.5);
      expect(find.text('Uploading...'), findsOneWidget);

      server.holdUpload!.complete();
      await tester.pumpAndSettle();

      final request = app.requestsTo(ApiEndpoints.profilePicture).single;
      expect(request.method, 'PATCH');
      final file = (request.data as FormData).files.single;
      expect(file.key, 'profile-picture');
      expect(file.value.filename, 'me.jpg');
      expect(file.value.contentType?.mimeType, 'image/jpeg');
      expect(field('upload-progress'), findsNothing);
      expect(
        app.container.read(sessionProvider).value!.user.profilePicture,
        uploadedUrl,
      );
    });

    testWidgets('the banner goes to its own field', (tester) async {
      final server = FakeProfileServer();
      final device = FakeDevice()
        ..next = pickedImage(name: 'banner.png', mimeType: 'image/png');
      final app = await pumpRoute(
        tester,
        server,
        AppRoutes.settingsProfile,
        device: device,
      );

      await tester.tap(field('pick-background'));
      await tester.pumpAndSettle();
      await tester.tap(_save);
      await tester.pumpAndSettle();

      final file =
          (app.requestsTo(ApiEndpoints.backgroundPicture).single.data
                  as FormData)
              .files
              .single;
      expect(file.key, 'background-picture');
      expect(
        app.container.read(sessionProvider).value!.user.backgroundPicture,
        uploadedUrl,
      );
    });

    testWidgets('a file over 10 MB is refused before any upload', (
      tester,
    ) async {
      final device = FakeDevice()
        ..next = pickedImage(bytes: 10 * 1024 * 1024 + 1);
      final app = await pumpRoute(
        tester,
        FakeProfileServer(),
        AppRoutes.settingsProfile,
        device: device,
      );

      await tester.tap(field('pick-avatar'));
      await tester.pumpAndSettle();

      expect(find.text('File size exceeds 10 MB'), findsOneWidget);
      await tester.tap(_save);
      await tester.pumpAndSettle();
      expect(app.requestsTo(ApiEndpoints.profilePicture), isEmpty);
    });

    testWidgets('a photo library that will not open says so', (tester) async {
      final device = FakeDevice()..pickFails = true;
      await pumpRoute(
        tester,
        FakeProfileServer(),
        AppRoutes.settingsProfile,
        device: device,
      );

      await tester.tap(field('pick-avatar'));
      await tester.pumpAndSettle();

      expect(find.text('Could not open your photos'), findsOneWidget);
    });

    testWidgets('a profile picture of an unsupported type is refused', (
      tester,
    ) async {
      final device = FakeDevice()
        ..next = pickedImage(name: 'a.png', mimeType: 'image/png');
      await pumpRoute(
        tester,
        FakeProfileServer(),
        AppRoutes.settingsProfile,
        device: device,
      );

      await tester.tap(field('pick-avatar'));
      await tester.pumpAndSettle();

      expect(find.text('Unsupported image, use JPEG or WebP.'), findsOneWidget);
    });

    testWidgets('a profile picture outside 80..10000 px is refused', (
      tester,
    ) async {
      final device = FakeDevice()
        ..next = pickedImage()
        ..size = (width: 79, height: 400);
      await pumpRoute(
        tester,
        FakeProfileServer(),
        AppRoutes.settingsProfile,
        device: device,
      );

      await tester.tap(field('pick-avatar'));
      await tester.pumpAndSettle();

      expect(
        find.text('Photo must be between 80x80 and 10000x10000 pixels.'),
        findsOneWidget,
      );
    });

    testWidgets('cancelling the picker changes nothing', (tester) async {
      await pumpRoute(
        tester,
        FakeProfileServer(),
        AppRoutes.settingsProfile,
        device: FakeDevice(),
      );

      await tester.tap(field('pick-avatar'));
      await tester.pumpAndSettle();

      expect(tester.takeException(), isNull);
      expect(find.byType(SnackBar), findsNothing);
    });

    testWidgets('a server refusal of the picture shows its message', (
      tester,
    ) async {
      final server = FakeProfileServer()..uploadStatus = 400;
      final device = FakeDevice()..next = pickedImage();
      final app = await pumpRoute(
        tester,
        server,
        AppRoutes.settingsProfile,
        device: device,
      );

      await tester.tap(field('pick-avatar'));
      await tester.pumpAndSettle();
      await tester.tap(_save);
      await tester.pumpAndSettle();

      expect(find.text('Unsupported image, use JPEG or WebP.'), findsOneWidget);
      expect(
        app.container.read(sessionProvider).value!.user.profilePicture,
        isNull,
      );
    });
  });
}
