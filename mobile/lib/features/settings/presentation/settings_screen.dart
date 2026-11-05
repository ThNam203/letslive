import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/router/app_routes.dart';
import 'package:letslive_mobile/features/auth/auth_providers.dart';

/// The settings hub opened from the Profile tab (web's `/settings` tabs):
/// edit profile, security, and log out. Logout runs through the session,
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
