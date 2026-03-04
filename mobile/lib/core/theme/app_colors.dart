import 'dart:ui';

/// Color constants matching the web app's CSS variables.
/// Web uses HSL; these are converted to Flutter's Color (ARGB).
/// Used as the source of truth for FColors in app_theme.dart.
abstract final class AppColors {
  // ── Light theme ──────────────────────────────────────────────
  static const lightBackground = Color(0xFFFFFFFF); // 0 0% 100%
  static const lightForeground = Color(0xFF0F172A); // 222 47% 11%
  static const lightForegroundMuted = Color(0xFF3A3F4B); // 222 10.3% 25.4%

  static const lightMuted = Color(0xFFF1F5F9); // 210 40% 96%
  static const lightBorder = Color(0xFF2D3748); // 210 10% 23%

  static const lightPrimary = Color(0xFF6209B5); // 276 91% 38%
  static const lightPrimaryForeground = Color(0xFFEBEBEB); // 0 0% 92%

  static const lightSecondary = Color(0xFFE07B1A); // 28 80% 52%
  static const lightSecondaryForeground = Color(0xFFFFFFFF); // 0 0% 100%

  static const lightDestructive = Color(0xFFEF4444); // 0 90.3% 60%
  static const lightDestructiveForeground = Color(0xFFFFFFFF);

  static const lightSuccess = Color(0xFF22C55E); // 142 70.6% 45.3%

  // ── Dark theme ───────────────────────────────────────────────
  static const darkBackground = Color(0xFF0F172A); // 222 47% 11%
  static const darkForeground = Color(0xFFFFFFFF); // 0 0% 100%
  static const darkForegroundMuted = Color(0xFFCCCCCC); // 0 0% 80%

  static const darkMuted = Color(0xFF334155); // 220 20% 25%
  static const darkBorder = Color(0xFF3B4A5C); // 220 15% 30%

  static const darkPrimary = Color(0xFF6209B5); // 276 91% 38%
  static const darkPrimaryForeground = Color(0xFFEBEBEB); // 0 0% 92%

  static const darkSecondary = Color(0xFFF59E0B); // 28 90% 60%
  static const darkSecondaryForeground = Color(0xFF000000); // 0 0% 0%

  static const darkDestructive = Color(0xFFEF4444); // 0 90.3% 60%
  static const darkDestructiveForeground = Color(0xFFE5E5E5);

  static const darkSuccess = Color(0xFF22C55E); // 142 70.6% 45.3%

  static const darkCard = Color(0xFF1E293B); // 222 30% 20%
}
