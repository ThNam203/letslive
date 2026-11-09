import 'package:intl/intl.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';

/// "Joined Nov 4, 2026" in the active language (web
/// `users:profile.joined_prefix` + `formatLocaleDate`).
String joinedLabel(AppLocalizations l10n, DateTime createdAt) =>
    '${l10n.users_profile_joined_prefix} '
    '${DateFormat.yMMMd(l10n.localeName).format(createdAt.toLocal())}';

/// What a social link shows: host without `www.` plus the path, e.g.
/// `github.com/octocat` (web `toSocialLinks`). Null for anything that is not
/// an http(s) URL: links are user supplied, only those are opened.
String? socialLinkLabel(String url) {
  final uri = Uri.tryParse(url);
  if (uri == null || !uri.hasAuthority || uri.host.isEmpty) return null;
  if (uri.scheme != 'http' && uri.scheme != 'https') return null;
  final host = uri.host.replaceFirst(RegExp(r'^www\.'), '');
  final path = uri.path == '/' ? '' : uri.path;
  return '$host$path';
}
