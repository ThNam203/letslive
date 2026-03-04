import 'package:flutter/material.dart';
import 'package:forui/forui.dart';

class SettingsScreen extends StatelessWidget {
  const SettingsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return FScaffold(
      header: const FHeader(title: Text('Settings')),
      child: ListView(
        children: [
          _SettingsSection(
            title: 'Account',
            children: [
              FTile(
                prefix: const Icon(FIcons.user),
                title: const Text('Profile'),
                subtitle:
                    const Text('Change identifying details for your account'),
                suffix: const Icon(FIcons.chevronRight),
                onPress: () {
                  // TODO: Navigate to profile settings
                },
              ),
              FTile(
                prefix: const Icon(FIcons.shield),
                title: const Text('Security'),
                subtitle: const Text('Keep your account safe and sound'),
                suffix: const Icon(FIcons.chevronRight),
                onPress: () {
                  // TODO: Navigate to security settings
                },
              ),
            ],
          ),
          _SettingsSection(
            title: 'Streaming',
            children: [
              FTile(
                prefix: const Icon(FIcons.video),
                title: const Text('Stream'),
                subtitle:
                    const Text('Configure your livestream settings'),
                suffix: const Icon(FIcons.chevronRight),
                onPress: () {
                  // TODO: Navigate to stream settings
                },
              ),
              FTile(
                prefix: const Icon(FIcons.film),
                title: const Text('VODs'),
                subtitle: const Text('Manage your videos on demand'),
                suffix: const Icon(FIcons.chevronRight),
                onPress: () {
                  // TODO: Navigate to VODs settings
                },
              ),
            ],
          ),
          _SettingsSection(
            title: 'Preferences',
            children: [
              FTile(
                prefix: const Icon(FIcons.palette),
                title: const Text('Theme'),
                subtitle:
                    const Text('Customize the look and feel of the app'),
                suffix: const Icon(FIcons.chevronRight),
                onPress: () {
                  // TODO: Show theme picker
                },
              ),
              FTile(
                prefix: const Icon(FIcons.globe),
                title: const Text('Language'),
                subtitle: const Text('Select your preferred language'),
                suffix: const Icon(FIcons.chevronRight),
                onPress: () {
                  // TODO: Show language picker
                },
              ),
            ],
          ),
          Padding(
            padding: const EdgeInsets.all(16),
            child: FButton(
              variant: FButtonVariant.destructive,
              onPress: () {
                // TODO: Logout
              },
              prefix: const Icon(FIcons.logOut),
              child: const Text('Log out'),
            ),
          ),
        ],
      ),
    );
  }
}

class _SettingsSection extends StatelessWidget {
  final String title;
  final List<Widget> children;

  const _SettingsSection({
    required this.title,
    required this.children,
  });

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 24, 16, 8),
          child: Text(
            title,
            style: typography.sm.copyWith(
              color: colors.primary,
              fontWeight: FontWeight.w600,
            ),
          ),
        ),
        ...children,
      ],
    );
  }
}
