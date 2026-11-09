import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';

/// How long ago [at] was, as web's `timeAgo` says it: "just now", then
/// minutes, hours and days.
String notificationTimeAgo(DateTime at, DateTime now, AppLocalizations l10n) {
  final diff = now.difference(at);
  if (diff.inSeconds < 60) return l10n.notification_time_just_now;
  if (diff.inMinutes < 60) return l10n.notification_time_m_ago(diff.inMinutes);
  if (diff.inHours < 24) return l10n.notification_time_h_ago(diff.inHours);
  return l10n.notification_time_d_ago(diff.inDays);
}
