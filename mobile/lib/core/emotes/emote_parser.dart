import 'package:flutter/widgets.dart';

import 'emote_registry.dart';

final _emoteRegex = RegExp(r':([a-z0-9_]+):');

/// Parse message text and return a list of [InlineSpan] with emotes rendered
/// as larger emoji text. If the entire message is a single emote, it renders
/// at [singleEmoteSize] (sticker-style).
List<InlineSpan> parseEmotes(String text, TextStyle baseStyle) {
  final matches = _emoteRegex.allMatches(text).toList();
  if (matches.isEmpty) return [TextSpan(text: text, style: baseStyle)];

  // Check if the entire message is a single emote
  final isSingle =
      matches.length == 1 &&
      matches.first.start == 0 &&
      matches.first.end == text.length;

  final spans = <InlineSpan>[];
  var lastIndex = 0;

  for (final match in matches) {
    final code = match.group(1)!;
    final emote = emoteMap[code];

    if (emote == null) {
      // Not a valid emote, keep as text
      continue;
    }

    // Add preceding text
    if (match.start > lastIndex) {
      spans.add(TextSpan(text: text.substring(lastIndex, match.start), style: baseStyle));
    }

    // Add emote
    final emoteStyle = baseStyle.copyWith(
      fontSize: isSingle ? 32 : (baseStyle.fontSize ?? 14) + 4,
    );
    spans.add(
      WidgetSpan(
        alignment: PlaceholderAlignment.middle,
        child: Text(emote.emoji, style: emoteStyle),
      ),
    );

    lastIndex = match.end;
  }

  // Add remaining text
  if (lastIndex < text.length) {
    spans.add(TextSpan(text: text.substring(lastIndex), style: baseStyle));
  }

  return spans.isNotEmpty ? spans : [TextSpan(text: text, style: baseStyle)];
}
