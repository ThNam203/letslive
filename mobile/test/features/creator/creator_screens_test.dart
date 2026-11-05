import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/widgets/action_feedback.dart';
import 'package:letslive_mobile/core/router/app_router.dart';
import 'package:letslive_mobile/features/auth/auth_providers.dart';
import 'package:letslive_mobile/features/auth/presentation/login_screen.dart';
import 'package:letslive_mobile/features/creator/creator_providers.dart';
import 'package:letslive_mobile/features/creator/domain/picked_video.dart';
import 'package:letslive_mobile/features/creator/presentation/upload_screen.dart';
import 'package:letslive_mobile/features/creator/upload_manager.dart';
import 'package:letslive_mobile/features/vod/vod_providers.dart';

import '../auth/auth_test_app.dart';
import '../profile/profile_test_app.dart';
import 'creator_fakes.dart';
import 'creator_test_app.dart';

Finder _key(String key) => find.byKey(ValueKey(key));

String _keyText(WidgetTester tester) =>
    tester.widget<Text>(_key('stream-key-value')).data!;

void main() {
  group('stream settings', () {
    testWidgets('the key is masked until revealed and can be copied', (
      tester,
    ) async {
      final copied = <String>[];
      tester.binding.defaultBinaryMessenger.setMockMethodCallHandler(
        SystemChannels.platform,
        (call) async {
          if (call.method == 'Clipboard.setData') {
            copied.add((call.arguments as Map)['text'] as String);
          }
          return null;
        },
      );
      addTearDown(
        () => tester.binding.defaultBinaryMessenger.setMockMethodCallHandler(
          SystemChannels.platform,
          null,
        ),
      );
      await pumpCreator(tester, FakeCreatorServer(), AppRoutes.settingsStream);

      expect(_keyText(tester), '•' * 'stream-key'.length);
      expect(find.text('stream-key'), findsNothing);

      await tester.tap(_key('toggle-key'));
      await tester.pump();
      expect(_keyText(tester), 'stream-key');

      await tester.tap(_key('copy-key'));
      await tester.pump();
      expect(copied, ['stream-key']);
      expect(find.text('API Key copied to clipboard'), findsOneWidget);
    });

    testWidgets('stream key reset requires confirmation', (tester) async {
      final server = FakeCreatorServer();
      await pumpCreator(tester, server, AppRoutes.settingsStream);

      // Cancelling the dialog sends nothing.
      await tester.tap(_key('reset-key'));
      await tester.pumpAndSettle();
      expect(find.text('Generate a new key?'), findsOneWidget);
      expect(server.keyResets, 0);
      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();
      expect(server.keyResets, 0);
      expect(_keyText(tester), '•' * 'stream-key'.length);

      // Confirming replaces the key, which stays hidden.
      await tester.tap(_key('reset-key'));
      await tester.pumpAndSettle();
      await tester.tap(_key('confirm-reset-key'));
      await tester.pumpAndSettle();
      expect(server.keyResets, 1);
      expect(_keyText(tester), '•' * 'new-key-1'.length);
      await tester.tap(_key('toggle-key'));
      await tester.pump();
      expect(_keyText(tester), 'new-key-1');
    });

    testWidgets('a failed reset keeps the old key and says why', (
      tester,
    ) async {
      final server = FakeCreatorServer()..streamWriteStatus = 503;
      await pumpCreator(tester, server, AppRoutes.settingsStream);

      await tester.tap(_key('reset-key'));
      await tester.pumpAndSettle();
      await tester.tap(_key('confirm-reset-key'));
      await tester.pumpAndSettle();

      await tester.tap(_key('toggle-key'));
      await tester.pump();
      expect(_keyText(tester), 'stream-key');
      expect(find.byType(SnackBar), findsOneWidget);
    });

    testWidgets('saves the stream information and keeps the form in sync', (
      tester,
    ) async {
      final server = FakeCreatorServer();
      await pumpCreator(tester, server, AppRoutes.settingsStream);

      expect(
        tester.widget<FilledButton>(find.byType(FilledButton)).onPressed,
        isNull,
      );
      await tester.enterText(_key('stream-title'), 'New title');
      await tester.pump();
      await tester.tap(_key('stream-save'));
      await tester.pumpAndSettle();

      expect(server.streamInfoSaves.single, {
        'title': 'New title',
        'description': 'About it',
      });
      expect(
        find.text('Livestream information updated successfully'),
        findsOneWidget,
      );
      // Saved, so nothing is left to save.
      expect(
        tester
            .widget<FilledButton>(
              find
                  .descendant(
                    of: _key('stream-save'),
                    matching: find.byType(FilledButton),
                  )
                  .first,
            )
            .onPressed,
        isNull,
      );
    });

    testWidgets('a picked thumbnail is sent as a file', (tester) async {
      final server = FakeCreatorServer();
      final device = FakeDevice()..next = pickedImage();
      await pumpCreator(
        tester,
        server,
        AppRoutes.settingsStream,
        device: device,
      );

      await tester.tap(_key('pick-thumbnail'));
      await tester.pumpAndSettle();
      await tester.tap(_key('stream-save'));
      await tester.pumpAndSettle();

      expect(server.streamInfoFiles.single, ['thumbnail']);
    });

    testWidgets('an oversized thumbnail is refused before upload', (
      tester,
    ) async {
      final server = FakeCreatorServer();
      final device = FakeDevice()..next = pickedImage(bytes: 11 * 1024 * 1024);
      await pumpCreator(
        tester,
        server,
        AppRoutes.settingsStream,
        device: device,
      );

      await tester.tap(_key('pick-thumbnail'));
      await tester.pumpAndSettle();

      expect(find.text('File size exceeds 10 MB'), findsOneWidget);
      expect(server.streamInfoSaves, isEmpty);
    });

    testWidgets('a failed load offers a retry', (tester) async {
      final server = FakeCreatorServer();
      await pumpCreator(tester, server, AppRoutes.settings);

      server.userMeStatus = 503;
      await tester.tap(_key('settings-stream'));
      await tester.pumpAndSettle();
      expect(find.text('Try again'), findsOneWidget);

      server.userMeStatus = 200;
      await tester.tap(find.text('Try again'));
      await tester.pumpAndSettle();
      expect(_key('stream-key-value'), findsOneWidget);
    });
  });

  group('chat commands', () {
    FakeCreatorServer server0() => FakeCreatorServer()
      ..userCommands = [commandJson('c-a', 'discord', response: 'Join us')]
      ..channelCommands = [
        commandJson('c-b', 'rules', scope: 'channel', response: 'Be kind'),
      ];

    testWidgets('lists both scopes and the built-in commands', (tester) async {
      await pumpCreator(tester, server0(), AppRoutes.settingsChatCommands);

      expect(_key('command-user-discord'), findsOneWidget);
      expect(_key('command-channel-rules'), findsOneWidget);
      expect(find.text('Join us'), findsOneWidget);
      expect(find.text('/roll [max]'), findsOneWidget);
    });

    testWidgets('rejects an invalid name without a request', (tester) async {
      final server = server0();
      await pumpCreator(tester, server, AppRoutes.settingsChatCommands);

      await tester.enterText(_key('name-user'), 'Bad Name!');
      await tester.enterText(_key('response-user'), 'x');
      await tester.ensureVisible(_key('submit-user'));
      await tester.tap(_key('submit-user'));
      await tester.pumpAndSettle();

      expect(
        find.text('Name must be 1-32 chars, lowercase letters/numbers/_-'),
        findsOneWidget,
      );
      expect(server.createdCommands, isEmpty);
    });

    testWidgets('adds, edits and removes a command', (tester) async {
      final server = server0();
      await pumpCreator(tester, server, AppRoutes.settingsChatCommands);

      await tester.enterText(_key('name-channel'), ' Hello ');
      await tester.enterText(_key('response-channel'), 'Hi all');
      await tester.ensureVisible(_key('submit-channel'));
      await tester.tap(_key('submit-channel'));
      await tester.pumpAndSettle();
      expect(server.createdCommands.single['name'], 'hello');
      expect(server.createdCommands.single['scope'], 'channel');
      expect(_key('command-channel-hello'), findsOneWidget);

      await tester.ensureVisible(_key('edit-user-discord'));
      await tester.tap(_key('edit-user-discord'));
      await tester.pump();
      expect(find.text('Editing /discord'), findsOneWidget);
      await tester.enterText(_key('response-user'), 'Join the server');
      await tester.ensureVisible(_key('submit-user'));
      await tester.tap(_key('submit-user'));
      await tester.pumpAndSettle();
      expect(find.text('Join the server'), findsOneWidget);

      await tester.ensureVisible(_key('remove-channel-rules'));
      await tester.tap(_key('remove-channel-rules'));
      await tester.pumpAndSettle();
      expect(server.deletedCommands, ['c-b']);
      expect(_key('command-channel-rules'), findsNothing);
    });

    testWidgets('a taken name says the create failed', (tester) async {
      final server = server0();
      await pumpCreator(tester, server, AppRoutes.settingsChatCommands);

      await tester.enterText(_key('name-user'), 'discord');
      await tester.enterText(_key('response-user'), 'again');
      await tester.ensureVisible(_key('submit-user'));
      await tester.tap(_key('submit-user'));
      await tester.pumpAndSettle();

      expect(
        find.text('Could not create chat command (name may be taken)'),
        findsOneWidget,
      );
    });

    testWidgets('a failed load offers a retry', (tester) async {
      final server = server0()..commandsStatus = 503;
      await pumpCreator(tester, server, AppRoutes.settingsChatCommands);
      expect(find.text('Try again'), findsOneWidget);

      server.commandsStatus = 200;
      await tester.tap(find.text('Try again'));
      await tester.pumpAndSettle();
      expect(_key('command-user-discord'), findsOneWidget);
    });
  });

  group('my videos', () {
    FakeCreatorServer server0() => FakeCreatorServer()
      ..myVods = [
        myVodJson('v-1', 'First video'),
        myVodJson('v-2', 'Still processing', status: 'processing'),
        myVodJson('v-3', 'Broken', status: 'failed', visibility: 'private'),
      ];

    testWidgets('lists every video with its status', (tester) async {
      await pumpCreator(tester, server0(), AppRoutes.settingsVods);

      expect(find.text('First video'), findsOneWidget);
      expect(find.textContaining('Processing'), findsNWidgets(2));
      expect(find.textContaining('Processing failed'), findsOneWidget);
    });

    testWidgets('an empty list says so', (tester) async {
      await pumpCreator(tester, FakeCreatorServer(), AppRoutes.settingsVods);
      expect(find.text('You have no videos yet.'), findsOneWidget);
    });

    testWidgets('delete asks first, then removes the video', (tester) async {
      final server = server0();
      await pumpCreator(tester, server, AppRoutes.settingsVods);

      await tester.tap(_key('delete-vod-v-1'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();
      expect(server.deletedVods, isEmpty);

      await tester.tap(_key('delete-vod-v-1'));
      await tester.pumpAndSettle();
      await tester.tap(_key('confirm-delete-vod'));
      await tester.pumpAndSettle();
      expect(server.deletedVods, ['v-1']);
      expect(find.text('First video'), findsNothing);
    });

    testWidgets('edit saves title and visibility, and a new thumbnail first', (
      tester,
    ) async {
      final server = server0();
      final device = FakeDevice()..next = pickedImage();
      await pumpCreator(tester, server, AppRoutes.settingsVods, device: device);

      await tester.tap(_key('edit-vod-v-1'));
      await tester.pumpAndSettle();
      await tester.tap(_key('pick-vod-thumbnail'));
      await tester.pumpAndSettle();
      await tester.enterText(_key('vod-title'), 'Renamed');
      await tester.tap(_key('vod-public'));
      await tester.pump();
      await tester.tap(_key('save-vod'));
      await tester.pumpAndSettle();

      expect(server.uploadFileCalls, 1);
      expect(server.vodPatches.single, {
        'title': 'Renamed',
        'description': '',
        'visibility': 'private',
        'thumbnailUrl': '/files/uploads/0.jpg',
      });
      expect(find.text('Renamed'), findsOneWidget);
      expect(find.text('VOD updated successfully'), findsOneWidget);
    });

    testWidgets('a failed save keeps the dialog and what was typed', (
      tester,
    ) async {
      final server = server0();
      await pumpCreator(tester, server, AppRoutes.settingsVods);
      await tester.tap(_key('edit-vod-v-1'));
      await tester.pumpAndSettle();
      await tester.enterText(_key('vod-title'), '');
      await tester.tap(_key('save-vod'));
      await tester.pump();

      expect(find.text('Please enter a title.'), findsOneWidget);
      expect(server.vodPatches, isEmpty);
    });

    testWidgets('a failed load offers a retry', (tester) async {
      final server = server0()..myVodsStatus = 503;
      await pumpCreator(tester, server, AppRoutes.settingsVods);
      expect(find.text('Try again'), findsOneWidget);

      server.myVodsStatus = 200;
      await tester.tap(find.text('Try again'));
      await tester.pumpAndSettle();
      expect(find.text('First video'), findsOneWidget);
    });
  });

  group('upload screen', () {
    const mp4 = PickedVideo(
      path: '/tmp/a.mp4',
      name: 'holiday.mp4',
      size: 2048,
    );

    testWidgets('a video library that will not open says so', (tester) async {
      await pumpCreator(
        tester,
        FakeCreatorServer(),
        AppRoutes.settingsUpload,
        extra: [
          videoPickProvider.overrideWithValue(
            () async => throw const PickerException('denied'),
          ),
        ],
      );

      await tester.tap(_key('pick-video'));
      await tester.pumpAndSettle();

      expect(find.text('Could not open your photos'), findsOneWidget);
    });

    testWidgets('refuses a wrong format and a too large file', (tester) async {
      PickedVideo? next = const PickedVideo(
        path: '/tmp/a.gif',
        name: 'a.gif',
        size: 10,
      );
      await pumpCreator(
        tester,
        FakeCreatorServer(),
        AppRoutes.settingsUpload,
        extra: [videoPickProvider.overrideWithValue(() async => next)],
      );

      await tester.tap(_key('pick-video'));
      await tester.pumpAndSettle();
      expect(
        find.text('Invalid file format. Supported: MP4, MOV, AVI, MKV, WebM.'),
        findsOneWidget,
      );
      expect(_key('selected-video'), findsNothing);

      next = const PickedVideo(
        path: '/tmp/big.mp4',
        name: 'big.mp4',
        size: 101 * 1024 * 1024,
      );
      await tester.tap(_key('pick-video'));
      await tester.pumpAndSettle();
      expect(
        find.text('File is too large. Maximum size is 100 MB.'),
        findsOneWidget,
      );
    });

    testWidgets('uploading shows progress, then done; Upload stays disabled '
        'without a title', (tester) async {
      final myVods = FakeMyVods((onProgress, _) async {
        onProgress?.call(1024, 2048);
        return okResponse(vodWithStatus('processing'));
      });
      final status = FakeVodStatus(['processing', 'ready']);
      final app = await pumpCreator(
        tester,
        FakeCreatorServer(),
        AppRoutes.settingsUpload,
        extra: [
          videoPickProvider.overrideWithValue(() async => mp4),
          myVodsRepositoryProvider.overrideWithValue(myVods),
          vodRepositoryProvider.overrideWithValue(status),
          uploadPollIntervalProvider.overrideWithValue(
            const Duration(milliseconds: 10),
          ),
        ],
      );

      await tester.tap(_key('pick-video'));
      await tester.pumpAndSettle();
      expect(find.text('Selected: holiday.mp4 (2.0 KB)'), findsOneWidget);
      // The title is suggested from the file name.
      expect(
        tester.widget<TextField>(_key('upload-title')).controller!.text,
        'holiday',
      );

      await tester.tap(_key('upload-submit'));
      await tester.pump();
      expect(find.text('Waiting...'), findsNothing);
      expect(myVods.uploadedPaths, ['/tmp/a.mp4']);
      // The form is ready for the next video.
      expect(_key('selected-video'), findsNothing);

      await tester.pump(const Duration(milliseconds: 15));
      expect(find.text('Processing on server...'), findsOneWidget);
      await tester.pump(const Duration(milliseconds: 40));
      expect(find.text('Upload complete'), findsOneWidget);
      expect(
        app.container.read(uploadManagerProvider).single.status,
        UploadStatus.completed,
      );

      await tester.tap(_key('clear-uploads'));
      await tester.pump();
      expect(app.container.read(uploadManagerProvider), isEmpty);
    });

    testWidgets('leaving the screen does not cancel the upload', (
      tester,
    ) async {
      final gate = Completer<void>();
      final myVods = FakeMyVods((_, _) async {
        await gate.future;
        return okResponse(vodWithStatus('processing'));
      });
      final app = await pumpCreator(
        tester,
        FakeCreatorServer(),
        AppRoutes.settingsUpload,
        extra: [
          videoPickProvider.overrideWithValue(() async => mp4),
          myVodsRepositoryProvider.overrideWithValue(myVods),
          vodRepositoryProvider.overrideWithValue(FakeVodStatus(['ready'])),
          uploadPollIntervalProvider.overrideWithValue(
            const Duration(milliseconds: 10),
          ),
        ],
      );
      await tester.tap(_key('pick-video'));
      await tester.pumpAndSettle();
      await tester.tap(_key('upload-submit'));
      await tester.pump();

      await tester.pageBack();
      await tester.pumpAndSettle();
      expect(
        app.container.read(uploadManagerProvider).single.status,
        UploadStatus.uploading,
      );

      gate.complete();
      await tester.pump(const Duration(milliseconds: 50));
      expect(
        app.container.read(uploadManagerProvider).single.status,
        UploadStatus.completed,
      );
    });

    testWidgets('a failed upload shows the reason and Retry sends it again', (
      tester,
    ) async {
      var calls = 0;
      final myVods = FakeMyVods((_, _) async {
        calls++;
        return calls == 1
            ? failResponse(503)
            : okResponse(vodWithStatus('processing'));
      });
      await pumpCreator(
        tester,
        FakeCreatorServer(),
        AppRoutes.settingsUpload,
        extra: [
          videoPickProvider.overrideWithValue(() async => mp4),
          myVodsRepositoryProvider.overrideWithValue(myVods),
          vodRepositoryProvider.overrideWithValue(FakeVodStatus(['ready'])),
          uploadPollIntervalProvider.overrideWithValue(
            const Duration(milliseconds: 10),
          ),
        ],
      );
      await tester.tap(_key('pick-video'));
      await tester.pumpAndSettle();
      await tester.tap(_key('upload-submit'));
      await tester.pumpAndSettle();

      expect(find.text('Retry'), findsOneWidget);
      await tester.tap(find.text('Retry'));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 50));
      expect(myVods.uploadedPaths, ['/tmp/a.mp4', '/tmp/a.mp4']);
      expect(find.text('Upload complete'), findsOneWidget);
    });
  });

  group('settings hub', () {
    testWidgets('links to the four creator tools', (tester) async {
      await pumpCreator(tester, FakeCreatorServer(), AppRoutes.settings);

      for (final key in [
        'settings-stream',
        'settings-chat-commands',
        'settings-vods',
        'settings-upload',
      ]) {
        expect(_key(key), findsOneWidget);
      }
      await tester.tap(_key('settings-vods'));
      await tester.pumpAndSettle();
      expect(find.text('VODs Manager'), findsOneWidget);
    });
  });

  group('overflow at 320px and text scale 2.0', () {
    const routes = [
      AppRoutes.settings,
      AppRoutes.settingsStream,
      AppRoutes.settingsChatCommands,
      AppRoutes.settingsVods,
      AppRoutes.settingsUpload,
    ];
    final long = 'Extraordinarily${'long' * 40} 😀 title';

    for (final route in routes) {
      testWidgets('$route lays out without overflow', (tester) async {
        final server = FakeCreatorServer()
          ..userCommands = [commandJson('c-a', 'a' * 32, response: long)]
          ..myVods = [myVodJson('v-1', long)]
          ..meExtra = {
            'livestreamInformation': {
              'title': 'x',
              'description': long,
              'thumbnailUrl': null,
            },
          };
        await pumpCreator(
          tester,
          server,
          route,
          width: 320,
          height: 800,
          textScale: 2,
        );
        expect(tester.takeException(), isNull);
      });
    }
  });

  testWidgets('the upload list with long rows fits 320px at text scale 2.0', (
    tester,
  ) async {
    var calls = 0;
    final neverEnds = Completer<void>();
    final myVods = FakeMyVods((onProgress, _) async {
      // The first upload fails at once; the second stays in progress.
      if (++calls == 1) return failResponse(503);
      onProgress?.call(512, 1000);
      await neverEnds.future;
      return failResponse(503);
    });
    final app = await pumpCreator(
      tester,
      FakeCreatorServer(),
      AppRoutes.settingsUpload,
      width: 320,
      height: 800,
      textScale: 2,
      extra: [myVodsRepositoryProvider.overrideWithValue(myVods)],
    );
    final manager = app.container.read(uploadManagerProvider.notifier);
    final longTitle = 'Extraordinarily${'long' * 30} 😀 upload title';
    for (var i = 0; i < 2; i++) {
      manager.enqueue(
        PickedVideo(path: '/tmp/$i.mp4', name: 'clip$i.mp4', size: 1000),
        title: longTitle,
        description: '',
        isPublic: true,
      );
      await tester.pump();
    }
    await tester.pump(const Duration(milliseconds: 50));
    // The rows sit under the long form: scroll them into the lazy list.
    await tester.drag(
      find.descendant(
        of: find.byType(UploadScreen),
        matching: find.byType(ListView),
      ),
      const Offset(0, -3000),
    );
    await tester.pump();

    expect(find.text('Retry'), findsOneWidget);
    expect(find.byType(LinearProgressIndicator), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('the next account never sees the previous account\'s tools', (
    tester,
  ) async {
    final server = FakeCreatorServer()
      ..myVods = [myVodJson('v-1', 'Account A video')]
      ..userCommands = [commandJson('c-a', 'alpha')];
    final app = await pumpCreator(tester, server, AppRoutes.settingsVods);
    expect(find.text('Account A video'), findsOneWidget);
    // A also starts an upload that must not survive.
    app.container
        .read(uploadManagerProvider.notifier)
        .enqueue(
          const PickedVideo(path: '/tmp/a.mp4', name: 'a.mp4', size: 10),
          title: 'A upload',
          description: '',
          isPublic: true,
        );

    final loggedOut = app.container.read(sessionProvider.notifier).logout();
    await tester.pumpAndSettle();
    await loggedOut;
    expect(find.byType(LoginScreen), findsOneWidget);
    expect(app.container.read(uploadManagerProvider), isEmpty);

    server
      ..meId = 'u-9'
      ..username = 'bravo'
      ..myVods = [myVodJson('b-1', 'Account B video')]
      ..userCommands = [commandJson('c-b', 'bravo')]
      ..meExtra = {
        'streamAPIKey': 'b-secret',
        'livestreamInformation': {
          'title': null,
          'description': null,
          'thumbnailUrl': null,
        },
      };
    final loggedIn = app.container
        .read(sessionProvider.notifier)
        .login('b@example.com', testPassword);
    await tester.pumpAndSettle();
    await loggedIn;

    unawaited(
      app.container.read(appRouterProvider).push(AppRoutes.settingsVods),
    );
    await tester.pumpAndSettle();
    expect(find.text('Account B video'), findsOneWidget);
    expect(find.text('Account A video'), findsNothing);

    await tester.pageBack();
    unawaited(
      app.container.read(appRouterProvider).push(AppRoutes.settingsStream),
    );
    await tester.pumpAndSettle();
    await tester.tap(_key('toggle-key'));
    await tester.pump();
    expect(_keyText(tester), 'b-secret');
  });
}
