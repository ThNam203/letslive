/// Finds which parts of a text match a search query, ignoring case and
/// Vietnamese accents ("tro choi" matches "Trò chơi"). A port of
/// `web/utils/highlight.ts`, so both apps highlight the same way.
///
/// Exact substrings win; a query word of 4+ letters with no exact hit
/// highlights the words it is a likely typo of, like the backend's fuzzy
/// search. Ranges are UTF-16 offsets into the original text.
library;

/// A half-open range `[start, end)` of the original text.
typedef TextRange = ({int start, int end});

/// A slice of a text and whether it matched.
typedef TextSegment = ({String text, bool isMatch});

/// Words shorter than this are matched exactly only (no typo tolerance).
const int _minFuzzyTokenLength = 4;

final RegExp _combiningMarks = RegExp(r'\p{M}', unicode: true);
final RegExp _word = RegExp(r'[\p{L}\p{N}]+', unicode: true);

/// Precomposed Vietnamese letters to their base letter. Dart has no Unicode
/// normalisation (NFD) in the SDK, so this table stands in for it; combining
/// marks (decomposed input) are stripped by [_combiningMarks].
final Map<String, String> _baseLetter = {
  for (final entry in const {
    'a': 'àáạảãâầấậẩẫăằắặẳẵ',
    'e': 'èéẹẻẽêềếệểễ',
    'i': 'ìíịỉĩ',
    'o': 'òóọỏõôồốộổỗơờớợởỡ',
    'u': 'ùúụủũưừứựửữ',
    'y': 'ỳýỵỷỹ',
    'd': 'đ',
  }.entries)
    for (final letter in entry.value.runes)
      String.fromCharCode(letter): entry.key,
};

/// One character without accent or case.
String _foldChar(String ch) {
  final lower = ch.toLowerCase().replaceAll(_combiningMarks, '');
  return _baseLetter[lower] ?? lower;
}

/// [text] folded, plus for every folded UTF-16 unit the offset in [text] of
/// the character it came from (folding can change the length).
({String folded, List<int> origin}) _foldWithMap(String text) {
  final folded = StringBuffer();
  final origin = <int>[];
  var index = 0;
  for (final rune in text.runes) {
    final ch = String.fromCharCode(rune);
    final f = _foldChar(ch);
    for (var i = 0; i < f.length; i++) {
      origin.add(index);
    }
    folded.write(f);
    index += ch.length;
  }
  return (folded: folded.toString(), origin: origin);
}

int _maxEdits(int length) => length <= 5
    ? 1
    : length <= 9
    ? 2
    : 3;

/// Edit distance of [a] and [b] counting a swap of neighbours as one typo;
/// `limit + 1` when it is certainly above [limit].
int _editDistance(String a, String b, int limit) {
  if ((a.length - b.length).abs() > limit) return limit + 1;
  var before = <int>[];
  var prev = List<int>.generate(b.length + 1, (j) => j);
  for (var i = 1; i <= a.length; i++) {
    final curr = <int>[i];
    var rowMin = i;
    for (var j = 1; j <= b.length; j++) {
      final cost = a[i - 1] == b[j - 1] ? 0 : 1;
      var best = [
        prev[j] + 1,
        curr[j - 1] + 1,
        prev[j - 1] + cost,
      ].reduce((x, y) => x < y ? x : y);
      if (i > 1 && j > 1 && a[i - 1] == b[j - 2] && a[i - 2] == b[j - 1]) {
        final swapped = before[j - 2] + 1;
        if (swapped < best) best = swapped;
      }
      curr.add(best);
      if (best < rowMin) rowMin = best;
    }
    if (rowMin > limit) return limit + 1;
    before = prev;
    prev = curr;
  }
  return prev[b.length];
}

bool _isFuzzyWordMatch(String token, String word) {
  final limit = _maxEdits(token.length);
  if (_editDistance(token, word, limit) <= limit) return true;
  // "jan" should light up "jane_streams": compare against the word's prefix.
  return word.length > token.length &&
      _editDistance(token, word.substring(0, token.length), limit) <= limit;
}

List<TextRange> _substringRanges(String haystack, String needle) {
  final ranges = <TextRange>[];
  if (needle.isEmpty) return ranges;
  var from = 0;
  while (from <= haystack.length - needle.length) {
    final at = haystack.indexOf(needle, from);
    if (at == -1) break;
    ranges.add((start: at, end: at + needle.length));
    from = at + needle.length;
  }
  return ranges;
}

List<TextRange> _mergeRanges(List<TextRange> ranges) {
  final sorted = [...ranges]..sort((a, b) => a.start - b.start);
  final merged = <TextRange>[];
  for (final range in sorted) {
    if (merged.isNotEmpty && range.start <= merged.last.end) {
      final last = merged.removeLast();
      merged.add((
        start: last.start,
        end: range.end > last.end ? range.end : last.end,
      ));
    } else {
      merged.add(range);
    }
  }
  return merged;
}

/// The parts of [text] that match [query], merged and in order. Empty for a
/// blank query or text.
List<TextRange> findHighlightRanges(String text, String query) {
  final tokens = _foldWithMap(query).folded
      .split(RegExp(r'\s+'))
      .where((t) => t.isNotEmpty)
      .toList();
  if (text.isEmpty || tokens.isEmpty) return const [];

  final (:folded, :origin) = _foldWithMap(text);
  final words = [
    for (final m in _word.allMatches(folded))
      (word: m.group(0)!, start: m.start, end: m.end),
  ];

  final foldedRanges = <TextRange>[
    ..._substringRanges(folded, tokens.join(' ')),
    for (final token in tokens) ..._tokenRanges(folded, words, token),
  ];

  return _mergeRanges([
    for (final range in foldedRanges)
      (
        start: origin[range.start],
        // The end covers the whole last character (an emoji is two units).
        end: origin[range.end - 1] + _charLengthAt(text, origin[range.end - 1]),
      ),
  ]);
}

List<TextRange> _tokenRanges(
  String folded,
  List<({String word, int start, int end})> words,
  String token,
) {
  if (token.length == 1) {
    return [
      for (final w in words)
        if (w.word.startsWith(token)) (start: w.start, end: w.start + 1),
    ];
  }
  final exact = _substringRanges(folded, token);
  if (exact.isNotEmpty || token.length < _minFuzzyTokenLength) return exact;
  return [
    for (final w in words)
      if (_isFuzzyWordMatch(token, w.word)) (start: w.start, end: w.end),
  ];
}

/// Length in UTF-16 units of the character starting at [index] (2 for an
/// emoji and other characters outside the basic plane).
int _charLengthAt(String text, int index) {
  final unit = text.codeUnitAt(index);
  return unit >= 0xD800 && unit <= 0xDBFF ? 2 : 1;
}

/// Cuts [text] at the [ranges] (from [findHighlightRanges]) into matched and
/// plain slices that together make up the whole text.
List<TextSegment> splitByRanges(String text, List<TextRange> ranges) {
  final segments = <TextSegment>[];
  var cursor = 0;
  for (final range in ranges) {
    if (range.start > cursor) {
      segments.add((text: text.substring(cursor, range.start), isMatch: false));
    }
    segments.add((text: text.substring(range.start, range.end), isMatch: true));
    cursor = range.end;
  }
  if (cursor < text.length) {
    segments.add((text: text.substring(cursor), isMatch: false));
  }
  return segments;
}

/// A short piece of [text] around the first match of [query], with "…" where
/// the text was cut; null when nothing matches. Cuts at spaces when it can.
String? excerptAroundMatch(String text, String query, {int radius = 60}) {
  final ranges = findHighlightRanges(text, query);
  if (ranges.isEmpty) return null;
  final first = ranges.first;

  var start = first.start - radius < 0 ? 0 : first.start - radius;
  var end = first.end + radius > text.length ? text.length : first.end + radius;
  if (start > 0) {
    final space = text.indexOf(' ', start);
    start = space == -1 || space >= first.start ? start : space + 1;
  }
  if (end < text.length) {
    final space = text.lastIndexOf(' ', end);
    end = space <= first.end ? end : space;
  }
  final body = text
      .substring(start, end)
      .replaceAll(RegExp(r'\s+'), ' ')
      .trim();
  return '${start > 0 ? '…' : ''}$body${end < text.length ? '…' : ''}';
}
