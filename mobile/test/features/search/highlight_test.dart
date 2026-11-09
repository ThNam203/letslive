import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/features/search/domain/highlight.dart';

/// The matched parts of [text] for [query], as strings.
List<String> matches(String text, String query) => [
  for (final range in findHighlightRanges(text, query))
    text.substring(range.start, range.end),
];

void main() {
  group('findHighlightRanges', () {
    test('accent-less query highlights the accented text', () {
      expect(matches('Trò chơi vui', 'tro choi'), ['Trò chơi']);
    });

    test('ignores case and đ', () {
      expect(matches('Đường lên ĐỈNH', 'duong dinh'), ['Đường', 'ĐỈNH']);
    });

    test('a decomposed (combining mark) text matches too', () {
      // "ò" written as o + U+0300; like web, the mark itself is not marked.
      const text = 'Tro\u0300 choi';
      expect(matches(text, 'tro choi'), [text]);
    });

    test('a query word with a typo lights up the word it likely means', () {
      expect(matches('hello jane streams', 'jnae'), ['jane']);
    });

    test('a short word has no typo tolerance', () {
      expect(matches('cat bat', 'at'), ['at', 'at']);
      expect(findHighlightRanges('cat', 'cot'), isEmpty);
    });

    test('no match or a blank query highlights nothing', () {
      expect(findHighlightRanges('hello', 'xyz'), isEmpty);
      expect(findHighlightRanges('hello', '   '), isEmpty);
      expect(findHighlightRanges('', 'a'), isEmpty);
    });

    test('an emoji stays whole', () {
      expect(matches('party 🎉 time', 'party'), ['party']);
      expect(matches('🎉party', 'party'), ['party']);
    });
  });

  group('splitByRanges', () {
    test('cuts the text into matched and plain parts', () {
      const text = 'Trò chơi vui';
      final parts = splitByRanges(text, findHighlightRanges(text, 'choi'));
      expect(
        [for (final p in parts) (p.text, p.isMatch)],
        [('Trò ', false), ('chơi', true), (' vui', false)],
      );
    });
  });

  group('excerptAroundMatch', () {
    test('is null without a match', () {
      expect(excerptAroundMatch('hello world', 'zzz'), isNull);
    });

    test('never cuts an emoji in half', () {
      // No spaces near the cut, so the cut lands inside the emoji run.
      final text = '${'🎉' * 40}needle${'🎉' * 40}';
      final excerpt = excerptAroundMatch(text, 'needle', radius: 5)!;
      final units = excerpt.codeUnits;
      for (var i = 0; i < units.length; i++) {
        final isHigh = units[i] >= 0xD800 && units[i] <= 0xDBFF;
        final isLow = units[i] >= 0xDC00 && units[i] <= 0xDFFF;
        if (isHigh) expect(units[i + 1] >= 0xDC00, isTrue, reason: 'lone high');
        if (isLow) expect(units[i - 1] <= 0xDBFF, isTrue, reason: 'lone low');
      }
      expect(excerpt, contains('needle'));
    });

    test('keeps a short text whole', () {
      expect(excerptAroundMatch('hello world', 'world'), 'hello world');
    });

    test('shortens a long text around the first match with ellipses', () {
      final text = '${'lorem ipsum ' * 20}needle${' dolor sit' * 20}';
      final excerpt = excerptAroundMatch(text, 'needle')!;
      expect(excerpt, contains('needle'));
      expect(excerpt.startsWith('…'), isTrue);
      expect(excerpt.endsWith('…'), isTrue);
      expect(excerpt.length, lessThan(text.length ~/ 2));
    });
  });
}
