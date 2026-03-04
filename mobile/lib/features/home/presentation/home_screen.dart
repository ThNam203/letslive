import 'package:flutter/widgets.dart';
import 'package:forui/forui.dart';

class HomeScreen extends StatelessWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;

    return FScaffold(
      header: FHeader(
        title: const Text("Let's Live"),
        suffixes: [
          FButton.icon(
            onPress: () {
              // TODO: Navigate to search
            },
            child: const Icon(FIcons.search),
          ),
        ],
      ),
      child: Center(
        child: Text(
          'Home - Livestreams & VODs',
          style: typography.base.copyWith(color: colors.mutedForeground),
        ),
      ),
    );
  }
}
