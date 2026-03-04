import 'package:flutter/widgets.dart';
import 'package:forui/forui.dart';

class NotificationsScreen extends StatelessWidget {
  const NotificationsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;

    return FScaffold(
      header: FHeader(
        title: const Text('Notifications'),
        suffixes: [
          FButton(
            variant: FButtonVariant.ghost,
            onPress: () {
              // TODO: Mark all as read
            },
            child: const Text('Mark all as read'),
          ),
        ],
      ),
      child: Center(
        child: Text(
          'No notifications yet',
          style: typography.base.copyWith(color: colors.mutedForeground),
        ),
      ),
    );
  }
}
