import 'package:flutter/widgets.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:intl/date_symbol_data_local.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/features/messages/data/dm_models.dart';
import 'package:letslive_mobile/features/messages/domain/dm_event.dart';
import 'package:letslive_mobile/features/messages/presentation/conversation_display.dart';

import 'message_fakes.dart';

void main() {
  test('a message with a mistyped field is rejected loudly', () {
    final json = messageJson('a', 'hi')..['isDeleted'] = 'no';
    expect(() => DmMessage.fromJson(json), throwsFormatException);
    expect(
      () => DmMessage.fromJson(messageJson('a', 'hi', type: 'video')),
      throwsFormatException,
    );
  });

  test('push payloads parse; malformed ones are ignored', () {
    final event = parseDmEvent('dm:new_message', {
      'conversationId': 'c-1',
      'message': messageJson('a', 'hi'),
    });
    expect(event, isA<DmNewMessage>());
    expect(parseDmEvent('dm:new_message', {'conversationId': 'c-1'}), isNull);
    expect(parseDmEvent('dm:new_message', 'junk'), isNull);
    expect(parseDmEvent('dm:user_typing', {'conversationId': 'c'}), isNull);
    expect(parseDmEvent('dm:something_new', {'a': 1}), isNull);
    expect(
      parseDmEvent('dm:message_deleted', {
        'conversationId': 'c-1',
        'messageId': 'a',
      }),
      isA<DmMessageDeleted>(),
    );
  });

  test('a conversation update changes only the fields it names', () {
    final c = conversation('c-1', type: 'group', name: 'Old');
    final updated = c.withUpdate({
      'name': 'New',
      'avatarUrl': 'http://a/b.png',
    });
    expect(updated.name, 'New');
    expect(updated.avatarUrl, 'http://a/b.png');
    expect(updated.participants, hasLength(2));
    expect(() => c.withUpdate({'name': 5}), throwsFormatException);
  });

  test('conversation times: today, yesterday, weekday, date', () async {
    await initializeDateFormatting('en');
    final l10n = await AppLocalizations.delegate.load(const Locale('en'));
    final now = DateTime(2026, 11, 10, 15);
    // intl 0.20 puts a narrow no-break space before AM/PM.
    String at(DateTime d) =>
        formatConversationTime(d, now, 'en', l10n).replaceAll('\u202f', ' ');

    expect(at(DateTime(2026, 11, 10, 9, 5)), '9:05 AM');
    expect(at(DateTime(2026, 11, 9, 23, 30)), '11:30 PM Yesterday');
    expect(at(DateTime(2026, 11, 7, 12)), 'Sat');
    expect(at(DateTime(2026, 10, 1, 12)), 'Oct 1');
  });

  test(
    'names fall back when the other person or group name is missing',
    () async {
      final l10n = await AppLocalizations.delegate.load(const Locale('en'));
      final dm = conversation('c-1');
      expect(conversationDisplay(dm, me, l10n).name, 'bob');
      expect(conversationDisplay(dm, me, l10n).otherUserId, 'u-2');
      final group = conversation('c-2', type: 'group');
      expect(conversationDisplay(group, me, l10n).name, 'Group');
      expect(conversationDisplay(group, me, l10n).otherUserId, isNull);
    },
  );
}
