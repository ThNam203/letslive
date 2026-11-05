import 'dart:convert';

import 'package:cookie_jar/cookie_jar.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_client.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/features/auth/data/auth_repository.dart';
import 'package:letslive_mobile/features/auth/data/bot_proof.dart';
import 'package:letslive_mobile/features/user/data/me_repository.dart';
import 'package:letslive_mobile/features/user/data/me_user.dart';
import 'package:letslive_mobile/features/user/data/picked_image.dart';
import 'package:letslive_mobile/features/user/data/user_repository.dart';
import 'package:letslive_mobile/features/vod/data/vod_repository.dart';

import 'harness.dart';
import 'png.dart';

/// A 100x100 solid-colour JPEG (quality 50), the one format besides WebP the
/// server accepts for a profile picture.
const String _jpeg100 =
    '/9j/4AAQSkZJRgABAQAAAQABAAD/2wBDABALDA4MChAODQ4SERATGCgaGBYWGDEjJR0oOjM9PDkzODdASFxOQERXRTc4UG1RV19iZ2hnPk1xeXBkeFxlZ2P/2wBDARESEhgVGC8aGi9jQjhCY2NjY2NjY2NjY2NjY2NjY2NjY2NjY2NjY2NjY2NjY2NjY2NjY2NjY2NjY2NjY2NjY2P/wAARCABkAGQDASIAAhEBAxEB/8QAHwAAAQUBAQEBAQEAAAAAAAAAAAECAwQFBgcICQoL/8QAtRAAAgEDAwIEAwUFBAQAAAF9AQIDAAQRBRIhMUEGE1FhByJxFDKBkaEII0KxwRVS0fAkM2JyggkKFhcYGRolJicoKSo0NTY3ODk6Q0RFRkdISUpTVFVWV1hZWmNkZWZnaGlqc3R1dnd4eXqDhIWGh4iJipKTlJWWl5iZmqKjpKWmp6ipqrKztLW2t7i5usLDxMXGx8jJytLT1NXW19jZ2uHi4+Tl5ufo6erx8vP09fb3+Pn6/8QAHwEAAwEBAQEBAQEBAQAAAAAAAAECAwQFBgcICQoL/8QAtREAAgECBAQDBAcFBAQAAQJ3AAECAxEEBSExBhJBUQdhcRMiMoEIFEKRobHBCSMzUvAVYnLRChYkNOEl8RcYGRomJygpKjU2Nzg5OkNERUZHSElKU1RVVldYWVpjZGVmZ2hpanN0dXZ3eHl6goOEhYaHiImKkpOUlZaXmJmaoqOkpaanqKmqsrO0tba3uLm6wsPExcbHyMnK0tPU1dbX2Nna4uPk5ebn6Onq8vP09fb3+Pn6/9oADAMBAAIRAxEAPwCnRRRWZ9AFFFFABRRRQAUUUUAFFFFABRRRQAUUUUAFFFFABRRRQAUUUUAFFFFABRRRQAUUUUAFFFFABRRRQAUUUUAFFFFABRRRQAUUUUAFFFFABRRRQAUUUUAFFFFABRRRQAUUUUAFFFFABRRRQAUUUUAFFFFABRRRQAUUUUAFFFFABRRRQAUUUUAFFFFABRRRQAUUUUAFFFFABRRRQAUUUUAFFFFABRRRQAUUUUAFFFFABRRRQAUUUUAFFFFABRRRQAUUUUAf/9k=';

/// Profiles against the real dev stack: public user, follow / unfollow, own
/// profile update, picture uploads, change password and the channel's
/// videos. User A is the profile owner, user B the visitor.
void main() {
  ContractContext? ctx;
  late String aId;
  late UserRepository asVisitor;
  late MeRepository meA;
  final insertedVods = <String>[];

  setUpAll(() async {
    final context = ctx = await startContractContext();
    final me = await context.userA.client.get<MeUser>(
      ApiEndpoints.userMe,
      parse: MeUser.parse,
    );
    aId = me.data!.id;
    asVisitor = UserRepository(context.userB.client);
    meA = MeRepository(context.userA.client);
  });

  tearDownAll(() async {
    try {
      for (final id in insertedVods) {
        await runSql('letslive_vod', "DELETE FROM vods WHERE id = '$id'");
      }
    } finally {
      await ctx?.dispose();
    }
  });

  test('GET /user/<id> parses the public profile for a visitor', () async {
    final res = await asVisitor.byId(aId);

    expect(res.success, isTrue, reason: '${res.key}: ${res.message}');
    final user = res.data!;
    expect(user.id, aId);
    expect(user.username, ctx!.userA.username);
    expect(user.followerCount, 0);
    expect(user.isFollowing, isFalse);
    expect(user.createdAt.year, greaterThan(2000));
  });

  test('follow and unfollow between the two users', () async {
    try {
      final follow = await asVisitor.follow(aId);
      expect(follow.success, isTrue, reason: '${follow.key} ${follow.message}');

      final after = (await asVisitor.byId(aId)).data!;
      expect(after.followerCount, 1);
      expect(after.isFollowing, isTrue);
      final following = (await asVisitor.following()).data!;
      expect(following.map((u) => u.id), contains(aId));

      final unfollow = await asVisitor.unfollow(aId);
      expect(
        unfollow.success,
        isTrue,
        reason: '${unfollow.key} ${unfollow.message}',
      );
      final back = (await asVisitor.byId(aId)).data!;
      expect(back.followerCount, 0);
      expect(back.isFollowing, isFalse);
    } finally {
      await asVisitor.unfollow(aId);
    }
  });

  test('PUT /user/me updates bio and a social link', () async {
    final res = await meA.update(
      bio: 'Contract bio Việt 🎉',
      socialMediaLinks: {'github': 'https://github.com/octocat'},
    );

    expect(res.success, isTrue, reason: '${res.key}: ${res.message}');
    expect(res.data?.bio, 'Contract bio Việt 🎉');
    final me = (await meA.fetch()).data!;
    expect(me.socialMediaLinks['github'], 'https://github.com/octocat');
    final public = (await asVisitor.byId(aId)).data!;
    expect(public.bio, 'Contract bio Việt 🎉');
    expect(public.socialMediaLinks['github'], 'https://github.com/octocat');
  });

  test(
    'the server refuses an empty social link (so the app does too)',
    () async {
      final res = await meA.update(socialMediaLinks: {'github': ''});

      expect(res.success, isFalse);
      expect(res.key, 'res_err_invalid_input');
    },
  );

  test('PUT /user/me with a too short username is refused', () async {
    final res = await meA.update(username: 'abc');

    expect(res.success, isFalse);
    expect(res.statusCode, 400);
  });

  test('multipart upload of a generated PNG as the banner', () async {
    final image = PickedImage(
      name: 'banner.png',
      bytes: solidPng(200, 80),
      mimeType: 'image/png',
    );
    var lastSent = 0;

    final res = await meA.uploadBackgroundPicture(
      image,
      onProgress: (sent, total) => lastSent = sent,
    );

    expect(res.success, isTrue, reason: '${res.key}: ${res.message}');
    expect(res.data, startsWith('http'));
    expect(lastSent, greaterThan(0));
    expect((await meA.fetch()).data?.backgroundPicture, res.data);
  });

  test('multipart upload of a JPEG as the profile picture', () async {
    final image = PickedImage(
      name: 'avatar.jpg',
      bytes: base64Decode(_jpeg100),
      mimeType: 'image/jpeg',
    );

    final res = await meA.uploadProfilePicture(image);

    expect(res.success, isTrue, reason: '${res.key}: ${res.message}');
    expect(res.data, startsWith('http'));
    expect((await meA.fetch()).data?.profilePicture, res.data);
  });

  test(
    'a PNG profile picture is refused: the server decodes JPEG/WebP',
    () async {
      final res = await meA.uploadProfilePicture(
        PickedImage(
          name: 'avatar.png',
          bytes: solidPng(100, 100),
          mimeType: 'image/png',
        ),
      );

      // Why the app checks the type before uploading (FieldLimits).
      expect(res.success, isFalse);
      expect(res.key, 'res_err_invalid_image');
    },
  );

  test('a profile picture below 80x80 is refused with the web key', () async {
    final res = await meA.uploadProfilePicture(
      PickedImage(
        name: 'tiny.png',
        bytes: solidPng(40, 40),
        mimeType: 'image/png',
      ),
    );

    expect(res.success, isFalse);
    expect(
      res.key,
      anyOf('res_err_image_dimensions_out_of_range', 'res_err_invalid_image'),
    );
  });

  test('GET /vods?userId= pages with the channel page size', () async {
    for (var i = 0; i < 3; i++) {
      insertedVods.add(
        (await runSql(
          'letslive_vod',
          "INSERT INTO vods (user_id, title, visibility, duration, "
              "playback_url) VALUES ('$aId', 'Channel vod $i', 'public', 60, "
              "'http://localhost:8000/files/vods/x/index.m3u8') RETURNING id",
        )).split('\n').first,
      );
    }
    final repo = VodRepository(ctx!.userB.client);

    final first = await repo.ofUser(aId, limit: 2);
    final second = await repo.ofUser(aId, page: 1, limit: 2);

    expect(first.success, isTrue, reason: '${first.key}: ${first.message}');
    expect(first.data, hasLength(2));
    expect(second.data, hasLength(1));
    expect({...first.data!, ...second.data!}.map((v) => v.id).toSet(), {
      ...insertedVods,
    });
  });

  test('change password, then log in with the new one', () async {
    final user = ctx!.userB;
    final auth = AuthRepository(
      client: user.client,
      botProof: BotProof.turnstile(ctx!.config.turnstileDevToken ?? ''),
    );
    const next = 'Changed-Pass2!';

    final wrongOld = await auth.changePassword(
      oldPassword: 'Not-The-Pass1!',
      newPassword: next,
    );
    expect(wrongOld.success, isFalse);
    expect(wrongOld.key, 'res_err_password_not_match');

    final changed = await auth.changePassword(
      oldPassword: user.password,
      newPassword: next,
    );
    expect(
      changed.success,
      isTrue,
      reason: '${changed.key} ${changed.message}',
    );
    expect(changed.statusCode, 204);

    // The change does not revoke the session: the refresh token still works
    // (backend/auth `UpdatePassword` only rewrites the password hash).
    expect(await user.client.refreshSession(), isTrue);
    expect((await MeRepository(user.client).fetch()).success, isTrue);

    final fresh = ApiClient(
      baseUrl: ctx!.config.apiBaseUrl,
      cookieJar: CookieJar(),
    );
    try {
      final freshAuth = AuthRepository(
        client: fresh,
        botProof: BotProof.turnstile(ctx!.config.turnstileDevToken ?? ''),
      );
      final oldLogin = await freshAuth.login(user.email, user.password);
      expect(oldLogin.success, isFalse);
      final login = await freshAuth.login(user.email, next);
      expect(login.success, isTrue, reason: '${login.key} ${login.message}');
      expect((await MeRepository(fresh).fetch()).data?.email, user.email);
    } finally {
      await fresh.dispose();
    }
  });
}
