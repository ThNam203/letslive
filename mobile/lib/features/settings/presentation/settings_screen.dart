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
    final themeMode = ref.watch(themeModeProvider);
    final locale = ref.watch(localeProvider);

    // Resolve the effective language code for display.
    final effectiveLanguageCode =
        locale?.languageCode ?? Localizations.localeOf(context).languageCode;
    final languageName =
        supportedLanguageNames[effectiveLanguageCode] ?? 'English';

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
          if (user != null)
            _SettingsSection(
              title: l10n.walletTitle,
              children: [
                FTile(
                  prefix: const Icon(FIcons.wallet),
                  title: Text(l10n.walletTitle),
                  subtitle: Text(l10n.walletDescription),
                  suffix: const Icon(FIcons.chevronRight),
                  onPress: () => context.push(AppRoutes.wallet),
                ),
              ],
            ),
          _SettingsSection(
            title: l10n.settingsThemesTitle,
            children: [
              FTile(
                prefix: const Icon(FIcons.palette),
                title: Text(l10n.settingsThemesTitle),
                subtitle: Text(_themeModeName(l10n, themeMode)),
                suffix: const Icon(FIcons.chevronRight),
                onPress: () => _showThemePicker(context, ref),
              ),
              FTile(
                prefix: const Icon(FIcons.globe),
                title: Text(l10n.settingsLanguageTitle),
                subtitle: Text(languageName),
                suffix: const Icon(FIcons.chevronRight),
                onPress: () => _showLanguagePicker(context, ref),
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

  // ---------------------------------------------------------------------------
  // Helpers
  // ---------------------------------------------------------------------------

  String _themeModeName(AppLocalizations l10n, AppThemeMode mode) {
    return switch (mode) {
      AppThemeMode.light => l10n.themeLight,
      AppThemeMode.dark => l10n.themeDark,
      AppThemeMode.system => l10n.themeSystem,
    };
  }

  void _showThemePicker(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final colors = context.theme.colors;
    final typography = context.theme.typography;
    final currentMode = ref.read(themeModeProvider);

    showModalBottomSheet(
      context: context,
      builder: (sheetContext) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
              child: Text(
                l10n.settingsThemesTitle,
                style: typography.lg.copyWith(fontWeight: FontWeight.w600),
              ),
            ),
            for (final mode in AppThemeMode.values)
              FTile(
                prefix: Icon(switch (mode) {
                  AppThemeMode.light => FIcons.sun,
                  AppThemeMode.dark => FIcons.moon,
                  AppThemeMode.system => FIcons.smartphone,
                }),
                title: Text(_themeModeName(l10n, mode)),
                suffix: currentMode == mode
                    ? Icon(FIcons.check, color: colors.primary)
                    : null,
                onPress: () {
                  ref.read(themeModeProvider.notifier).setThemeMode(mode);
                  Navigator.pop(sheetContext);
                },
              ),
            const SizedBox(height: 8),
          ],
        ),
      ),
    );
  }

  void _showLanguagePicker(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final colors = context.theme.colors;
    final typography = context.theme.typography;
    final savedLocale = ref.read(localeProvider);
    final effectiveCode =
        savedLocale?.languageCode ??
        Localizations.localeOf(context).languageCode;

    showModalBottomSheet(
      context: context,
      builder: (sheetContext) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
              child: Text(
                l10n.settingsLanguageTitle,
                style: typography.lg.copyWith(fontWeight: FontWeight.w600),
              ),
            ),
            for (final entry in supportedLanguageNames.entries)
              FTile(
                prefix: const Icon(FIcons.globe),
                title: Text(entry.value),
                suffix: effectiveCode == entry.key
                    ? Icon(FIcons.check, color: colors.primary)
                    : null,
                onPress: () {
                  ref
                      .read(localeProvider.notifier)
                      .setLocale(Locale(entry.key));
                  Navigator.pop(sheetContext);
                },
              ),
            const SizedBox(height: 8),
          ],
        ),
      ),
    );
  }
}

class _SettingsSection extends StatelessWidget {
  final String title;
  final List<Widget> children;

  const _SettingsSection({required this.title, required this.children});

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
