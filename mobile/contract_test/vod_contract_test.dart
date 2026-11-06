import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/features/user/data/me_user.dart';
import 'package:letslive_mobile/features/vod/data/vod_comment_repository.dart';
import 'package:letslive_mobile/features/vod/data/vod_reaction.dart';
import 'package:letslive_mobile/features/vod/data/vod_repository.dart';

import 'harness.dart';

/// The video screen's endpoints against the real dev stack: the video, view
/// count, reactions and every comment endpoint. A VOD normally comes from a
/// transcoded recording, so one public row is inserted straight into the vod
/// database (removed afterwards; its comments go with it, `ON DELETE
/// CASCADE`). User A owns the video, user B watches and comments.
void main() {
  ContractContext? ctx;
  late String ownerId;
  late String vodId;
  late VodRepository asViewer;
  late VodCommentRepository comments;
  late VodCommentRepository ownerComments;
  String? insertedVodId;

  setUpAll(() async {
    final context = ctx = await startContractContext();
    final me = await context.userA.client.get<MeUser>(
      ApiEndpoints.userMe,
      parse: MeUser.parse,
    );
    ownerId = me.data!.id;
    insertedVodId = vodId = (await runSql(
      'letslive_vod',
      'INSERT INTO vods (user_id, title, description, visibility, duration, '
          'playback_url) VALUES (\'$ownerId\', \'Contract vod Việt 🎉\', '
          '\'Intro at 01:23\', \'public\', 3725, '
          '\'http://localhost:8000/files/vods/x/index.m3u8\') RETURNING id',
    )).split('\n').first;
    asViewer = VodRepository(context.userB.client);
    comments = VodCommentRepository(context.userB.client);
    ownerComments = VodCommentRepository(context.userA.client);
  });

  tearDownAll(() async {
    try {
      final id = insertedVodId;
      if (id != null) {
        await runSql('letslive_vod', "DELETE FROM vods WHERE id = '$id'");
      }
    } finally {
      await ctx?.dispose();
    }
  });

  test('GET /vods/<id> parses the video', () async {
    final res = await asViewer.byId(vodId);

    expect(res.success, isTrue, reason: '${res.key}: ${res.message}');
    final vod = res.data!;
    expect(vod.id, vodId);
    expect(vod.userId, ownerId);
    expect(vod.title, 'Contract vod Việt 🎉');
    expect(vod.description, 'Intro at 01:23');
    expect(vod.duration, 3725);
    expect(vod.playbackUrl, 'http://localhost:8000/files/vods/x/index.m3u8');
  });

  test(
    'GET /vods/<id> answers 404 for an unknown id, 400 for a bad one',
    () async {
      final unknown = await asViewer.byId(
        '00000000-0000-4000-8000-000000000000',
      );
      expect(unknown.success, isFalse);
      expect(unknown.statusCode, 404);
      expect(unknown.key, 'res_err_vod_not_found');

      final malformed = await asViewer.byId('not-a-uuid');
      expect(malformed.success, isFalse);
      expect(malformed.statusCode, 400);
    },
  );

  test('GET /vods?userId= lists the uploader\'s public videos', () async {
    final res = await asViewer.ofUser(ownerId);

    expect(res.success, isTrue, reason: '${res.key}: ${res.message}');
    expect(res.data!.map((v) => v.id), contains(vodId));
  });

  test(
    'POST /vods/<id>/view refuses too little watching, counts enough',
    () async {
      final tooShort = await asViewer.registerView(vodId, 1);
      expect(tooShort.success, isFalse);
      expect(tooShort.key, 'res_err_vod_view_threshold');

      final enough = await asViewer.registerView(vodId, 15);
      expect(
        enough.success,
        isTrue,
        reason: '${enough.key}: ${enough.message}',
      );
      final vod = await asViewer.byId(vodId);
      expect(vod.data!.viewCount, 1);
    },
  );

  test('reaction: none at first, like, dislike, remove', () async {
    final none = await asViewer.myReaction(vodId);
    expect(none.success, isTrue, reason: '${none.key}: ${none.message}');
    expect(none.data!.reaction, isNull);
    expect(none.data!.likeCount, 0);

    final liked = await asViewer.react(vodId, VodReaction.like);
    expect(liked.data!.reaction, VodReaction.like);
    expect(liked.data!.likeCount, 1);

    final disliked = await asViewer.react(vodId, VodReaction.dislike);
    expect(disliked.data!.reaction, VodReaction.dislike);
    expect(disliked.data!.likeCount, 0);

    final removed = await asViewer.removeReaction(vodId);
    expect(removed.data!.reaction, isNull);
    expect(removed.data!.likeCount, 0);
  });

  test('comments: create, list, edit, like, reply, delete', () async {
    // create
    final created = await comments.create(vodId, 'Hay quá 🎉 at 01:23');
    expect(
      created.success,
      isTrue,
      reason: '${created.key}: ${created.message}',
    );
    final comment = created.data!;
    expect(comment.vodId, vodId);
    expect(comment.content, 'Hay quá 🎉 at 01:23');
    expect(comment.parentId, isNull);
    expect(comment.isDeleted, isFalse);
    expect(comment.replyCount, 0);

    // list: parses with its author and reports the total
    final list = await comments.list(vodId);
    expect(list.success, isTrue, reason: '${list.key}: ${list.message}');
    expect(list.meta!.total, 1);
    final listed = list.data!.single;
    expect(listed.id, comment.id);
    expect(listed.user!.username, ctx!.userB.username);

    // edit
    final edited = await comments.edit(comment.id, 'Edited text');
    expect(edited.success, isTrue, reason: '${edited.key}: ${edited.message}');
    expect(edited.data!.content, 'Edited text');
    expect(edited.data!.isEdited, isTrue);

    // like, liked-ids, unlike
    final like = await ownerComments.like(comment.id);
    expect(like.success, isTrue, reason: '${like.key}: ${like.message}');
    final liked = await ownerComments.likedIds([comment.id]);
    expect(liked.data, [comment.id]);
    expect((await comments.likedIds([comment.id])).data, isEmpty);
    expect((await comments.list(vodId)).data!.single.likeCount, 1);
    final unlike = await ownerComments.unlike(comment.id);
    expect(unlike.success, isTrue, reason: '${unlike.key}: ${unlike.message}');
    expect((await ownerComments.likedIds([comment.id])).data, isEmpty);

    // reply by the owner, then the replies list
    final reply = await ownerComments.create(
      vodId,
      'Thanks!',
      parentId: comment.id,
    );
    expect(reply.success, isTrue, reason: '${reply.key}: ${reply.message}');
    expect(reply.data!.parentId, comment.id);
    final replies = await comments.replies(comment.id);
    expect(
      replies.success,
      isTrue,
      reason: '${replies.key}: ${replies.message}',
    );
    expect(replies.data!.single.id, reply.data!.id);
    expect(replies.data!.single.user!.id, ownerId);
    expect((await comments.list(vodId)).data!.single.replyCount, 1);

    // only the author may edit or delete
    final foreign = await comments.delete(reply.data!.id);
    expect(foreign.success, isFalse);
    final foreignEdit = await comments.edit(reply.data!.id, 'hijack');
    expect(foreignEdit.success, isFalse);

    // delete the reply, then the comment: it stays as "deleted", no text
    final deleteReply = await ownerComments.delete(reply.data!.id);
    expect(deleteReply.success, isTrue, reason: deleteReply.key);
    final deleted = await comments.delete(comment.id);
    expect(
      deleted.success,
      isTrue,
      reason: '${deleted.key}: ${deleted.message}',
    );
    final after = (await comments.list(vodId)).data!.single;
    expect(after.isDeleted, isTrue);
    expect(after.content, isEmpty);
  });

  test(
    'comment pages hold $commentsPageSize and the next page follows',
    () async {
      for (var i = 0; i < commentsPageSize + 2; i++) {
        final res = await comments.create(vodId, 'Paged comment $i');
        expect(res.success, isTrue, reason: '${res.key}: ${res.message}');
      }

      final first = await comments.list(vodId);
      expect(first.data, hasLength(commentsPageSize));
      final second = await comments.list(vodId, page: 1);
      expect(second.success, isTrue);
      expect(second.data!.length, greaterThanOrEqualTo(2));
      final ids = {
        ...first.data!.map((c) => c.id),
        ...second.data!.map((c) => c.id),
      };
      expect(ids.length, first.data!.length + second.data!.length);
    },
  );
}
