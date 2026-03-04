import 'package:flutter/widgets.dart';
import 'package:forui/forui.dart';

class MessagesScreen extends StatelessWidget {
  const MessagesScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;

    return FScaffold(
      header: FHeader(
        title: const Text('Messages'),
        suffixes: [
          FButton.icon(
            onPress: () {
              // TODO: New conversation
            },
            child: const Icon(FIcons.squarePen),
          ),
        ],
      ),
      child: Center(
        child: Text(
          'No conversations yet',
          style: typography.base.copyWith(color: colors.mutedForeground),
        ),
      ),
    );
  }
}
