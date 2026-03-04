import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:forui/forui.dart';
import 'package:go_router/go_router.dart';

import '../../../core/router/app_router.dart';
import '../../../l10n/app_localizations.dart';
import '../../../providers.dart';

class SettingsScreen extends ConsumerWidget {
  const SettingsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final user = ref.watch(currentUserProvider);

    return FScaffold(
      header: FHeader(title: Text(l10n.settingsTitle)),
      child: ListView(
        children: [
          _SettingsSection(
            title: l10n.settingsNavProfile,
            children: [
              FTile(
                prefix: const Icon(FIcons.user),
                title: Text(l10n.settingsNavProfile),
                subtitle: Text(l10n.settingsProfileDescription),
                suffix: const Icon(FIcons.chevronRight),
                onPress: () => context.push(AppRoutes.settingsProfile),
              ),
              FTile(
                prefix: const Icon(FIcons.shield),
                title: Text(l10n.settingsNavSecurity),
                subtitle: Text(l10n.settingsSecurityDescription),
                suffix: const Icon(FIcons.chevronRight),
                onPress: () => context.push(AppRoutes.settingsSecurity),
              ),
            ],
          ),
          _SettingsSection(
            title: l10n.settingsNavStream,
            children: [
              FTile(
                prefix: const Icon(FIcons.video),
                title: Text(l10n.settingsNavStream),
                subtitle: Text(l10n.settingsStreamDescription),
                suffix: const Icon(FIcons.chevronRight),
                onPress: () => context.push(AppRoutes.settingsStream),
              ),
              FTile(
                prefix: const Icon(FIcons.film),
                title: Text(l10n.settingsNavVods),
                subtitle: Text(l10n.settingsVodsDescription),
                suffix: const Icon(FIcons.chevronRight),
                onPress: () => context.push(AppRoutes.settingsVods),
              ),
            ],
          ),
          _SettingsSection(
            title: l10n.settingsThemesTitle,
            children: [
              FTile(
                prefix: const Icon(FIcons.palette),
                title: Text(l10n.settingsThemesTitle),
                subtitle: Text(l10n.settingsThemesDescription),
                suffix: const Icon(FIcons.chevronRight),
                onPress: () {
                  // TODO: Show theme picker
                },
              ),
              FTile(
                prefix: const Icon(FIcons.globe),
                title: Text(l10n.settingsLanguageTitle),
                subtitle: Text(l10n.settingsLanguageDescription),
                suffix: const Icon(FIcons.chevronRight),
                onPress: () {
                  // TODO: Show language picker
                },
              ),
            ],
          ),
          if (user != null)
            Padding(
              padding: const EdgeInsets.all(16),
              child: FButton(
                variant: FButtonVariant.destructive,
                onPress: () async {
                  await ref.read(currentUserProvider.notifier).logout();
                  if (context.mounted) context.go(AppRoutes.login);
                },
                prefix: const Icon(FIcons.logOut),
                child: Text(l10n.authLogout),
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
