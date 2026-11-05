import 'dart:io';

import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/validation/field_limits.dart';
import 'package:letslive_mobile/features/auth/domain/auth_validators.dart';
import 'package:letslive_mobile/features/profile/domain/profile_format.dart';
import 'package:letslive_mobile/features/settings/domain/image_checks.dart';
import 'package:letslive_mobile/features/user/data/me_user.dart';
import 'package:letslive_mobile/features/user/data/public_user.dart';
import 'package:letslive_mobile/features/user/data/social_links.dart';

import 'profile_test_app.dart';

/// Reads `export const NAME = <number>;` from a web constants file.
int _web(String file, String name) {
  final source = File('../web/constant/$file').readAsStringSync();
  final match = RegExp('export const $name = (\\d+);').firstMatch(source);
  if (match == null) fail('$name not found in web/constant/$file');
  return int.parse(match.group(1)!);
}

void main() {
  test('FieldLimits equal the web constants they cite', () {
    expect(FieldLimits.emailMax, _web('field-limits.ts', 'EMAIL_MAX_LENGTH'));
    expect(FieldLimits.passwordMin, _web('password.ts', 'PASSWORD_MIN_LENGTH'));
    expect(FieldLimits.passwordMax, _web('password.ts', 'PASSWORD_MAX_LENGTH'));
    expect(
      FieldLimits.usernameMin,
      _web('field-limits.ts', 'USERNAME_MIN_LENGTH'),
    );
    expect(
      FieldLimits.usernameMax,
      _web('field-limits.ts', 'USERNAME_MAX_LENGTH'),
    );
    expect(FieldLimits.bioMax, _web('field-limits.ts', 'BIO_MAX_LENGTH'));
    expect(
      FieldLimits.socialUrlMax,
      _web('field-limits.ts', 'SOCIAL_URL_MAX_LENGTH'),
    );
    expect(
      FieldLimits.searchQueryMax,
      _web('field-limits.ts', 'SEARCH_QUERY_MAX_LENGTH'),
    );
    expect(
      FieldLimits.chatMessageMax,
      _web('field-limits.ts', 'CHAT_MESSAGE_MAX_LENGTH'),
    );
    expect(
      FieldLimits.vodCommentMax,
      _web('field-limits.ts', 'VOD_COMMENT_MAX_LENGTH'),
    );
    expect(FieldLimits.avatarMaxFileMb, _web('image.ts', 'AVATAR_MAX_FILE_MB'));
    expect(
      FieldLimits.backgroundMaxFileMb,
      _web('image.ts', 'BACKGROUND_MAX_FILE_MB'),
    );
    expect(
      FieldLimits.avatarMinDimension,
      _web('image.ts', 'AVATAR_MIN_DIMENSION'),
    );
    expect(
      FieldLimits.avatarSourceMaxDimension,
      _web('image.ts', 'AVATAR_SOURCE_MAX_DIMENSION'),
    );
    final accepted = File('../web/constant/image.ts')
        .readAsStringSync()
        .contains('"image/jpeg,image/png,image/gif,image/webp"');
    expect(accepted, isTrue);
    expect(FieldLimits.avatarAcceptedTypes.join(','), contains('image/gif'));
  });

  group('validateChangePassword (web changePassword.ts)', () {
    ChangePasswordErrors run(String current, String next, String confirm) =>
        validateChangePassword(current: current, next: next, confirm: confirm);

    test('a valid change has no errors', () {
      final e = run('Old-pass-1!', 'New-pass-1!', 'New-pass-1!');
      expect((e.current, e.next, e.confirm), (null, null, null));
    });

    test('every field is 8..72 long', () {
      final long = 'a' * 73;
      final e = run('short', long, 'tiny');
      expect(e.current, AuthFieldError.passwordTooShort);
      expect(e.next, AuthFieldError.passwordTooLong);
      expect(e.confirm, AuthFieldError.passwordsDoNotMatch);
      expect(run('a' * 72, 'b' * 72, 'b' * 72).next, isNull);
    });

    test('the new password must differ from the current one', () {
      final e = run('Same-pass-1!', 'Same-pass-1!', 'Same-pass-1!');
      expect(e.next, AuthFieldError.newPasswordMustDiffer);
      expect(e.confirm, isNull);
    });

    test('the confirmation must match', () {
      expect(
        run('Old-pass-1!', 'New-pass-1!', 'New-pass-2!').confirm,
        AuthFieldError.passwordsDoNotMatch,
      );
    });
  });

  group('image checks (web image-hover and profile-banner)', () {
    test('size limit is in megabytes', () {
      const mb = 1024 * 1024;
      expect(checkFileSize(pickedImage(bytes: 10 * mb), 10), isNull);
      expect(
        checkFileSize(pickedImage(bytes: 10 * mb + 1), 10),
        ImageProblem.tooLarge,
      );
    });

    test('profile pictures must be a format the server decodes', () {
      for (final ok in ['image/jpeg', 'image/webp']) {
        expect(checkAvatarType(pickedImage(mimeType: ok)), isNull, reason: ok);
      }
      for (final bad in ['image/png', 'image/gif', 'image/heic', null]) {
        expect(
          checkAvatarType(pickedImage(mimeType: bad)),
          ImageProblem.unsupportedType,
          reason: '$bad',
        );
      }
    });

    test('profile pictures are 80..10000 px on both sides', () {
      expect(checkAvatarDimensions(80, 80), isNull);
      expect(checkAvatarDimensions(10000, 10000), isNull);
      expect(checkAvatarDimensions(79, 500), ImageProblem.dimensions);
      expect(checkAvatarDimensions(500, 10001), ImageProblem.dimensions);
    });

    test('the type comes from the extension when the picker has none', () {
      expect(mimeTypeOfName('a.JPG'), 'image/jpeg');
      expect(mimeTypeOfName('a.png'), 'image/png');
      expect(mimeTypeOfName('a.heic'), isNull);
      expect(mimeTypeOfName('noextension'), isNull);
    });
  });

  group('social links', () {
    test('unknown platforms and empty values are dropped, order is web\'s', () {
      final links = parseSocialLinks({
        'website': 'https://a.example',
        'github': 'https://github.com/x',
        'myspace': 'https://myspace.example',
        'facebook': '',
      }, 'Test');
      expect(links.keys, ['github', 'website']);
    });

    test('a missing value is no links; a wrong type throws', () {
      expect(parseSocialLinks(null, 'Test'), isEmpty);
      expect(() => parseSocialLinks('x', 'Test'), throwsFormatException);
    });

    test('only http(s) links get a label: host without www plus path', () {
      expect(
        socialLinkLabel('https://www.github.com/octocat'),
        'github.com/octocat',
      );
      expect(socialLinkLabel('http://example.com/'), 'example.com');
      expect(socialLinkLabel('javascript:alert(1)'), isNull);
      expect(socialLinkLabel('ftp://example.com/x'), isNull);
      expect(socialLinkLabel('not a url'), isNull);
    });
  });

  group('users parse the profile fields', () {
    test('PublicUser reads banner, join date, isFollowing and links', () {
      final user = PublicUser.fromJson({
        'id': 'u',
        'username': 'name',
        'followerCount': 2,
        'createdAt': '2026-01-15T00:00:00Z',
        'backgroundPicture': 'http://x/b.jpg',
        'isFollowing': true,
        'socialMediaLinks': {'github': 'https://github.com/x'},
      });
      expect(user.backgroundPicture, 'http://x/b.jpg');
      expect(user.isFollowing, isTrue);
      expect(user.socialMediaLinks, {'github': 'https://github.com/x'});
      expect(user.createdAt.year, 2026);
    });

    test('withFollowing moves the count by one and never below zero', () {
      final user = PublicUser.fromJson({
        'id': 'u',
        'username': 'name',
        'followerCount': 0,
        'createdAt': '2026-01-15T00:00:00Z',
      });
      expect(user.withFollowing(following: true).followerCount, 1);
      expect(user.withFollowing(following: false).followerCount, 0);
    });

    test('a missing join date is a malformed user', () {
      expect(
        () => PublicUser.fromJson({
          'id': 'u',
          'username': 'name',
          'followerCount': 0,
        }),
        throwsFormatException,
      );
    });

    test('MeUser reads social links and swaps pictures', () {
      final me = MeUser.fromJson({
        'id': 'u',
        'username': 'name',
        'email': 'a@b.c',
        'status': 'normal',
        'authProvider': 'local',
        'createdAt': '2026-01-15T00:00:00Z',
        'followerCount': 0,
        'streamAPIKey': 'k',
        'socialMediaLinks': {'website': 'https://a.example'},
      });
      expect(me.socialMediaLinks, {'website': 'https://a.example'});
      final changed = me.withPictures(profilePicture: 'http://x/p.jpg');
      expect(changed.profilePicture, 'http://x/p.jpg');
      expect(changed.backgroundPicture, isNull);
      expect(changed.socialMediaLinks, me.socialMediaLinks);
    });
  });
}
