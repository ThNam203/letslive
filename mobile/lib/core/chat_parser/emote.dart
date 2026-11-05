/// The chat emote set and shortcode expansion, ported from
/// `web/constant/emotes.ts` and `web/utils/chat-parser/emote.tsx`.
///
/// Web draws each emote as a Twemoji image of a Unicode code point; mobile
/// uses the same code points as native emoji text, so no image assets are
/// bundled.
library;

/// Picker tab an [Emote] belongs to (web `EMOTE_CATEGORIES`, same order).
enum EmoteCategory { smileys, gestures, hype, misc }

/// One emote: typed as `:code:`, shown as the emoji at [codepoint].
class Emote {
  const Emote({
    required this.code,
    required this.codepoint,
    required this.name,
    required this.category,
  });

  /// The shortcode between the colons, e.g. `smile`.
  final String code;

  /// Unicode code point(s) in hex, joined by `-` (e.g. `1f579-fe0f`).
  final String codepoint;

  /// English display name; the picker search matches it too.
  final String name;
  final EmoteCategory category;

  /// The emoji character(s) for [codepoint].
  String get emoji => String.fromCharCodes(
    codepoint.split('-').map((hex) => int.parse(hex, radix: 16)),
  );

  /// The text inserted into a message: `:code:`.
  String get shortcode => ':$code:';
}

/// Every emote, in web's order.
const List<Emote> emotes = [
  Emote(
    code: 'smile',
    codepoint: '1f60a',
    name: 'Smile',
    category: EmoteCategory.smileys,
  ),
  Emote(
    code: 'laugh',
    codepoint: '1f602',
    name: 'Laugh',
    category: EmoteCategory.smileys,
  ),
  Emote(
    code: 'love',
    codepoint: '1f60d',
    name: 'Love',
    category: EmoteCategory.smileys,
  ),
  Emote(
    code: 'cool',
    codepoint: '1f60e',
    name: 'Cool',
    category: EmoteCategory.smileys,
  ),
  Emote(
    code: 'thinking',
    codepoint: '1f914',
    name: 'Thinking',
    category: EmoteCategory.smileys,
  ),
  Emote(
    code: 'grin',
    codepoint: '1f601',
    name: 'Grin',
    category: EmoteCategory.smileys,
  ),
  Emote(
    code: 'joy',
    codepoint: '1f923',
    name: 'ROFL',
    category: EmoteCategory.smileys,
  ),
  Emote(
    code: 'sweat',
    codepoint: '1f605',
    name: 'Sweat Smile',
    category: EmoteCategory.smileys,
  ),
  Emote(
    code: 'plead',
    codepoint: '1f97a',
    name: 'Pleading',
    category: EmoteCategory.smileys,
  ),
  Emote(
    code: 'mindblown',
    codepoint: '1f92f',
    name: 'Mind Blown',
    category: EmoteCategory.smileys,
  ),
  Emote(
    code: 'melting',
    codepoint: '1fae0',
    name: 'Melting',
    category: EmoteCategory.smileys,
  ),
  Emote(
    code: 'partyface',
    codepoint: '1f973',
    name: 'Party Face',
    category: EmoteCategory.smileys,
  ),
  Emote(
    code: 'monocle',
    codepoint: '1f9d0',
    name: 'Monocle',
    category: EmoteCategory.smileys,
  ),
  Emote(
    code: 'cry',
    codepoint: '1f622',
    name: 'Cry',
    category: EmoteCategory.smileys,
  ),
  Emote(
    code: 'angry',
    codepoint: '1f621',
    name: 'Angry',
    category: EmoteCategory.smileys,
  ),
  Emote(
    code: 'wink',
    codepoint: '1f609',
    name: 'Wink',
    category: EmoteCategory.smileys,
  ),
  Emote(
    code: 'sus',
    codepoint: '1f928',
    name: 'Sus',
    category: EmoteCategory.smileys,
  ),
  Emote(
    code: 'skull',
    codepoint: '1f480',
    name: 'Skull',
    category: EmoteCategory.smileys,
  ),
  Emote(
    code: 'thumbsup',
    codepoint: '1f44d',
    name: 'Thumbs Up',
    category: EmoteCategory.gestures,
  ),
  Emote(
    code: 'thumbsdown',
    codepoint: '1f44e',
    name: 'Thumbs Down',
    category: EmoteCategory.gestures,
  ),
  Emote(
    code: 'clap',
    codepoint: '1f44f',
    name: 'Clap',
    category: EmoteCategory.gestures,
  ),
  Emote(
    code: 'wave',
    codepoint: '1f44b',
    name: 'Wave',
    category: EmoteCategory.gestures,
  ),
  Emote(
    code: 'ok',
    codepoint: '1f44c',
    name: 'OK',
    category: EmoteCategory.gestures,
  ),
  Emote(
    code: 'peace',
    codepoint: '270c-fe0f',
    name: 'Peace',
    category: EmoteCategory.gestures,
  ),
  Emote(
    code: 'fist',
    codepoint: '270a',
    name: 'Fist',
    category: EmoteCategory.gestures,
  ),
  Emote(
    code: 'muscle',
    codepoint: '1f4aa',
    name: 'Flex',
    category: EmoteCategory.gestures,
  ),
  Emote(
    code: 'pointup',
    codepoint: '1f446',
    name: 'Point Up',
    category: EmoteCategory.gestures,
  ),
  Emote(
    code: 'pray',
    codepoint: '1f64f',
    name: 'Pray',
    category: EmoteCategory.gestures,
  ),
  Emote(
    code: 'salute',
    codepoint: '1fae1',
    name: 'Salute',
    category: EmoteCategory.gestures,
  ),
  Emote(
    code: 'fire',
    codepoint: '1f525',
    name: 'Fire',
    category: EmoteCategory.hype,
  ),
  Emote(
    code: 'heart',
    codepoint: '2764-fe0f',
    name: 'Heart',
    category: EmoteCategory.hype,
  ),
  Emote(
    code: 'star',
    codepoint: '2b50',
    name: 'Star',
    category: EmoteCategory.hype,
  ),
  Emote(
    code: 'hype',
    codepoint: '1f389',
    name: 'Hype',
    category: EmoteCategory.hype,
  ),
  Emote(
    code: 'gg',
    codepoint: '1f3c6',
    name: 'GG',
    category: EmoteCategory.hype,
  ),
  Emote(
    code: 'sparkles',
    codepoint: '2728',
    name: 'Sparkles',
    category: EmoteCategory.hype,
  ),
  Emote(
    code: 'boom',
    codepoint: '1f4a5',
    name: 'Boom',
    category: EmoteCategory.hype,
  ),
  Emote(
    code: 'zap',
    codepoint: '26a1',
    name: 'Zap',
    category: EmoteCategory.hype,
  ),
  Emote(
    code: 'target',
    codepoint: '1f3af',
    name: 'Target',
    category: EmoteCategory.hype,
  ),
  Emote(
    code: 'medal',
    codepoint: '1f3c5',
    name: 'Medal',
    category: EmoteCategory.hype,
  ),
  Emote(
    code: 'tada',
    codepoint: '1f38a',
    name: 'Tada',
    category: EmoteCategory.hype,
  ),
  Emote(
    code: 'pog',
    codepoint: '1f62e',
    name: 'Pog',
    category: EmoteCategory.hype,
  ),
  Emote(
    code: 'ez',
    codepoint: '1f60f',
    name: 'EZ',
    category: EmoteCategory.hype,
  ),
  Emote(
    code: 'goat',
    codepoint: '1f410',
    name: 'GOAT',
    category: EmoteCategory.hype,
  ),
  Emote(
    code: 'eyes',
    codepoint: '1f440',
    name: 'Eyes',
    category: EmoteCategory.misc,
  ),
  Emote(
    code: '100',
    codepoint: '1f4af',
    name: '100',
    category: EmoteCategory.misc,
  ),
  Emote(
    code: 'money',
    codepoint: '1f4b0',
    name: 'Money',
    category: EmoteCategory.misc,
  ),
  Emote(
    code: 'ghost',
    codepoint: '1f47b',
    name: 'Ghost',
    category: EmoteCategory.misc,
  ),
  Emote(
    code: 'rocket',
    codepoint: '1f680',
    name: 'Rocket',
    category: EmoteCategory.misc,
  ),
  Emote(
    code: 'crown',
    codepoint: '1f451',
    name: 'Crown',
    category: EmoteCategory.misc,
  ),
  Emote(
    code: 'gem',
    codepoint: '1f48e',
    name: 'Gem',
    category: EmoteCategory.misc,
  ),
  Emote(
    code: 'cookie',
    codepoint: '1f36a',
    name: 'Cookie',
    category: EmoteCategory.misc,
  ),
  Emote(
    code: 'coffee',
    codepoint: '2615',
    name: 'Coffee',
    category: EmoteCategory.misc,
  ),
  Emote(
    code: 'pizza',
    codepoint: '1f355',
    name: 'Pizza',
    category: EmoteCategory.misc,
  ),
  Emote(
    code: 'joystick',
    codepoint: '1f579-fe0f',
    name: 'Joystick',
    category: EmoteCategory.misc,
  ),
  Emote(
    code: 'keyboard',
    codepoint: '2328-fe0f',
    name: 'Keyboard',
    category: EmoteCategory.misc,
  ),
  Emote(
    code: 'headphones',
    codepoint: '1f3a7',
    name: 'Headphones',
    category: EmoteCategory.misc,
  ),
  Emote(
    code: 'moon',
    codepoint: '1f319',
    name: 'Moon',
    category: EmoteCategory.misc,
  ),
];

final Map<String, Emote> _emotesByCode = {for (final e in emotes) e.code: e};

/// The emote with shortcode [code] (without colons), or null.
Emote? emoteByCode(String code) => _emotesByCode[code];

/// Emotes whose code or name contains [query] (case-insensitive); all of
/// them for an empty query (web `EmotePicker` filter).
List<Emote> searchEmotes(String query) {
  if (query.isEmpty) return emotes;
  final q = query.toLowerCase();
  return [
    for (final e in emotes)
      if (e.code.contains(q) || e.name.toLowerCase().contains(q)) e,
  ];
}

final RegExp _emotePattern = RegExp(r':([a-z0-9_]+):');

/// Replaces known `:shortcode:` with its emoji; unknown ones stay as typed.
/// Scans like web's global regex, so `:a:b:` consumes the same colons.
String expandEmotes(String text) => text.replaceAllMapped(
  _emotePattern,
  (m) => emoteByCode(m[1]!)?.emoji ?? m[0]!,
);

/// An unfinished shortcode at the end of the text (`:smi`, only at the start
/// or after whitespace), which the next emote picked completes (web
/// `activeEmotePattern`).
final RegExp _activeEmotePattern = RegExp(
  r'(^|\s):([a-z0-9_]*)$',
  caseSensitive: false,
);

/// Inserts [emote] into [text] at [cursor] and returns the new text and the
/// cursor after it. An unfinished shortcode right before the cursor is
/// completed instead of left behind; a space follows the emote then, as on
/// web. A [cursor] outside the text appends.
({String text, int cursor}) insertEmote(String text, int cursor, Emote emote) {
  final at = cursor < 0 || cursor > text.length ? text.length : cursor;
  final before = text.substring(0, at);
  final after = text.substring(at);
  final active = _activeEmotePattern.firstMatch(before);
  final start = active == null ? at : active.start + active.group(1)!.length;
  final inserted = active == null ? emote.shortcode : '${emote.shortcode} ';
  return (
    text: '${text.substring(0, start)}$inserted$after',
    cursor: start + inserted.length,
  );
}
