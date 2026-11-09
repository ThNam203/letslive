import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/features/vod/domain/timestamps.dart';

/// Ports of web's `splitTimestamps` rules (`web/utils/vod-timestamps.ts`).
void main() {
  List<String> kinds(List<TextSegment> parts) => [
    for (final part in parts)
      switch (part) {
        PlainSegment(:final text) => 'text:$text',
        TimestampSegment(:final text, :final seconds) => 'ts:$text=$seconds',
      },
  ];

  test('mm:ss and h:mm:ss become timestamps', () {
    expect(kinds(splitTimestamps('go to 01:23 or 1:02:03 now')), [
      'text:go to ',
      'ts:01:23=83',
      'text: or ',
      'ts:1:02:03=3723',
      'text: now',
    ]);
  });

  test('plain text without timestamps is one segment', () {
    expect(kinds(splitTimestamps('nothing here')), ['text:nothing here']);
    expect(splitTimestamps(''), isEmpty);
  });

  test('a timestamp at the very start and end has no empty text around it', () {
    expect(kinds(splitTimestamps('0:05')), ['ts:0:05=5']);
  });

  test('timestamps glued to word characters, colons or dots stay text', () {
    for (final text in ['a1:23', '1:23:4', '16:9 ratio', '1:234']) {
      expect(
        splitTimestamps(text).whereType<TimestampSegment>(),
        isEmpty,
        reason: text,
      );
    }
  });

  test('a timestamp followed by sentence punctuation still links', () {
    expect(kinds(splitTimestamps('see 1:23.')), [
      'text:see ',
      'ts:1:23=83',
      'text:.',
    ]);
  });

  test('seconds must be below 60, and so must minutes after an hour', () {
    expect(splitTimestamps('1:75').whereType<TimestampSegment>(), isEmpty);
    expect(splitTimestamps('1:75:00').whereType<TimestampSegment>(), isEmpty);
    expect(kinds(splitTimestamps('75:00')), [
      'ts:75:00=4500',
    ], reason: 'without hours, minutes may go past 59 like web');
  });

  test('timestamps past the video length stay plain text', () {
    expect(kinds(splitTimestamps('at 2:00 and 0:30', maxSeconds: 60)), [
      'text:at 2:00 and ',
      'ts:0:30=30',
    ]);
    expect(
      splitTimestamps('2:00', maxSeconds: 0).whereType<TimestampSegment>(),
      hasLength(1),
      reason: 'an unknown length (0) limits nothing',
    );
  });

  test('Vietnamese and emoji around a timestamp keep their text intact', () {
    expect(kinds(splitTimestamps('Hay quá 🎉 ở 01:23 nhé')), [
      'text:Hay quá 🎉 ở ',
      'ts:01:23=83',
      'text: nhé',
    ]);
  });
}
