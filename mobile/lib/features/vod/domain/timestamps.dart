/// Finds "1:23" / "01:23" / "1:02:03" timestamps in a video description or
/// comment so they can be tapped to seek. A port of web's `splitTimestamps`
/// (`web/utils/vod-timestamps.ts`); the rules are the same on purpose.
library;

import 'package:flutter/foundation.dart';

/// "1:23", "01:23" or "1:02:03" standing on their own: not glued to a word
/// character, colon or dot before it (`a1:23`, `1:23:4`, `16:9`), nor to a
/// word character or colon after it. Seconds are 00-59.
final RegExp _timestamp = RegExp(
  r'(?<![\w:.])(?:(\d{1,2}):)?(\d{1,2}):([0-5]\d)(?![\w:])',
);

/// One piece of a text split by [splitTimestamps].
@immutable
sealed class TextSegment {
  const TextSegment(this.text);

  /// The piece exactly as it appears in the source text.
  final String text;
}

/// Ordinary text.
final class PlainSegment extends TextSegment {
  const PlainSegment(super.text);
}

/// A timestamp such as `01:23`, [seconds] into the video.
final class TimestampSegment extends TextSegment {
  const TimestampSegment(super.text, this.seconds);

  /// Position the timestamp points to.
  final int seconds;
}

/// Splits [text] into plain runs and timestamps, in order, so joining the
/// [TextSegment.text]s gives [text] back. Timestamps past [maxSeconds] (the
/// video length; 0 or null means unknown) stay plain, so a comment cannot
/// link beyond the end of the video; so does `1:75:00` (minutes after an
/// hour must stay below 60).
List<TextSegment> splitTimestamps(String text, {int? maxSeconds}) {
  final segments = <TextSegment>[];
  var last = 0;
  for (final match in _timestamp.allMatches(text)) {
    final hours = match.group(1);
    final minutes = int.parse(match.group(2)!);
    if (hours != null && minutes > 59) continue;

    final seconds =
        (hours == null ? 0 : int.parse(hours)) * 3600 +
        minutes * 60 +
        int.parse(match.group(3)!);
    if (maxSeconds != null && maxSeconds > 0 && seconds > maxSeconds) continue;

    if (match.start > last) {
      segments.add(PlainSegment(text.substring(last, match.start)));
    }
    segments.add(TimestampSegment(match.group(0)!, seconds));
    last = match.end;
  }
  if (last < text.length) segments.add(PlainSegment(text.substring(last)));
  return segments;
}
