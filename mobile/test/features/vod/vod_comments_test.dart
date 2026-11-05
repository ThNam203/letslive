import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';

import '../watch/fake_live_player.dart';
import 'vod_test_app.dart';

/// Finds the "Comment actions" menu buttons: one per own comment.
final _menus = find.byTooltip('Comment actions');

Finder _like(String id) => find.byKey(ValueKey('like-$id'));

Finder _heart(String id, IconData icon) =>
    find.descendant(of: _like(id), matching: find.byIcon(icon));

/// The comments under a video: states, paging, replies, likes, edit, delete.
void main() {
  group('states', () {
    testWidgets('empty list shows the invitation to comment', (tester) async {
      await pumpVod(tester, FakeVodServer(), FakeLivePlayers());

      expect(
        find.text('No comments yet. Be the first to comment!'),
        findsOneWidget,
      );
    });

    testWidgets('a failed load shows the error with a retry', (tester) async {
      final server = FakeVodServer()..commentsStatus = 503;
      await pumpVod(tester, server, FakeLivePlayers());

      expect(find.text('Try again'), findsOneWidget);

      server.commentsStatus = 200;
      server.comments = [commentJson(content: 'Hello there')];
      await tester.tap(find.text('Try again'));
      await tester.pumpAndSettle();

      expect(find.text('Hello there'), findsOneWidget);
    });

    testWidgets('shows comments with the total, author, badges and time', (
      tester,
    ) async {
      final server = FakeVodServer()
        ..comments = [
          commentJson(id: 'c-1', userId: 'u-2', content: 'First'),
          commentJson(
            id: 'c-2',
            userId: 'u-1',
            username: 'streamer',
            content: 'By the owner and me',
            isEdited: true,
          ),
        ];
      await pumpVod(tester, server, FakeLivePlayers());

      expect(find.text('Comments (2)'), findsOneWidget);
      expect(find.text('First'), findsOneWidget);
      expect(find.text('someone'), findsOneWidget);
      expect(find.text('Owner'), findsOneWidget);
      expect(find.text('You'), findsOneWidget);
      expect(find.text('(edited)'), findsOneWidget);
      expect(find.text('3 minutes ago'), findsNWidgets(2));
    });

    testWidgets('"Show more" loads the next page of ten', (tester) async {
      final server = FakeVodServer()
        ..comments = [
          for (var i = 0; i < 12; i++)
            commentJson(id: 'c-$i', content: 'Comment number $i'),
        ];
      final app = await pumpVod(tester, server, FakeLivePlayers());

      expect(find.text('Comment number 9'), findsOneWidget);
      expect(find.text('Comment number 10'), findsNothing);

      await tester.tap(find.text('Show more'));
      await tester.pumpAndSettle();

      expect(find.text('Comment number 11'), findsOneWidget);
      expect(find.text('Show more'), findsNothing);
      final pages = app
          .requestsTo('${ApiEndpoints.vods}/vod-1/comments')
          .map((r) => r.queryParameters['page']);
      expect(pages, [0, 1]);
    });

    testWidgets('a deleted comment shows a notice instead of its text', (
      tester,
    ) async {
      final server = FakeVodServer()
        ..comments = [commentJson(content: '', isDeleted: true)];
      await pumpVod(tester, server, FakeLivePlayers());

      expect(find.text('This comment has been deleted.'), findsOneWidget);
      expect(_like('c-1'), findsNothing);
    });
  });

  group('timestamps', () {
    testWidgets('tapping 01:23 seeks to 83s', (tester) async {
      final server = FakeVodServer()
        ..comments = [commentJson(content: 'Best part at 01:23, wow')];
      final players = FakeLivePlayers();
      await pumpVod(tester, server, players);

      await tester.tapOnText(find.textRange.ofSubstring('01:23'));
      await tester.pumpAndSettle();

      expect(players.last.seeks, [const Duration(seconds: 83)]);
      expect(players.last.value.value.isPlaying, isTrue);
    });

    testWidgets('h:mm:ss seeks too, a time past the video is plain', (
      tester,
    ) async {
      final server = FakeVodServer()
        ..comments = [commentJson(content: 'At 1:02:03 or 9:59:59 maybe')];
      final players = FakeLivePlayers();
      await pumpVod(tester, server, players);

      await tester.tapOnText(find.textRange.ofSubstring('1:02:03'));
      await tester.pumpAndSettle();
      expect(players.last.seeks, [const Duration(seconds: 3723)]);

      // 9:59:59 is longer than the 3725s video: not a link.
      await tester.tapOnText(find.textRange.ofSubstring('9:59:59'));
      await tester.pumpAndSettle();
      expect(players.last.seeks, hasLength(1));
    });

    testWidgets('long Vietnamese and emoji comment wraps at 320px, scale 2', (
      tester,
    ) async {
      final long = 'Rất hay 🎉🎉 không dấu 01:23 ' * 15;
      final server = FakeVodServer()
        ..comments = [
          commentJson(
            userId: 'u-1',
            username: 'ten_nguoi_dung_rat_dai_khong_co_dau_cach_nao',
            content: long,
            likeCount: 1234,
            replyCount: 5,
          ),
          commentJson(id: 'c-2', content: 'x' * 400),
        ];
      await pumpVod(
        tester,
        server,
        FakeLivePlayers(),
        width: 320,
        height: 568,
        textScale: 2,
      );
      await tester.scrollUntilVisible(
        find.text('View 5 replies'),
        200,
        scrollable: find.byType(Scrollable).first,
      );

      expect(tester.takeException(), isNull);
    });
  });

  group('likes', () {
    testWidgets('liked comments come from liked-ids and show a filled heart', (
      tester,
    ) async {
      final server = FakeVodServer()
        ..comments = [
          commentJson(id: 'c-1', likeCount: 4),
          commentJson(id: 'c-2'),
        ]
        ..likedIds = {'c-1'};
      final app = await pumpVod(tester, server, FakeLivePlayers());

      expect(_heart('c-1', Icons.favorite), findsOneWidget);
      expect(_heart('c-2', Icons.favorite_border), findsOneWidget);
      final asked = app
          .requestsTo('${ApiEndpoints.vodComments}/liked-ids')
          .single
          .data;
      expect(asked, {
        'commentIds': ['c-1', 'c-2'],
      });
    });

    testWidgets('like is optimistic and rolls back on failure', (tester) async {
      final server = FakeVodServer()
        ..comments = [commentJson(likeCount: 4)]
        ..holdLike = Completer<void>()
        ..writeStatus = 503;
      await pumpVod(tester, server, FakeLivePlayers());
      expect(
        find.descendant(of: _like('c-1'), matching: find.text('4')),
        findsOne,
      );

      await tester.tap(_like('c-1'));
      await tester.pump();
      // Shown before the server answered.
      expect(_heart('c-1', Icons.favorite), findsOneWidget);
      expect(
        find.descendant(of: _like('c-1'), matching: find.text('5')),
        findsOne,
      );

      server.holdLike!.complete();
      await tester.pumpAndSettle();

      expect(_heart('c-1', Icons.favorite_border), findsOneWidget);
      expect(
        find.descendant(of: _like('c-1'), matching: find.text('4')),
        findsOne,
      );
      expect(find.text('Something went wrong.'), findsOneWidget);
    });

    testWidgets('a successful like stays and unlike goes back', (tester) async {
      final server = FakeVodServer()..comments = [commentJson(likeCount: 4)];
      final app = await pumpVod(tester, server, FakeLivePlayers());

      await tester.tap(_like('c-1'));
      await tester.pumpAndSettle();
      expect(_heart('c-1', Icons.favorite), findsOneWidget);
      expect(server.likedIds, {'c-1'});

      await tester.tap(_like('c-1'));
      await tester.pumpAndSettle();
      expect(_heart('c-1', Icons.favorite_border), findsOneWidget);
      expect(server.likedIds, isEmpty);
      final methods = app.adapter.requests
          .where((r) => r.path == '${ApiEndpoints.vodComments}/c-1/like')
          .map((r) => r.method);
      expect(methods, ['POST', 'DELETE']);
    });
  });

  group('own comments', () {
    testWidgets('edit and delete are offered only on own comments', (
      tester,
    ) async {
      final server = FakeVodServer()
        ..comments = [
          commentJson(id: 'c-1', userId: 'u-2', content: 'Not mine'),
          commentJson(
            id: 'c-2',
            userId: 'u-1',
            username: 'me',
            content: 'Mine',
          ),
        ];
      await pumpVod(tester, server, FakeLivePlayers());

      expect(_menus, findsOneWidget);
      await tester.tap(_menus);
      await tester.pumpAndSettle();
      expect(find.text('Edit'), findsOneWidget);
      expect(find.text('Delete'), findsOneWidget);
    });

    testWidgets('edit saves the new text and marks it edited', (tester) async {
      final server = FakeVodServer()
        ..comments = [commentJson(userId: 'u-1', content: 'Old text')];
      await pumpVod(tester, server, FakeLivePlayers());

      await tester.tap(_menus);
      await tester.pumpAndSettle();
      await tester.tap(find.text('Edit'));
      await tester.pumpAndSettle();
      await tester.enterText(find.byType(TextField).last, '  New text ');
      await tester.pump();
      await tester.tap(find.text('Save'));
      await tester.pumpAndSettle();

      expect(server.editBodies, [
        {'content': 'New text'},
      ]);
      expect(find.text('New text'), findsOneWidget);
      expect(find.text('(edited)'), findsOneWidget);
      expect(find.text('Save'), findsNothing);
    });

    testWidgets('a failed edit keeps the editor and the typed text', (
      tester,
    ) async {
      final server = FakeVodServer()
        ..comments = [commentJson(userId: 'u-1', content: 'Old text')]
        ..writeStatus = 503;
      await pumpVod(tester, server, FakeLivePlayers());

      await tester.tap(_menus);
      await tester.pumpAndSettle();
      await tester.tap(find.text('Edit'));
      await tester.pumpAndSettle();
      await tester.enterText(find.byType(TextField).last, 'Changed');
      await tester.pump();
      await tester.tap(find.text('Save'));
      await tester.pumpAndSettle();

      expect(find.text('Something went wrong.'), findsOneWidget);
      expect(find.text('Changed'), findsOneWidget);
    });

    testWidgets('delete asks in an in-app dialog, then removes the text', (
      tester,
    ) async {
      final server = FakeVodServer()
        ..comments = [commentJson(userId: 'u-1', content: 'Bye')];
      await pumpVod(tester, server, FakeLivePlayers());
      expect(find.text('Comments (1)'), findsOneWidget);

      await tester.tap(_menus);
      await tester.pumpAndSettle();
      await tester.tap(find.text('Delete'));
      await tester.pumpAndSettle();
      expect(find.byType(AlertDialog), findsOneWidget);
      expect(find.text('Delete comment?'), findsOneWidget);

      // Cancel keeps it.
      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();
      expect(server.deletedIds, isEmpty);
      expect(find.text('Bye'), findsOneWidget);

      await tester.tap(_menus);
      await tester.pumpAndSettle();
      await tester.tap(find.text('Delete'));
      await tester.pumpAndSettle();
      await tester.tap(
        find.descendant(
          of: find.byType(AlertDialog),
          matching: find.text('Delete'),
        ),
      );
      await tester.pumpAndSettle();

      expect(server.deletedIds, ['c-1']);
      expect(find.text('Bye'), findsNothing);
      expect(find.text('This comment has been deleted.'), findsOneWidget);
      expect(find.text('Comments'), findsOneWidget, reason: 'total is 0 now');
    });
  });

  group('writing', () {
    testWidgets('posting adds the comment on top, by me, trimmed', (
      tester,
    ) async {
      final server = FakeVodServer()
        ..comments = [commentJson(content: 'Older one')];
      await pumpVod(tester, server, FakeLivePlayers());

      await tester.enterText(find.byType(TextField).first, '  Hay quá 🎉  ');
      await tester.pump();
      await tester.tap(find.text('Post'));
      await tester.pumpAndSettle();

      expect(server.createdBodies, [
        {'content': 'Hay quá 🎉'},
      ]);
      expect(find.text('Comments (2)'), findsOneWidget);
      expect(find.text('Hay quá 🎉'), findsOneWidget);
      // The server's answer has no author: the app fills in "me".
      expect(find.text('abcdef'), findsOneWidget);
      expect(find.text('You'), findsOneWidget);
    });

    testWidgets('post is disabled for blank text and shows chars left', (
      tester,
    ) async {
      await pumpVod(tester, FakeVodServer(), FakeLivePlayers());
      final post = find.widgetWithText(FilledButton, 'Post');
      expect(tester.widget<FilledButton>(post).onPressed, isNull);

      await tester.enterText(find.byType(TextField).first, '   ');
      await tester.pump();
      expect(tester.widget<FilledButton>(post).onPressed, isNull);

      await tester.enterText(find.byType(TextField).first, 'abc');
      await tester.pump();
      expect(tester.widget<FilledButton>(post).onPressed, isNotNull);
      expect(find.text('1997 characters remaining'), findsOneWidget);
    });

    testWidgets('text is cut at 2000 characters', (tester) async {
      await pumpVod(tester, FakeVodServer(), FakeLivePlayers());

      await tester.enterText(find.byType(TextField).first, 'a' * 2500);
      await tester.pump();

      expect(find.text('0 characters remaining'), findsOneWidget);
    });

    testWidgets('a failed post keeps the typed text', (tester) async {
      final server = FakeVodServer()..writeStatus = 503;
      await pumpVod(tester, server, FakeLivePlayers());

      await tester.enterText(find.byType(TextField).first, 'Keep me');
      await tester.pump();
      await tester.tap(find.text('Post'));
      await tester.pumpAndSettle();

      expect(find.text('Something went wrong.'), findsOneWidget);
      expect(find.text('Keep me'), findsOneWidget);
    });
  });

  group('replies', () {
    FakeVodServer withReplies() => FakeVodServer()
      ..comments = [commentJson(id: 'c-1', replyCount: 2, content: 'Parent')]
      ..replies = {
        'c-1': [
          commentJson(id: 'r-1', parentId: 'c-1', content: 'First reply'),
          commentJson(
            id: 'r-2',
            parentId: 'c-1',
            userId: 'u-1',
            username: 'me',
            content: 'My reply',
          ),
        ],
      };

    testWidgets('replies load on demand, not with the comments', (
      tester,
    ) async {
      final app = await pumpVod(tester, withReplies(), FakeLivePlayers());

      expect(find.text('First reply'), findsNothing);
      expect(
        app.requestsTo('${ApiEndpoints.vodComments}/c-1/replies'),
        isEmpty,
      );

      await tester.tap(find.text('View 2 replies'));
      await tester.pumpAndSettle();

      expect(find.text('First reply'), findsOneWidget);
      expect(find.text('My reply'), findsOneWidget);
      expect(find.text('View 2 replies'), findsNothing);
      final request = app
          .requestsTo('${ApiEndpoints.vodComments}/c-1/replies')
          .single;
      expect(request.queryParameters, {'page': 0, 'limit': 20});
    });

    testWidgets('only my own reply has edit and delete', (tester) async {
      await pumpVod(tester, withReplies(), FakeLivePlayers());
      await tester.tap(find.text('View 2 replies'));
      await tester.pumpAndSettle();

      expect(_menus, findsOneWidget);
    });

    testWidgets('replying adds the reply and grows the count', (tester) async {
      final server = withReplies();
      await pumpVod(tester, server, FakeLivePlayers());

      await tester.tap(find.text('Reply').first);
      await tester.pumpAndSettle();
      await tester.enterText(find.byType(TextField).last, 'Thanks!');
      await tester.pump();
      await tester.tap(find.text('Post').last);
      await tester.pumpAndSettle();

      expect(server.createdBodies, [
        {'content': 'Thanks!', 'parentId': 'c-1'},
      ]);
      expect(find.text('Thanks!'), findsOneWidget);
    });

    testWidgets('deleting a reply keeps a notice', (tester) async {
      final server = withReplies();
      await pumpVod(tester, server, FakeLivePlayers());
      await tester.tap(find.text('View 2 replies'));
      await tester.pumpAndSettle();

      await tester.tap(_menus);
      await tester.pumpAndSettle();
      await tester.tap(find.text('Delete'));
      await tester.pumpAndSettle();
      await tester.tap(
        find.descendant(
          of: find.byType(AlertDialog),
          matching: find.text('Delete'),
        ),
      );
      await tester.pumpAndSettle();

      expect(server.deletedIds, ['r-2']);
      expect(find.text('This comment has been deleted.'), findsOneWidget);
    });

    testWidgets('a failed replies load shows the reason and can retry', (
      tester,
    ) async {
      final server = withReplies()..repliesStatus = 503;
      await pumpVod(tester, server, FakeLivePlayers());

      await tester.tap(find.text('View 2 replies'));
      await tester.pumpAndSettle();
      expect(find.text('Something went wrong.'), findsOneWidget);
      expect(find.text('First reply'), findsNothing);

      server.repliesStatus = 200;
      await tester.tap(find.text('View 2 replies'));
      await tester.pumpAndSettle();
      expect(find.text('First reply'), findsOneWidget);
      expect(find.text('Something went wrong.'), findsNothing);
    });
  });
}
