import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/push/push_tap.dart';
import 'package:letslive_mobile/core/router/app_routes.dart';

void main() {
  group('resolveDeepLink maps each notification type', () {
    test('a DM opens its conversation by actionUrl or by referenceId', () {
      expect(
        resolveDeepLink(
          const PushTap(
            type: 'dm',
            referenceId: 'c-1',
            actionUrl: '/messages/c-1',
          ),
        ),
        '/messages/c-1',
      );
      expect(
        resolveDeepLink(const PushTap(type: 'dm', referenceId: 'c-1')),
        '/messages/c-1',
      );
    });

    test('a comment and a reply open the video', () {
      for (final type in ['vod_comment', 'vod_comment_reply']) {
        expect(
          resolveDeepLink(
            PushTap(
              type: type,
              referenceId: 'comment-1',
              actionUrl: '/users/u-1/vods/v-1',
            ),
          ),
          '/users/u-1/vods/v-1',
        );
      }
    });

    test('a channel url opens the channel', () {
      expect(
        resolveDeepLink(const PushTap(actionUrl: '/users/u-1')),
        '/users/u-1',
      );
    });

    test('a gift, whose page the app has no screen for, falls back', () {
      expect(
        resolveDeepLink(
          const PushTap(
            type: 'gift_received',
            referenceId: 'g-1',
            actionUrl: '/users/u-1/gifts',
          ),
        ),
        AppRoutes.notifications,
      );
    });

    test('an unknown type without a usable url falls back', () {
      expect(
        resolveDeepLink(const PushTap(type: 'new_thing')),
        '/notifications',
      );
      expect(
        resolveDeepLink(const PushTap(type: 'vod_comment', referenceId: 'x')),
        '/notifications',
      );
      expect(resolveDeepLink(const PushTap()), '/notifications');
    });
  });

  group('an actionUrl never leaves the app routes', () {
    const hostile = [
      'https://evil.example/messages/c-1',
      '//evil.example/messages/c-1',
      'javascript:alert(1)',
      'messages/c-1',
      '/messages/new',
      '/messages/c-1/extra',
      '/messages/',
      '/users//vods/v-1',
      '/users/u-1/vods',
      '/settings',
      '/login',
      '/account-setup',
      '/../settings',
      '/users/..',
      '/users/./vods/v-1',
      '/messages/.',
      '/users/u-1/vods/..',
      '',
    ];

    for (final url in hostile) {
      test('"$url" falls back to the notifications screen', () {
        expect(
          resolveDeepLink(PushTap(actionUrl: url)),
          AppRoutes.notifications,
        );
      });
    }

    test('ids are rebuilt encoded, so they cannot add path parts', () {
      expect(
        resolveDeepLink(const PushTap(actionUrl: '/users/a%2Fb')),
        '/users/a%2Fb',
      );
    });

    test('a query or fragment is dropped', () {
      expect(
        resolveDeepLink(const PushTap(actionUrl: '/users/u-1/vods/v-1?t=5#x')),
        '/users/u-1/vods/v-1',
      );
    });
  });

  group('PushTap.fromData', () {
    test('reads the three keys and ignores values that are not text', () {
      final tap = PushTap.fromData({
        'type': 'dm',
        'referenceId': 12,
        'actionUrl': '/messages/c-1',
        'other': 'x',
      });

      expect(tap.type, 'dm');
      expect(tap.referenceId, isNull);
      expect(tap.actionUrl, '/messages/c-1');
    });
  });
}
