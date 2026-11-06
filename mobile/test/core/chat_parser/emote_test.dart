import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/chat_parser/emote.dart';

void main() {
  test('the set matches web: 57 emotes with unique codes', () {
    expect(emotes, hasLength(57));
    expect({for (final e in emotes) e.code}, hasLength(57));
  });

  test('code points become emoji, including two-part ones', () {
    expect(emoteByCode('smile')!.emoji, '😊');
    expect(emoteByCode('heart')!.emoji, '❤️');
    expect(emoteByCode('joystick')!.emoji, '🕹️');
  });

  group('expandEmotes', () {
    test('replaces known shortcodes and keeps the rest', () {
      expect(
        expandEmotes('hi :smile: and :nope: :fire:'),
        'hi 😊 and :nope: 🔥',
      );
    });

    test('consumes colons like web: ":a:smile:" leaves the second pair', () {
      expect(expandEmotes(':a:smile:'), ':a:smile:');
    });

    test('keeps Vietnamese text and emoji untouched', () {
      expect(expandEmotes('Xin chào 🎉 :100:'), 'Xin chào 🎉 💯');
    });
  });

  test('searchEmotes matches code or name', () {
    expect(searchEmotes(''), emotes);
    expect(searchEmotes('THUMBS').map((e) => e.code), [
      'thumbsup',
      'thumbsdown',
    ]);
    expect(searchEmotes('rofl').single.code, 'joy');
  });

  group('insertEmote at the cursor', () {
    final smile = emoteByCode('smile')!;

    test('inserts in the middle of the text', () {
      final r = insertEmote('helloworld', 5, smile);
      expect(r.text, 'hello:smile:world');
      expect(r.cursor, 12);
    });

    test('appends when the cursor is unknown', () {
      expect(insertEmote('hi ', -1, smile).text, 'hi :smile:');
    });

    test('completes an unfinished shortcode before the cursor', () {
      final r = insertEmote('hi :sm', 6, smile);
      expect(r.text, 'hi :smile: ');
      expect(r.cursor, 11);
    });
  });
}
