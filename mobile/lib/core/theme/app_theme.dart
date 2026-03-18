import 'package:flutter/services.dart';
import 'package:forui/forui.dart';

import 'app_colors.dart';

/// App theme configuration using Forui's theming system.
/// Colors are mapped from the web app's CSS variables to FColors.
abstract final class AppTheme {
  static FThemeData get light => FThemeData(
    debugLabel: 'LetSlive Light',
    colors: const FColors(
      brightness: Brightness.light,
      systemOverlayStyle: SystemUiOverlayStyle.dark,
      barrier: Color(0x33000000),
      background: AppColors.lightBackground,
      foreground: AppColors.lightForeground,
      primary: AppColors.lightPrimary,
      primaryForeground: AppColors.lightPrimaryForeground,
      secondary: AppColors.lightMuted,
      secondaryForeground: AppColors.lightForeground,
      muted: AppColors.lightMuted,
      mutedForeground: AppColors.lightForegroundMuted,
      destructive: AppColors.lightDestructive,
      destructiveForeground: AppColors.lightDestructiveForeground,
      error: AppColors.lightDestructive,
      errorForeground: AppColors.lightDestructiveForeground,
      card: AppColors.lightBackground,
      border: AppColors.lightBorder,
    ),
  );

  static FThemeData get dark => FThemeData(
    debugLabel: 'LetSlive Dark',
    colors: const FColors(
      brightness: Brightness.dark,
      systemOverlayStyle: SystemUiOverlayStyle.light,
      barrier: Color(0x7A000000),
      background: AppColors.darkBackground,
      foreground: AppColors.darkForeground,
      primary: AppColors.darkPrimary,
      primaryForeground: AppColors.darkPrimaryForeground,
      secondary: AppColors.darkMuted,
      secondaryForeground: AppColors.darkForeground,
      muted: AppColors.darkMuted,
      mutedForeground: AppColors.darkForegroundMuted,
      destructive: AppColors.darkDestructive,
      destructiveForeground: AppColors.darkDestructiveForeground,
      error: AppColors.darkDestructive,
      errorForeground: AppColors.darkDestructiveForeground,
      card: AppColors.darkCard,
      border: AppColors.darkBorder,
    ),
  );
}
