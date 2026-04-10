import 'package:flutter/widgets.dart';
import 'package:forui/forui.dart';

class EmptyStateView extends StatelessWidget {
  final IconData icon;
  final String title;
  final String? description;
  final double iconSize;
  final EdgeInsetsGeometry padding;

  const EmptyStateView({
    super.key,
    required this.icon,
    required this.title,
    this.description,
    this.iconSize = 48,
    this.padding = const EdgeInsets.all(24),
  });

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;

    return Center(
      child: Padding(
        padding: padding,
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(icon, size: iconSize, color: colors.mutedForeground),
            const SizedBox(height: 16),
            Text(
              title,
              style: typography.lg.copyWith(fontWeight: FontWeight.w600),
              textAlign: TextAlign.center,
            ),
            if (description != null && description!.isNotEmpty) ...[
              const SizedBox(height: 8),
              Text(
                description!,
                style: typography.sm.copyWith(color: colors.mutedForeground),
                textAlign: TextAlign.center,
              ),
            ],
          ],
        ),
      ),
    );
  }
}
