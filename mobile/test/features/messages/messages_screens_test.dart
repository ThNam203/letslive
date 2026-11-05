import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/features/auth/auth_providers.dart';
import 'package:letslive_mobile/features/auth/presentation/login_screen.dart';
import 'package:letslive_mobile/features/messages/presentation/conversation_screen.dart';

import '../auth/auth_test_app.dart';
import '../home/home_fixtures.dart';
import 'message_fakes.dart';
import 'messages_test_app.dart';

Map<String, Object?> _lastMessage(
  String text, {
  String at = '2026-11-04T10:00:00.000Z',
}) => {
  '_id': 'lm',
  'senderId': 'u-2',
  'senderUsername': 'bob',
  'text': text,
  'createdAt': at,
};

FakeMessagesServer _server() => FakeMessagesServer()
  ..conversations.add(
    conversationJson('c-1', lastMessage: _lastMessage('see you soon')),
  )
  ..threads['c-1'] = threadJson('c-1', 3);

/// The badge on the Messages tab icon.
Finder _tabBadge() => find.ancestor(
  of: find.descendant(
    of: find.byType(NavigationBar),
    matching: find.byIcon(Icons.chat_bubble_outline),
  ),
  matching: find.byType(Badge),
);

void main() {
  testWidgets('the list shows the person, last message and unread count', (
    tester,
  ) async {
    final server = _server()..unread['c-1'] = 3;
    await pumpMessages(tester, server);

    expect(find.text('bob'), findsOneWidget);
    expect(find.text('see you soon'), findsOneWidget);
    expect(
      find.descendant(
        of: find.byKey(const ValueKey('conversation-c-1')),
        matching: find.text('3'),
      ),
      findsOneWidget,
    );
    expect(tester.widget<Badge>(_tabBadge()).isLabelVisible, isTrue);
  });

  testWidgets('an empty list says so; a failed one offers a retry', (
    tester,
  ) async {
    final server = FakeMessagesServer()..listStatus = 503;
    await pumpMessages(tester, server);
    expect(find.text('Try again'), findsOneWidget);

    server
      ..listStatus = 200
      ..conversations.clear();
    await tester.tap(find.text('Try again'));
    await tester.pumpAndSettle();

    expect(find.text('No conversations yet'), findsOneWidget);
  });

  testWidgets('opening a conversation marks it read and drops the badge', (
    tester,
  ) async {
    final server = _server()..unread['c-1'] = 2;
    final app = await pumpMessages(tester, server);
    expect(find.text('2'), findsWidgets);

    await tester.tap(find.byKey(const ValueKey('conversation-c-1')));
    await tester.pumpAndSettle();

    expect(find.byType(ConversationScreen), findsOneWidget);
    expect(server.readCalls, ['c-1']);
    expect(find.text('text 0'), findsOneWidget);
    expect(find.text('text 2'), findsOneWidget);
    expect(find.text('Beginning of conversation'), findsOneWidget);

    await tester.pageBack();
    await tester.pumpAndSettle();
    expect(
      find.descendant(
        of: find.byKey(const ValueKey('conversation-c-1')),
        matching: find.text('2'),
      ),
      findsNothing,
    );
    expect(app.requestsTo('/conversations/unread-counts'), isNotEmpty);
  });

  testWidgets('sending adds the message and clears the box; a failure keeps '
      'the text', (tester) async {
    final server = _server();
    final app = await pumpMessages(tester, server);
    await openConversation(tester, app, 'c-1');

    server.sendStatus = 503;
    await tester.enterText(
      find.byKey(const ValueKey('message-input')),
      'hello',
    );
    await tester.tap(find.byKey(const ValueKey('send-message')));
    await tester.pumpAndSettle();
    expect(find.textContaining('Message not sent'), findsOneWidget);
    expect(find.text('hello'), findsOneWidget);

    server.sendStatus = 200;
    await tester.tap(find.text('Retry'));
    await tester.pumpAndSettle();

    expect(server.sentTexts, ['hello', 'hello']);
    expect(find.textContaining('Message not sent'), findsNothing);
    expect(
      tester
          .widget<TextField>(find.byKey(const ValueKey('message-input')))
          .controller!
          .text,
      isEmpty,
    );
    expect(find.text('hello'), findsOneWidget);
  });

  testWidgets('typing shows to the others after the first keystroke and '
      'stops after two seconds', (tester) async {
    final server = _server();
    final app = await pumpMessages(tester, server);
    await openConversation(tester, app, 'c-1');

    await tester.enterText(find.byKey(const ValueKey('message-input')), 'h');
    await tester.pump();
    expect(server.typingPosts, [('c-1', 'start')]);
    await tester.pump(const Duration(seconds: 2));
    await tester.pump();
    expect(server.typingPosts, [('c-1', 'start'), ('c-1', 'stop')]);
  });

  testWidgets('only own messages offer edit and delete', (tester) async {
    final server = _server();
    server.threads['c-1']!.add(
      messageJson(
        'mine',
        'my words',
        senderId: me,
        senderUsername: 'abcdef',
        at: '2026-11-04T11:00:00.000Z',
      ),
    );
    final app = await pumpMessages(tester, server);
    await openConversation(tester, app, 'c-1');

    await tester.longPress(find.text('text 1'));
    await tester.pumpAndSettle();
    expect(find.text('Edit message'), findsNothing);

    await tester.longPress(find.text('my words'));
    await tester.pumpAndSettle();
    expect(find.text('Edit message'), findsOneWidget);
    expect(find.text('Delete message'), findsOneWidget);

    await tester.tap(find.text('Edit message'));
    await tester.pumpAndSettle();
    await tester.enterText(
      find.byKey(const ValueKey('edit-message-input')),
      'better words',
    );
    await tester.tap(find.byKey(const ValueKey('save-message')));
    await tester.pumpAndSettle();
    expect(find.text('better words'), findsOneWidget);
    expect(find.text('(edited)'), findsOneWidget);

    await tester.longPress(find.text('better words'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Delete message'));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const ValueKey('confirm-delete')));
    await tester.pumpAndSettle();
    expect(find.text('This message was deleted'), findsOneWidget);
    expect(find.text('better words'), findsNothing);
  });

  testWidgets('pushes show typing, the online dot, new messages and the '
      'badge', (tester) async {
    final server = _server();
    server.conversations.add(conversationJson('c-2', otherName: 'carol'));
    final socket = FakeSocket();
    addTearDown(socket.dispose);
    final app = await pumpMessages(tester, server, socket: socket);
    await socket.connect(tester);

    await socket.push(tester, 'dm:user_online', {'userId': 'u-2'});
    expect(find.byKey(const ValueKey('online-dot')), findsNWidgets(2));

    // A message in a conversation that is not open counts as unread.
    await socket.push(tester, 'dm:new_message', {
      'conversationId': 'c-2',
      'message': messageJson('x1', 'ping', conversationId: 'c-2'),
    });
    expect(
      find.descendant(
        of: find.byKey(const ValueKey('conversation-c-2')),
        matching: find.text('1'),
      ),
      findsOneWidget,
    );
    expect(find.text('ping'), findsOneWidget);

    await openConversation(tester, app, 'c-1');
    await socket.push(tester, 'dm:user_typing', {
      'conversationId': 'c-1',
      'userId': 'u-2',
      'username': 'bob',
    });
    expect(find.text('bob is typing...'), findsOneWidget);
    expect(find.text('Online'), findsOneWidget);

    await socket.push(tester, 'dm:new_message', {
      'conversationId': 'c-1',
      'message': messageJson('x2', 'live one', at: '2026-11-04T12:00:00.000Z'),
    });
    await tester.pumpAndSettle();
    expect(find.text('live one'), findsOneWidget);
    // Open conversation: read at once, not counted as unread.
    expect(server.readCalls, contains('c-1'));

    // Stops the socket's ping timer before the test ends.
    await socket.dispose();
  });

  testWidgets('an unknown conversation shows a not-found notice with a way '
      'back', (tester) async {
    final server = _server();
    final app = await pumpMessages(tester, server);
    await openConversation(tester, app, 'gone');

    expect(find.text('Conversation not found.'), findsOneWidget);
    await tester.tap(find.text('Back'));
    await tester.pumpAndSettle();
    expect(find.byType(ConversationScreen), findsNothing);
  });

  testWidgets('a new direct message: search, pick, create, open', (
    tester,
  ) async {
    final server = _server()
      ..searchUsers = [
        publicUserJson(id: 'u-1', username: 'abcdef'),
        publicUserJson(id: 'u-7', username: 'pickme'),
      ];
    await pumpMessages(tester, server);

    await tester.tap(find.byKey(const ValueKey('new-conversation')));
    await tester.pumpAndSettle();
    await tester.enterText(find.byKey(const ValueKey('user-search')), 'pi');
    await tester.pump(const Duration(milliseconds: 300));
    await tester.pumpAndSettle();
    // The user themself is not offered.
    expect(find.byKey(const ValueKey('result-u-1')), findsNothing);
    await tester.tap(find.byKey(const ValueKey('result-u-7')));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const ValueKey('create-conversation')));
    await tester.pumpAndSettle();

    expect(server.created.single, {
      'type': 'dm',
      'participantIds': ['u-7'],
    });
    expect(find.byType(ConversationScreen), findsOneWidget);
  });

  testWidgets('a group takes several people and a name', (tester) async {
    final server = _server()
      ..searchUsers = [
        publicUserJson(id: 'u-7', username: 'pickme'),
        publicUserJson(id: 'u-8', username: 'pickyou'),
      ];
    await pumpMessages(tester, server);
    await tester.tap(find.byKey(const ValueKey('new-conversation')));
    await tester.pumpAndSettle();

    await tester.tap(find.byKey(const ValueKey('mode-group')));
    await tester.pumpAndSettle();
    await tester.enterText(find.byKey(const ValueKey('group-name')), 'Crew');
    for (final id in ['u-7', 'u-8']) {
      await tester.enterText(find.byKey(const ValueKey('user-search')), 'pick');
      await tester.pump(const Duration(milliseconds: 300));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(ValueKey('result-$id')));
      await tester.pumpAndSettle();
    }
    expect(find.byKey(const ValueKey('selected-u-7')), findsOneWidget);
    expect(find.byKey(const ValueKey('selected-u-8')), findsOneWidget);
    await tester.tap(find.byKey(const ValueKey('create-conversation')));
    await tester.pumpAndSettle();

    expect(server.created.single, {
      'type': 'group',
      'participantIds': ['u-7', 'u-8'],
      'name': 'Crew',
    });
  });

  testWidgets('long names, Vietnamese and emoji do not overflow at 320px with '
      'large text', (tester) async {
    const long =
        'Xin chào các bạn rất dài dài dài 🎉🎉🎉 '
        'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa';
    final server = FakeMessagesServer()
      ..conversations.add(
        conversationJson(
          'c-1',
          otherName: 'Nguyễn Văn Một Cái Tên Rất Là Dài Dòng',
          lastMessage: _lastMessage(long),
        ),
      )
      ..threads['c-1'] = [
        messageJson('a', long),
        messageJson(
          'b',
          long,
          senderId: me,
          senderUsername: 'abcdef',
          at: '2026-11-04T10:01:00.000Z',
        ),
        messageJson(
          'c',
          'Sent an image',
          type: 'image',
          at: '2026-11-04T10:02:00.000Z',
        ),
      ]
      ..unread['c-1'] = 120;
    final app = await pumpMessages(
      tester,
      server,
      width: 320,
      height: 568,
      textScale: 2,
    );
    expect(tester.takeException(), isNull);

    await openConversation(tester, app, 'c-1');
    expect(tester.takeException(), isNull);
    expect(find.byKey(const ValueKey('message-input')), findsOneWidget);
  });

  testWidgets('the next account never sees the previous account\'s '
      'conversations', (tester) async {
    final server = _server()..unread['c-1'] = 4;
    final app = await pumpMessages(tester, server);
    expect(find.text('bob'), findsOneWidget);
    expect(app.requestsTo('/conversations'), hasLength(1));

    final loggedOut = app.container.read(sessionProvider.notifier).logout();
    await tester.pumpAndSettle();
    await loggedOut;
    expect(find.byType(LoginScreen), findsOneWidget);

    // Account B has one conversation with someone else and nothing unread.
    server
      ..conversations.clear()
      ..conversations.add(conversationJson('c-9', otherName: 'dave'))
      ..unread.clear();
    final loggedIn = app.container
        .read(sessionProvider.notifier)
        .login('b@example.com', testPassword);
    await tester.pumpAndSettle();
    await loggedIn;
    await tester.tap(
      find.descendant(
        of: find.byType(NavigationBar),
        matching: find.text('Messages'),
      ),
    );
    await tester.pumpAndSettle();

    expect(app.requestsTo('/conversations'), hasLength(2));
    expect(app.requestsTo('/conversations/unread-counts'), hasLength(2));
    expect(find.text('dave'), findsOneWidget);
    expect(find.text('bob'), findsNothing);
    expect(tester.widget<Badge>(_tabBadge()).isLabelVisible, isFalse);
  });
}
