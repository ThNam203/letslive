import 'package:flutter/widgets.dart';
import 'package:forui/forui.dart';

class ErrorDisplay extends StatelessWidget {
  final String title;
  final String? message;
  final VoidCallback? onRetry;

  const ErrorDisplay({
    super.key,
    required this.title,
    this.message,
    this.onRetry,
  });

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;

    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              FIcons.circleAlert,
              size: 48,
              color: colors.error,
            ),
            const SizedBox(height: 16),
            Text(
              title,
              style: typography.lg.copyWith(
                fontWeight: FontWeight.w600,
              ),
              textAlign: TextAlign.center,
            ),
            if (message != null) ...[
              const SizedBox(height: 8),
              Text(
                message!,
                style: typography.sm.copyWith(
                  color: colors.mutedForeground,
                ),
                textAlign: TextAlign.center,
              ),
            ],
            if (onRetry != null) ...[
              const SizedBox(height: 24),
              FButton(
                onPress: onRetry,
                prefix: const Icon(FIcons.refreshCw),
                child: const Text('Retry'),
              ),
            ],
          ],
        ),
      ),
    );
  }
}
