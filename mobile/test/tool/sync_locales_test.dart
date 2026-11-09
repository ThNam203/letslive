import 'dart:convert';
import 'dart:io';

import 'package:flutter_test/flutter_test.dart';

import '../../tool/sync_locales.dart';

void main() {
  late Directory tmp;
  late String webDir;
  late String outDir;

  void writeLocale(String folder, Map<String, Object?> namespaces) {
    for (final entry in namespaces.entries) {
      final file = File('$webDir/$folder/${entry.key}.json')
        ..createSync(recursive: true);
      file.writeAsStringSync(jsonEncode(entry.value));
    }
  }

  void writeBoth(Map<String, Object?> namespaces) {
    writeLocale('en-US', namespaces);
    writeLocale('vi-VN', namespaces);
  }

  Map<String, Object?> readArb(String name) =>
      jsonDecode(File('$outDir/$name').readAsStringSync())
          as Map<String, Object?>;

  int sync([List<String> extra = const []]) {
    final sink = StringBuffer();
    return run(
      ['--web-dir=$webDir', '--out-dir=$outDir', ...extra],
      out: sink,
      err: sink,
    );
  }

  setUp(() {
    tmp = Directory.systemTemp.createTempSync('sync_locales_test');
    webDir = '${tmp.path}/web';
    outDir = '${tmp.path}/l10n';
  });

  tearDown(() => tmp.deleteSync(recursive: true));

  test('converts nested keys with namespace prefix', () {
    writeBoth({
      'auth': {
        'login': {'title': 'Log in'},
        'plain': 'Hello',
      },
      'api-response': {'some.key': 'Oops'},
    });

    expect(sync(), 0);

    final arb = readArb('app_en.arb');
    expect(arb['@@locale'], 'en');
    expect(arb['auth_login_title'], 'Log in');
    expect(arb['auth_plain'], 'Hello');
    expect(arb['api_response_some_key'], 'Oops');
    expect(readArb('app_vi.arb')['@@locale'], 'vi');
  });

  test('preserves interpolation placeholders', () {
    writeBoth({
      'common': {'greeting': 'Hi {{username}}, you have {{n}} items'},
    });

    expect(sync(), 0);

    final arb = readArb('app_en.arb');
    expect(arb['common_greeting'], 'Hi {username}, you have {n} items');
    final meta = arb['@common_greeting'] as Map<String, Object?>;
    expect((meta['placeholders'] as Map).keys, containsAll(['username', 'n']));
  });

  test('escapes apostrophes and literal braces for ICU', () {
    writeBoth({
      'common': {
        'brand': "Let's Live",
        'share': "Watch {{username}}'s video on Let's Live",
        'braces': 'Use { and } literally',
      },
    });

    expect(sync(), 0);

    final arb = readArb('app_en.arb');
    expect(arb['common_brand'], "Let''s Live");
    expect(arb['common_share'], "Watch {username}''s video on Let''s Live");
    expect(arb['common_braces'], "Use '{' and '}' literally");
  });

  test('converts i18next plural families into one ICU plural message', () {
    writeLocale('en-US', {
      'common': {
        'views_one': '{{count}} view',
        'views_other': '{{count}} views',
        'typing_one': '{{name}} is typing',
        'typing_two': '{{a}} and {{b}} are typing',
      },
    });
    writeLocale('vi-VN', {
      'common': {
        'views_other': '{{count}} lượt xem',
        'typing_one': '{{name}} đang nhập',
        'typing_two': '{{a}} và {{b}} đang nhập',
      },
    });

    expect(sync(), 0);

    final en = readArb('app_en.arb');
    expect(
      en['common_views'],
      '{count, plural, one{{count} view} other{{count} views}}',
    );
    expect(readArb('app_vi.arb')['common_views'], contains('other{{count}'));
    // Suffix without an `_other` sibling is an ordinary key, not a plural.
    expect(en['common_typing_one'], '{name} is typing');
    expect(en['common_typing_two'], '{a} and {b} are typing');
    expect(en.containsKey('common_typing'), isFalse);
  });

  test('plural family shadows a plain key of the same name', () {
    writeBoth({
      'common': {
        'replies': 'View {{count}} replies',
        'replies_one': 'View {{count}} reply',
        'replies_other': 'View {{count}} replies',
      },
    });

    expect(sync(), 0);

    expect(
      readArb('app_en.arb')['common_replies'],
      startsWith('{count, plural'),
    );
  });

  test('fails when a plain key differs from its plural family', () {
    writeLocale('en-US', {
      'common': {
        'replies': 'Replies',
        'replies_one': 'View {{count}} reply',
        'replies_other': 'View {{count}} replies',
      },
    });
    writeLocale('vi-VN', {
      'common': {'replies_other': 'Xem {{count}} trả lời'},
    });
    final err = StringBuffer();

    final code = run(['--web-dir=$webDir', '--out-dir=$outDir'], err: err);

    expect(code, 1);
    expect(err.toString(), contains('common:replies'));
    expect(File('$outDir/app_en.arb').existsSync(), isFalse);
    expect(File('$outDir/app_vi.arb').existsSync(), isFalse);
  });

  test('fails loudly when sanitized keys collide', () {
    writeBoth({
      'common': {'a-b': 'one', 'a.b': 'two'},
    });

    expect(sync(), isNot(0));
    expect(File('$outDir/app_en.arb').existsSync(), isFalse);
  });

  test('rejects unsupported interpolation syntax', () {
    writeBoth({
      'common': {'price': '{{amount, number}}'},
    });

    expect(sync(), isNot(0));
  });

  test('fails when a translation has a key the template lacks', () {
    writeLocale('en-US', {
      'common': {'a': 'A'},
    });
    writeLocale('vi-VN', {
      'common': {'a': 'A', 'extra': 'X'},
    });

    expect(sync(), isNot(0));
  });

  test('--check passes right after a sync', () {
    writeBoth({
      'common': {'a': 'A'},
    });
    expect(sync(), 0);

    expect(sync(['--check']), 0);
  });

  test('--check fails when a web key is missing in arb', () {
    writeBoth({
      'common': {'a': 'A'},
    });
    expect(sync(), 0);
    writeBoth({
      'common': {'a': 'A', 'b': 'B'},
    });

    final err = StringBuffer();
    final code = run([
      '--web-dir=$webDir',
      '--out-dir=$outDir',
      '--check',
    ], err: err);

    expect(code, 1);
    expect(err.toString(), contains('missing key common_b'));
  });

  test('--check fails when arb value or a stale key differs', () {
    writeBoth({
      'common': {'a': 'A', 'b': 'B'},
    });
    expect(sync(), 0);
    writeBoth({
      'common': {'a': 'A2'},
    });

    final err = StringBuffer();
    final code = run([
      '--web-dir=$webDir',
      '--out-dir=$outDir',
      '--check',
    ], err: err);

    expect(code, 1);
    expect(err.toString(), contains('value differs for common_a'));
    expect(err.toString(), contains('stale key common_b'));
  });

  test('--check fails when the arb files do not exist', () {
    writeBoth({
      'common': {'a': 'A'},
    });

    expect(sync(['--check']), 1);
  });
}
