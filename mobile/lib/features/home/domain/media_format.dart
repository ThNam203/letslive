import 'package:intl/intl.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';

/// `3725` -> `1:02:05`, `65` -> `1:05` (web `formatSeconds`).
String formatDuration(int totalSeconds) {
  final hours = totalSeconds ~/ 3600;
  final minutes = (totalSeconds % 3600) ~/ 60;
  final seconds = totalSeconds % 60;
  String two(int n) => n.toString().padLeft(2, '0');
  return hours > 0
      ? '$hours:${two(minutes)}:${two(seconds)}'
      : '$minutes:${two(seconds)}';
}

/// Relative time like "5 minutes ago" (web `dateDiffFromNow`). A time in the
/// future (clock skew) counts as "0 seconds ago".
String timeAgo(AppLocalizations l10n, DateTime past, DateTime now) {
  final seconds = now.difference(past).inSeconds.clamp(0, 1 << 40);
  if (seconds < 60) return l10n.common_time_seconds_ago(seconds);
  final minutes = seconds ~/ 60;
  if (minutes < 60) return l10n.common_time_minutes_ago(minutes);
  final hours = minutes ~/ 60;
  if (hours < 24) return l10n.common_time_hours_ago(hours);
  final days = hours ~/ 24;
  if (days < 7) return l10n.common_time_days_ago(days);
  if (days < 30) return l10n.common_time_weeks_ago(days ~/ 7);
  if (days < 365) return l10n.common_time_months_ago(days ~/ 30);
  return l10n.common_time_years_ago(days ~/ 365);
}

/// "1.2K views" in the active language (web `common:vod.views_compact`).
String viewsLabel(AppLocalizations l10n, int count) {
  final formatted = NumberFormat.compact(locale: l10n.localeName).format(count);
  return l10n.common_vod_views_compact(formatted, count);
}

/// Where the transcoder stores a frame of a stream or video that has no
/// uploaded thumbnail (web `mediaThumbnailUrl`).
String mediaThumbnailUrl(String apiBaseUrl, String id, String? thumbnailUrl) =>
    thumbnailUrl ?? '$apiBaseUrl/files/livestreams/$id/thumbnail.jpeg';
