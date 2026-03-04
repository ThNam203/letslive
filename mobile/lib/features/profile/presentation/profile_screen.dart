import 'package:flutter/widgets.dart';
import 'package:forui/forui.dart';

class ProfileScreen extends StatelessWidget {
  final String userId;

  const ProfileScreen({super.key, required this.userId});

  @override
  Widget build(BuildContext context) {
    final typography = context.theme.typography;

    return FScaffold(
      header: const FHeader.nested(title: Text('Profile')),
      child: Center(
        child: Text(
          'User profile: $userId',
          style: typography.base,
        ),
      ),
    );
  }
}
