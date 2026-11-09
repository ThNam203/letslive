import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations_en.dart';
import 'package:letslive_mobile/features/home/domain/media_format.dart';

void main() {
  final l10n = AppLocalizationsEn();
  final now = DateTime.utc(2026, 11, 4, 12);

  test('formatDuration shows minutes:seconds, with hours when needed', () {
    expect(formatDuration(0), '0:00');
    expect(formatDuration(65), '1:05');
    expect(formatDuration(3725), '1:02:05');
  });

  test('timeAgo picks the largest unit like web', () {
    String ago(Duration d) => timeAgo(l10n, now.subtract(d), now);
    expect(ago(const Duration(seconds: 30)), '30 seconds ago');
    expect(ago(const Duration(minutes: 1)), '1 minute ago');
    expect(ago(const Duration(hours: 5)), '5 hours ago');
    expect(ago(const Duration(days: 3)), '3 days ago');
    expect(ago(const Duration(days: 14)), '2 weeks ago');
    expect(ago(const Duration(days: 90)), '3 months ago');
    expect(ago(const Duration(days: 800)), '2 years ago');
  });

  test('timeAgo treats a future time as just now', () {
    expect(
      timeAgo(l10n, now.add(const Duration(minutes: 5)), now),
      '0 seconds ago',
    );
  });

  test('viewsLabel compacts large counts', () {
    expect(viewsLabel(l10n, 1500), '1.5K views');
    expect(viewsLabel(l10n, 1), '1 view');
  });

  test('mediaThumbnailUrl falls back to the transcoder frame', () {
    expect(
      mediaThumbnailUrl('http://h:8000', 'id-1', null),
      'http://h:8000/files/livestreams/id-1/thumbnail.jpeg',
    );
    expect(
      mediaThumbnailUrl('http://h:8000', 'id-1', 'http://x/t.png'),
      'http://x/t.png',
    );
  });
}
