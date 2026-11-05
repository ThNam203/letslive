import 'dart:async';
import 'dart:developer' as developer;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/l10n/locale_provider.dart';
import 'package:letslive_mobile/core/router/app_routes.dart';
import 'package:letslive_mobile/core/theme/theme_mode_provider.dart';
import 'package:letslive_mobile/features/auth/auth_providers.dart';

/// A creator-tools entry of the hub (web's stream, chat commands, VODs and
/// upload tabs).
typedef _CreatorTool = ({
  String key,
  IconData icon,
  String Function(AppLocalizations l10n) title,
  String route,
});

final List<_CreatorTool> _creatorTools = [
  (
    key: 'settings-stream',
    icon: Icons.sensors,
    title: (l10n) => l10n.settings_navigation_stream,
    route: AppRoutes.settingsStream,
  ),
  (
    key: 'settings-chat-commands',
    icon: Icons.chat_bubble_outline,
    title: (l10n) => l10n.settings_navigation_chat_commands,
    route: AppRoutes.settingsChatCommands,
  ),
  (
    key: 'settings-vods',
    icon: Icons.video_library_outlined,
    title: (l10n) => l10n.settings_navigation_vods,
    route: AppRoutes.settingsVods,
  ),
  (
    key: 'settings-upload',
    icon: Icons.upload_outlined,
    title: (l10n) => l10n.settings_navigation_upload,
    route: AppRoutes.settingsUpload,
  ),
];

/// Each app language with its name in that language and the BCP-47 tag web
/// saves on the profile (web `lib/i18n/settings.ts`,
/// `I18N_LANGUAGE_COUNTRY_MAP`). Names are not translated, as on web.
const List<({Locale locale, String name, String tag})> _languages = [
  (locale: Locale('en'), name: 'English', tag: 'en-US'),
  (locale: Locale('vi'), name: 'Tiếng Việt', tag: 'vi-VN'),
];

/// The settings hub opened from the Profile tab (web's `/settings` tabs):
/// edit profile, security, theme and language (web shows those two on
/// `/settings/profile`), the creator tools, and log out. Logout runs through
/// the session, which runs the logout hooks, calls the server and clears the
/// cookies, and the router then shows the login screen.
class SettingsScreen extends ConsumerWidget {
  const SettingsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    return Scaffold(
      appBar: AppBar(title: Text(l10n.settings_page_title)),
      body: ListView(
        children: [
          ListTile(
            key: const ValueKey('settings-profile'),
            leading: const Icon(Icons.person_outline),
            title: Text(l10n.settings_profile_title),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => context.push(AppRoutes.settingsProfile),
          ),
          ListTile(
            key: const ValueKey('settings-security'),
            leading: const Icon(Icons.lock_outline),
            title: Text(l10n.settings_navigation_security),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => context.push(AppRoutes.settingsSecurity),
          ),
          const _ThemeTile(),
          const _LanguageTile(),
          const Divider(),
          for (final tool in _creatorTools)
            ListTile(
              key: ValueKey(tool.key),
              leading: Icon(tool.icon),
              title: Text(tool.title(l10n)),
              trailing: const Icon(Icons.chevron_right),
              onTap: () => context.push(tool.route),
            ),
          const Divider(),
          ListTile(
            key: const ValueKey('settings-logout'),
            leading: const Icon(Icons.logout),
            title: Text(l10n.auth_logout),
            onTap: () => ref.read(sessionProvider.notifier).logout(),
          ),
        ],
      ),
    );
  }
}

/// Shows the options of a picker and answers the chosen value, or null when
/// dismissed. The current [selected] value is checked.
Future<T?> _pick<T>(
  BuildContext context, {
  required String title,
  required T selected,
  required List<(T, String)> options,
}) => showDialog<T>(
  context: context,
  builder: (context) => SimpleDialog(
    title: Text(title),
    children: [
      for (final (value, label) in options)
        ListTile(
          title: Text(label),
          trailing: value == selected ? const Icon(Icons.check) : null,
          onTap: () => Navigator.of(context).pop(value),
        ),
    ],
  ),
);

/// Light / dark / follow the system, like web's theme list. The choice is
/// stored on the device ([ThemeModeNotifier]).
class _ThemeTile extends ConsumerWidget {
  const _ThemeTile();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final mode = ref.watch(themeModeProvider);
    final options = [
      (ThemeMode.light, l10n.theme_light),
      (ThemeMode.dark, l10n.theme_dark),
      (ThemeMode.system, l10n.theme_system),
    ];
    return ListTile(
      key: const ValueKey('settings-theme'),
      leading: const Icon(Icons.palette_outlined),
      title: Text(l10n.settings_themes_title),
      subtitle: Text(options.firstWhere((o) => o.$1 == mode).$2),
      onTap: () async {
        final picked = await _pick(
          context,
          title: l10n.settings_themes_title,
          selected: mode,
          options: options,
        );
        if (picked != null) {
          await ref.read(themeModeProvider.notifier).setThemeMode(picked);
        }
      },
    );
  }
}

/// The app language. Like web it is stored on the device
/// ([LocaleNotifier]) and saved on the profile (`PUT /user/me {locale}`);
/// a failed profile save is only logged, the app language changed anyway.
class _LanguageTile extends ConsumerWidget {
  const _LanguageTile();

  Future<void> _change(WidgetRef ref, Locale locale, String tag) async {
    // Read up front: the whole app rebuilds in the new language meanwhile.
    final session = ref.read(sessionProvider.notifier);
    await ref.read(localeProvider.notifier).setLocale(locale);
    final res = await session.updateProfile(locale: tag);
    if (!res.success) {
      developer.log(
        'saving the profile locale failed: ${res.statusCode} ${res.key}',
        name: 'settings',
      );
    }
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final current = ref.watch(localeProvider);
    final language = _languages.firstWhere(
      (l) => l.locale.languageCode == current.languageCode,
      orElse: () => _languages.first,
    );
    return ListTile(
      key: const ValueKey('settings-language'),
      leading: const Icon(Icons.language),
      title: Text(l10n.settings_language_title),
      subtitle: Text(language.name),
      onTap: () async {
        final picked = await _pick(
          context,
          title: l10n.settings_language_title,
          selected: language,
          options: [for (final l in _languages) (l, l.name)],
        );
        if (picked != null && picked != language) {
          unawaited(_change(ref, picked.locale, picked.tag));
        }
      },
    );
  }
}
