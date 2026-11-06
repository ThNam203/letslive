import 'dart:math';

import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/chat_parser/chat_command.dart';
import 'package:letslive_mobile/core/chat_parser/command.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations_en.dart';

/// Always draws the first value, so `/roll N` rolls 1.
class _FirstRandom implements Random {
  @override
  bool nextBool() => false;
  @override
  double nextDouble() => 0;
  @override
  int nextInt(int max) => 0;
}

ChatCommand _custom(
  String id,
  ChatCommandScope scope,
  String name,
  String response, {
  String description = '',
}) => ChatCommand(
  id: id,
  scope: scope,
  ownerId: 'o',
  name: name,
  response: response,
  description: description,
);

ParsedInput _parse(
  String input, {
  List<ChatCommand> commands = const [],
  String? pickedId,
}) => parseChatInput(
  input,
  commands: commands,
  pickedId: pickedId,
  random: _FirstRandom(),
);

String? _sent(ParsedInput p) => p is ParsedSend ? p.text : null;

void main() {
  // Expected values are web's output (utils/chat-parser/command.ts).
  group('parser expands built-in commands exactly as web', () {
    const cases = <String, String>{
      '/shrug': r'¯\_(ツ)_/¯',
      '/shrug hello there': r'hello there ¯\_(ツ)_/¯',
      '/shrug   padded  ': r'padded ¯\_(ツ)_/¯',
      '/SHRUG Caps': r'Caps ¯\_(ツ)_/¯',
      '/tableflip': '(╯°□°)╯︵ ┻━┻',
      '/tableflip rage': 'rage (╯°□°)╯︵ ┻━┻',
      '/unflip': '┬─┬ ノ( ゜-゜ノ)',
      '/unflip calm down': 'calm down ┬─┬ ノ( ゜-゜ノ)',
      '/roll': '🎲 rolled 1 (1-100)',
      '/roll 6': '🎲 rolled 1 (1-6)',
      '/roll 1000000': '🎲 rolled 1 (1-1000000)',
      '/roll 12abc': '🎲 rolled 1 (1-12)',
      '/roll +7': '🎲 rolled 1 (1-7)',
      'hello world': 'hello world',
      '  spaced out  ': 'spaced out',
    };
    cases.forEach((input, expected) {
      test('"$input"', () => expect(_sent(_parse(input)), expected));
    });

    test('/roll uses the injected random source', () {
      final p = parseChatInput(
        '/roll 10',
        commands: const [],
        random: Random(7),
        rollMessage: (value, max) => '$value of $max',
      );
      expect(_sent(p), '${Random(7).nextInt(10) + 1} of 10');
    });

    for (final bad in ['/roll 0', '/roll -3', '/roll 1000001', '/roll abc']) {
      test('"$bad" is a usage error', () {
        final p = _parse(bad);
        expect(p, isA<ParsedError>());
        expect((p as ParsedError).error, ChatInputError.rollUsage);
        expect(
          chatInputErrorText(AppLocalizationsEn(), p),
          'Usage: /roll [1-1000000]',
        );
      });
    }

    test('/help asks for the local help text', () {
      expect(_parse('/help'), isA<ParsedHelp>());
      expect(_parse('/help me'), isA<ParsedHelp>());
    });

    test('an unknown command is a local error naming it', () {
      final p = _parse('/nope arg') as ParsedError;
      expect(p.error, ChatInputError.unknownCommand);
      expect(
        chatInputErrorText(AppLocalizationsEn(), p),
        'Unknown chat command: /nope',
      );
    });

    test('a lone slash or blank input does nothing', () {
      expect(_parse('/'), isA<ParsedNothing>());
      expect(_parse('   '), isA<ParsedNothing>());
    });

    test('an expanded command is cut to 500 units, keeping pairs whole', () {
      final long = '🎉' * 300;
      // "a" shifts the pairs so unit 500 would split one.
      final text = _sent(_parse('/shrug a$long'))!;
      expect(text.length, 499);
      expect(text.runes.every((r) => r == 0x1F389 || r == 0x61), isTrue);
    });
  });

  group('custom channel command resolved from commands list', () {
    final channel = _custom(
      'c1',
      ChatCommandScope.channel,
      'discord',
      'Join discord.gg/x',
    );
    final mine = _custom('u1', ChatCommandScope.user, 'discord', 'my discord');

    test('runs the command by name', () {
      expect(
        _sent(_parse('/discord', commands: [channel])),
        'Join discord.gg/x',
      );
    });

    test('channel beats user when both share a name', () {
      expect(
        _sent(_parse('/discord', commands: [mine, channel])),
        'Join discord.gg/x',
      );
    });

    test('the picked suggestion beats the default order', () {
      expect(
        _sent(_parse('/discord', commands: [mine, channel], pickedId: 'u1')),
        'my discord',
      );
    });

    test('a built-in beats a custom command of the same name', () {
      final shadow = _custom('c2', ChatCommandScope.channel, 'shrug', 'no');
      expect(_sent(_parse('/shrug', commands: [shadow])), r'¯\_(ツ)_/¯');
    });
  });

  group('suggestions', () {
    final l10n = AppLocalizationsEn();
    final channel = _custom(
      'c1',
      ChatCommandScope.channel,
      'discord',
      'Join us',
      description: 'Our server',
    );
    final index = buildChatCommandIndex([channel], l10n);

    List<String> names(String input) => [
      for (final s in filterChatCommandSuggestions(index, input)) s.name,
    ];

    test('"/" lists everything, built-ins first', () {
      expect(names('/'), [
        'shrug',
        'tableflip',
        'unflip',
        'roll',
        'help',
        'discord',
      ]);
    });

    test('filters by prefix as you type, case-insensitively', () {
      expect(names('/t'), ['tableflip']);
      expect(names('/DI'), ['discord']);
      expect(names('/zzz'), isEmpty);
    });

    test('no suggestions for plain text or once an argument starts', () {
      expect(names('hello'), isEmpty);
      expect(names('/roll 5'), isEmpty);
    });

    test('caps at 8 entries', () {
      final many = [
        for (var i = 0; i < 12; i++)
          _custom('x$i', ChatCommandScope.user, 'cmd$i', 'r'),
      ];
      final big = buildChatCommandIndex(many, l10n);
      expect(filterChatCommandSuggestions(big, '/cmd'), hasLength(8));
    });

    test('a custom command shows its description, else its response', () {
      final bare = _custom('c3', ChatCommandScope.user, 'bare', 'Hi!');
      final bareIndex = buildChatCommandIndex([bare, channel], l10n);
      expect(bareIndex.firstWhere((s) => s.name == 'bare').description, 'Hi!');
      expect(
        bareIndex.firstWhere((s) => s.name == 'discord').description,
        'Our server',
      );
      expect(index.last.source, ChatCommandSource.channel);
    });

    test('chatCommandName reads the typed name', () {
      expect(chatCommandName('/Roll 5'), 'roll');
      expect(chatCommandName('hi'), '');
    });

    test('help text lists built-ins then custom commands with source', () {
      final text = buildChatCommandHelpText([channel], l10n);
      expect(text, startsWith('Available chat commands:\n/shrug [text] — '));
      expect(text, contains('Custom chat commands:'));
      expect(text, endsWith('/discord — Our server (channel)'));
    });
  });
}
