import 'dart:async';

import 'package:cookie_jar/cookie_jar.dart';
import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:flutter_riverpod/misc.dart' show Override;
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/core/router/app_router.dart';
import 'package:letslive_mobile/core/realtime/realtime_manager.dart';
import 'package:letslive_mobile/features/watch/chat/presentation/chat_panel.dart';

import '../../../core/realtime/fake_channel.dart';
import '../channel_screen_test.dart' show liveServer;
import '../fake_player.dart';
import '../watch_test_app.dart';
import 'chat_fakes.dart';

const _roomTopic = 'room:u-1';

/// A realtime manager over [FakeChannel]s the test drives, installed into
/// the app through [overrides].
class _Realtime {
  _Realtime() {
    manager = RealtimeManager(
      url: 'ws://localhost:8000/realtime',
      cookieJar: CookieJar(),
      refreshSession: () async => true,
      connect: (uri, headers) {
        final channel = FakeChannel(uri, headers);
        channels.add(channel);
        return channel;
      },
    );
  }

  late final RealtimeManager manager;
  final List<FakeChannel> channels = [];

  List<Override> get overrides => [
    realtimeManagerProvider.overrideWithValue(manager),
  ];

  FakeChannel get channel => channels.last;

  /// Stops the ping timer a connected socket leaves behind (a test fails
  /// on pending timers before its tear-downs run).
  Future<void> close() => manager.dispose();

  /// Opens the socket; the chat then receives pushes.
  Future<void> connect(WidgetTester tester) async {
    manager.start();
    await tester.pump();
    channel.open();
    await tester.pump();
  }

  Future<void> pushMessage(
    WidgetTester tester,
    Map<String, Object?> data, {
    String topic = _roomTopic,
  }) async {
    channel.serverSends({
      'op': 'event',
      'topic': topic,
      'type': 'chat.message',
      'data': data,
    });
    await tester.pump();
  }
}

Map<String, Object?> _command(String id, String name, String response) => {
  'id': id,
  'scope': 'channel',
  'ownerId': 'u-1',
  'name': name,
  'response': response,
  'description': 'Link to $name',
  'createdAt': '2026-11-04T00:00:00Z',
};

Finder _inPanel(Finder finder) =>
    find.descendant(of: find.byType(ChatPanel), matching: finder);

final _input = find.byType(TextField);

void main() {
  testWidgets('shows a spinner while the history loads', (tester) async {
    final server = liveServer()..holdChatHistory = Completer<void>();
    await pumpChannel(
      tester,
      server,
      FakePlayers(),
      'u-1',
      settle: false,
      overrides: _Realtime().overrides,
    );
    await tester.pump(const Duration(milliseconds: 100));
    await tester.pump(const Duration(milliseconds: 100));
    expect(_inPanel(find.byType(CircularProgressIndicator)), findsOneWidget);

    server.holdChatHistory!.complete();
    await tester.pumpAndSettle();
    expect(_inPanel(find.byType(CircularProgressIndicator)), findsNothing);
  });

  testWidgets('empty history shows only the message box', (tester) async {
    await pumpChannel(
      tester,
      liveServer(),
      FakePlayers(),
      'u-1',
      overrides: _Realtime().overrides,
    );

    expect(find.text('Chat'), findsOneWidget);
    expect(find.text('Type a message...'), findsOneWidget);
    expect(_inPanel(find.byType(CircularProgressIndicator)), findsNothing);
  });

  testWidgets('history failure shows an error that retries', (tester) async {
    final server = liveServer()..chatHistoryStatus = 503;
    await pumpChannel(
      tester,
      server,
      FakePlayers(),
      'u-1',
      overrides: _Realtime().overrides,
    );
    expect(_inPanel(find.text('Try again')), findsOneWidget);
    // The message box still works.
    expect(find.text('Type a message...'), findsOneWidget);

    server
      ..chatHistoryStatus = 200
      ..chatHistory = [historyJson('a', 'back again', username: 'alice')];
    await tester.tap(_inPanel(find.text('Try again')));
    await tester.pumpAndSettle();

    expect(
      find.textContaining('back again', findRichText: true),
      findsOneWidget,
    );
  });

  testWidgets('history shows senders, text and emotes', (tester) async {
    final server = liveServer()
      ..chatHistory = [
        historyJson('a', 'hello :fire: world', username: 'alice'),
        historyJson('b', 'unknown :nope:', username: 'bob'),
      ];
    await pumpChannel(
      tester,
      server,
      FakePlayers(),
      'u-1',
      overrides: _Realtime().overrides,
    );

    expect(
      find.textContaining('alice: hello 🔥 world', findRichText: true),
      findsOneWidget,
    );
    expect(
      find.textContaining('bob: unknown :nope:', findRichText: true),
      findsOneWidget,
    );
  });

  testWidgets('offline channel still shows the chat', (tester) async {
    final server = liveServer()
      ..liveByUser = {}
      ..chatHistory = [historyJson('a', 'still here')];
    await pumpChannel(
      tester,
      server,
      FakePlayers(),
      'u-1',
      overrides: _Realtime().overrides,
    );

    expect(
      find.textContaining('still here', findRichText: true),
      findsOneWidget,
    );
    expect(find.text('Type a message...'), findsOneWidget);
  });

  testWidgets('a live message appears once, even if pushed again', (
    tester,
  ) async {
    final realtime = _Realtime();
    final server = liveServer()..chatHistory = [historyJson('a', 'first')];
    await pumpChannel(
      tester,
      server,
      FakePlayers(),
      'u-1',
      overrides: realtime.overrides,
    );
    await realtime.connect(tester);
    expect(realtime.channel.opsFor('subscribe'), [_roomTopic]);

    await realtime.pushMessage(tester, pushJson('b', 'live one'));
    await realtime.pushMessage(tester, pushJson('b', 'live one'));
    await realtime.pushMessage(
      tester,
      pushJson('x', 'other room'),
      topic: 'room:someone-else',
    );

    expect(find.textContaining('live one', findRichText: true), findsOneWidget);
    expect(find.textContaining('first', findRichText: true), findsOneWidget);
    expect(find.textContaining('other room', findRichText: true), findsNothing);
    await realtime.close();
  });

  testWidgets('leaving the channel unsubscribes from the room', (tester) async {
    final realtime = _Realtime();
    final app = await pumpChannel(
      tester,
      liveServer(),
      FakePlayers(),
      'u-1',
      overrides: realtime.overrides,
    );
    await realtime.connect(tester);
    expect(realtime.channel.opsFor('subscribe'), [_roomTopic]);

    app.container.read(appRouterProvider).pop();
    await tester.pumpAndSettle();

    expect(realtime.channel.opsFor('unsubscribe'), [_roomTopic]);
    await realtime.close();
  });

  testWidgets('emoji, Vietnamese and long strings do not overflow at 320px', (
    tester,
  ) async {
    final server = liveServer()
      ..chatHistory = [
        historyJson(
          'a',
          'Chào các bạn, hôm nay stream rất vui 🎉🎉🎉',
          username: 'Nguyễn Văn Á',
        ),
        historyJson('b', 'x' * 400, username: 'u' * 60),
        historyJson('c', '😀' * 150),
      ];
    final realtime = _Realtime();
    await pumpChannel(
      tester,
      server,
      FakePlayers(),
      'u-1',
      width: 320,
      height: 568,
      textScale: 2,
      overrides: realtime.overrides,
    );
    await realtime.connect(tester);
    await realtime.pushMessage(
      tester,
      pushJson('d', 'Tiếng Việt không dấu va co dau ${'ệ' * 200}'),
    );
    await tester.enterText(_input, '/');
    await tester.pump();

    expect(tester.takeException(), isNull);
    expect(
      find.textContaining('Tiếng Việt không dấu', findRichText: true),
      findsOneWidget,
    );
    await realtime.close();
  });

  testWidgets('send posts the text and clears the box', (tester) async {
    final server = liveServer();
    await pumpChannel(
      tester,
      server,
      FakePlayers(),
      'u-1',
      overrides: _Realtime().overrides,
    );

    await tester.enterText(_input, '  hi :smile:  ');
    await tester.tap(find.byTooltip('Send message'));
    await tester.pumpAndSettle();

    expect(server.sentMessages.single, {'roomId': 'u-1', 'text': 'hi :smile:'});
    expect(tester.widget<TextField>(_input).controller!.text, isEmpty);
    expect(find.textContaining('hi 😊', findRichText: true), findsOneWidget);
  });

  testWidgets('send failure keeps text and offers retry', (tester) async {
    final server = liveServer()..sendStatus = 503;
    await pumpChannel(
      tester,
      server,
      FakePlayers(),
      'u-1',
      overrides: _Realtime().overrides,
    );

    await tester.enterText(_input, 'do not lose me');
    await tester.tap(find.byTooltip('Send message'));
    await tester.pumpAndSettle();

    expect(tester.widget<TextField>(_input).controller!.text, 'do not lose me');
    expect(
      find.text('Message not sent. Your text is kept. Something went wrong.'),
      findsOneWidget,
    );
    expect(
      find.textContaining('do not lose me', findRichText: true),
      findsOneWidget,
    );

    server.sendStatus = 200;
    await tester.tap(find.text('Retry'));
    await tester.pumpAndSettle();

    expect(server.sentMessages.map((m) => m['text']), [
      'do not lose me',
      'do not lose me',
    ]);
    expect(tester.widget<TextField>(_input).controller!.text, isEmpty);
    expect(find.text('Retry'), findsNothing);
    // Shown once: the send response; no push in this test.
    expect(
      find.textContaining('do not lose me', findRichText: true),
      findsOneWidget,
    );
  });

  testWidgets('command suggestions appear for "/" and filter as you type', (
    tester,
  ) async {
    final server = liveServer()
      ..chatCommands = [_command('c1', 'discord', 'Join discord.gg/x')];
    await pumpChannel(
      tester,
      server,
      FakePlayers(),
      'u-1',
      overrides: _Realtime().overrides,
    );

    await tester.enterText(_input, '/');
    await tester.pump();
    for (final usage in [
      '/shrug [text]',
      '/tableflip [text]',
      '/unflip [text]',
      '/roll [max]',
      '/help',
      '/discord',
    ]) {
      expect(find.text(usage), findsOneWidget, reason: usage);
    }

    await tester.enterText(_input, '/r');
    await tester.pump();
    expect(find.text('/roll [max]'), findsOneWidget);
    expect(find.text('/shrug [text]'), findsNothing);

    await tester.tap(find.text('/roll [max]'));
    await tester.pump();
    expect(tester.widget<TextField>(_input).controller!.text, '/roll ');
    expect(find.text('/roll [max]'), findsNothing);
  });

  testWidgets('a custom channel command sends its response', (tester) async {
    final server = liveServer()
      ..chatCommands = [_command('c1', 'discord', 'Join discord.gg/x')];
    await pumpChannel(
      tester,
      server,
      FakePlayers(),
      'u-1',
      overrides: _Realtime().overrides,
    );

    await tester.enterText(_input, '/discord');
    await tester.tap(find.byTooltip('Send message'));
    await tester.pumpAndSettle();

    expect(server.sentMessages.single['text'], 'Join discord.gg/x');
  });

  testWidgets('/shrug sends the expansion, /help and errors stay local', (
    tester,
  ) async {
    final server = liveServer();
    await pumpChannel(
      tester,
      server,
      FakePlayers(),
      'u-1',
      overrides: _Realtime().overrides,
    );

    await tester.enterText(_input, '/shrug oh well');
    await tester.tap(find.byTooltip('Send message'));
    await tester.pumpAndSettle();
    expect(server.sentMessages.single['text'], r'oh well ¯\_(ツ)_/¯');

    await tester.enterText(_input, '/help');
    await tester.tap(find.byTooltip('Send message'));
    await tester.pumpAndSettle();
    expect(
      find.textContaining('Available chat commands:', findRichText: true),
      findsOneWidget,
    );

    await tester.enterText(_input, '/nope');
    await tester.tap(find.byTooltip('Send message'));
    await tester.pumpAndSettle();
    expect(find.text('Unknown chat command: /nope'), findsOneWidget);
    expect(server.sentMessages, hasLength(1));
  });

  testWidgets('emote picker inserts the emote at the cursor', (tester) async {
    await pumpChannel(
      tester,
      liveServer(),
      FakePlayers(),
      'u-1',
      overrides: _Realtime().overrides,
    );

    await tester.enterText(_input, 'helloworld');
    tester.widget<TextField>(_input).controller!.selection =
        const TextSelection.collapsed(offset: 5);
    await tester.tap(find.byTooltip('Emotes'));
    await tester.pumpAndSettle();
    expect(find.text('Smileys'), findsOneWidget);

    await tester.tap(find.text('😊'));
    await tester.pumpAndSettle();

    final controller = tester.widget<TextField>(_input).controller!;
    expect(controller.text, 'hello:smile:world');
    expect(controller.selection.baseOffset, 12);
  });

  testWidgets('emote picker search filters and shows an empty state', (
    tester,
  ) async {
    await pumpChannel(
      tester,
      liveServer(),
      FakePlayers(),
      'u-1',
      overrides: _Realtime().overrides,
    );
    await tester.tap(find.byTooltip('Emotes'));
    await tester.pumpAndSettle();

    await tester.enterText(find.widgetWithText(TextField, '').last, 'rocket');
    await tester.pump();
    expect(find.text('🚀'), findsOneWidget);
    expect(find.text('😊'), findsNothing);

    await tester.enterText(find.widgetWithText(TextField, 'rocket'), 'zzzz');
    await tester.pump();
    expect(find.text('No emotes found'), findsOneWidget);
  });

  group('keyboard open on a 320x568 phone at text scale 2', () {
    final longHistory = [
      historyJson('a', 'Chào các bạn 🎉', username: 'Nguyễn Văn Á'),
      historyJson('b', 'x' * 300),
    ];

    Future<void> openKeyboard(WidgetTester tester) async {
      tester.view.viewInsets = const FakeViewPadding(bottom: 260);
      addTearDown(tester.view.resetViewInsets);
      await tester.pump();
    }

    testWidgets('live layout fits with command suggestions open', (
      tester,
    ) async {
      final server = liveServer()
        ..chatHistory = longHistory
        ..chatCommands = [_command('c1', 'discord', 'Join us')];
      await pumpChannel(
        tester,
        server,
        FakePlayers(),
        'u-1',
        width: 320,
        height: 568,
        textScale: 2,
        overrides: _Realtime().overrides,
      );
      expect(tester.takeException(), isNull, reason: 'keyboard closed');

      await openKeyboard(tester);
      await tester.enterText(_input, 'half typed');
      await tester.enterText(_input, '/');
      await tester.pump();

      expect(tester.takeException(), isNull, reason: 'keyboard open');
      // Stream info and the chat heading make room for the lines.
      expect(find.text('Live title'), findsNothing);
      expect(find.text('Chat'), findsNothing);
      expect(find.text('/shrug [text]'), findsOneWidget);
      // The typed text survived the layout change.
      await tester.enterText(_input, 'still here');
      expect(tester.widget<TextField>(_input).controller!.text, 'still here');
    });

    testWidgets('offline layout fits with command suggestions open', (
      tester,
    ) async {
      final server = liveServer()
        ..liveByUser = {}
        ..chatHistory = longHistory
        ..chatCommands = [_command('c1', 'discord', 'Join us')];
      await pumpChannel(
        tester,
        server,
        FakePlayers(),
        'u-1',
        width: 320,
        height: 568,
        textScale: 2,
        overrides: _Realtime().overrides,
      );
      expect(tester.takeException(), isNull, reason: 'keyboard closed');

      await tester.enterText(_input, 'kept while the keyboard opens');
      await openKeyboard(tester);
      await tester.enterText(_input, '/');
      await tester.pump();

      expect(tester.takeException(), isNull, reason: 'keyboard open');
      expect(find.text('/shrug [text]'), findsOneWidget);
    });
  });

  testWidgets('a pasted run of 300 emoji is cut to the server limit', (
    tester,
  ) async {
    await pumpChannel(
      tester,
      liveServer(),
      FakePlayers(),
      'u-1',
      overrides: _Realtime().overrides,
    );

    await tester.enterText(_input, '🎉' * 300);
    await tester.pump();

    final text = tester.widget<TextField>(_input).controller!.text;
    // 500 UTF-16 units, as the server counts, not 500 characters.
    expect(text.length, 500);
    expect(text.runes.length, 250);
  });

  testWidgets('text typed during a send is kept, not wiped', (tester) async {
    final server = liveServer()..holdSend = Completer<void>();
    await pumpChannel(
      tester,
      server,
      FakePlayers(),
      'u-1',
      overrides: _Realtime().overrides,
    );

    await tester.enterText(_input, 'one');
    await tester.tap(find.byTooltip('Send message'));
    await tester.pump();
    await tester.enterText(_input, 'two');
    // The keyboard action during a send does nothing; the text stays.
    await tester.testTextInput.receiveAction(TextInputAction.send);
    await tester.pump();
    expect(server.sentMessages.map((m) => m['text']), ['one']);

    server.holdSend!.complete();
    await tester.pumpAndSettle();
    expect(tester.widget<TextField>(_input).controller!.text, 'two');

    await tester.tap(find.byTooltip('Send message'));
    await tester.pumpAndSettle();
    expect(server.sentMessages.map((m) => m['text']), ['one', 'two']);
    expect(tester.widget<TextField>(_input).controller!.text, isEmpty);
  });

  testWidgets('retry resends the exact failed text, not a new roll', (
    tester,
  ) async {
    final server = liveServer()..sendStatus = 503;
    await pumpChannel(
      tester,
      server,
      FakePlayers(),
      'u-1',
      overrides: _Realtime().overrides,
    );

    await tester.enterText(_input, '/roll 1000000');
    await tester.tap(find.byTooltip('Send message'));
    await tester.pumpAndSettle();
    server.sendStatus = 200;
    await tester.tap(find.text('Retry'));
    await tester.pumpAndSettle();

    final texts = server.sentMessages.map((m) => m['text']).toList();
    expect(texts, hasLength(2));
    expect(texts[1], texts[0]);
  });

  testWidgets('typing a shortcode suggests emotes; tapping inserts one', (
    tester,
  ) async {
    await pumpChannel(
      tester,
      liveServer(),
      FakePlayers(),
      'u-1',
      overrides: _Realtime().overrides,
    );

    await tester.enterText(_input, 'hi :fi');
    await tester.pump();
    expect(find.text('🔥'), findsOneWidget);
    expect(find.text('😊'), findsNothing);

    await tester.tap(find.text('🔥'));
    await tester.pump();

    expect(tester.widget<TextField>(_input).controller!.text, 'hi :fire: ');
    expect(find.text('🔥'), findsNothing);
  });

  testWidgets('long press on a message shows when it was sent', (tester) async {
    final server = liveServer()
      ..chatHistory = [historyJson('a', 'timed', at: '2026-11-04T10:30:00Z')];
    await pumpChannel(
      tester,
      server,
      FakePlayers(),
      'u-1',
      overrides: _Realtime().overrides,
    );
    final expected = DateFormat.yMMMd('en')
        .add_jm()
        .format(DateTime.utc(2026, 11, 4, 10, 30).toLocal());

    await tester.longPress(find.textContaining('timed', findRichText: true));
    await tester.pumpAndSettle();

    expect(find.text(expected), findsOneWidget);
  });
}
