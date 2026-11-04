import 'dart:async';

import 'package:better_player_plus/better_player_plus.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/features/watch/data/better_live_player_controller.dart';
import 'package:letslive_mobile/features/watch/domain/live_player_controller.dart';

BetterPlayerAsmsTrack track(int? height, {int bitrate = 1}) =>
    BetterPlayerAsmsTrack('', null, height, bitrate, 0, '', '');

void main() {
  test('qualities are the variant heights in playlist order', () {
    expect(qualitiesOf([track(1080), track(720), track(360)]), const [
      VideoQuality(1080),
      VideoQuality(720),
      VideoQuality(360),
    ]);
  });

  test('a height listed twice is offered once, like web picks the first', () {
    expect(
      qualitiesOf([track(720, bitrate: 1), track(720, bitrate: 2)]),
      const [VideoQuality(720)],
    );
  });

  test('tracks without a height are not offered', () {
    expect(
      qualitiesOf([BetterPlayerAsmsTrack.defaultTrack(), track(null)]),
      isEmpty,
    );
  });

  group('playerErrorZone', () {
    test(
      "ignores the package timer's RangeError, also in timers it starts",
      () async {
        final errors = <Object>[];
        final zone = playerErrorZone((error, _) => errors.add(error));

        zone.run(() {
          Timer(Duration.zero, () => throw RangeError('absolutePosition'));
        });
        await Future<void>.delayed(const Duration(milliseconds: 10));

        expect(errors, isEmpty);
      },
    );

    test('reports any other uncaught error', () async {
      final errors = <Object>[];
      final zone = playerErrorZone((error, _) => errors.add(error));

      zone.run(() {
        Timer(Duration.zero, () => throw StateError('decoder'));
      });
      await Future<void>.delayed(const Duration(milliseconds: 10));

      expect(errors, [isA<StateError>()]);
    });
  });
}
