import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/features/user/data/me_user.dart';

Map<String, Object?> _json() => {
  'id': '0b6c5f0e-0000-4000-8000-000000000001',
  'username': 'tester01',
  'email': 'tester01@example.com',
  'status': 'normal',
  'authProvider': 'local',
  'createdAt': '2026-11-04T05:00:00.123456Z',
  'streamAPIKey': '5a1e7c2e-0000-4000-8000-000000000002',
  'followerCount': 3,
  'livestreamInformation': {
    'title': null,
    'description': null,
    'thumbnailUrl': null,
  },
};

void main() {
  test('parses a full user and ignores unmodelled fields', () {
    final me = MeUser.parse(_json());

    expect(me.username, 'tester01');
    expect(me.followerCount, 3);
    expect(me.createdAt.isUtc, isTrue);
    expect(me.bio, isNull);
    expect(me.locale, isNull);
  });

  test('parses optional strings when present', () {
    final me = MeUser.parse({..._json(), 'bio': 'hi', 'locale': 'vi-VN'});

    expect(me.bio, 'hi');
    expect(me.locale, 'vi-VN');
  });

  test('throws FormatException on a type mismatch of a modelled field', () {
    expect(
      () => MeUser.parse({..._json(), 'followerCount': '3'}),
      throwsFormatException,
    );
    expect(() => MeUser.parse({..._json(), 'bio': 5}), throwsFormatException);
  });

  test('throws FormatException when a required field is missing', () {
    final json = _json()..remove('email');

    expect(() => MeUser.parse(json), throwsFormatException);
  });

  test('throws FormatException when data is not an object', () {
    expect(() => MeUser.parse(<Object?>[]), throwsFormatException);
  });
}
