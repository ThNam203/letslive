import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/router/app_routes.dart';
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

/// The settings hub opened from the Profile tab (web's `/settings` tabs):
/// edit profile, security, the creator tools, and log out. Logout runs through the session,
/// which runs the logout hooks, calls the server and clears the cookies, and
/// the router then shows the login screen.
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
