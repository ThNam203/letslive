import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/router/app_routes.dart';

/// A centered icon, optional [title], [message] and a back button. Back
/// pops to where the user came from, or goes home after a deep link. Used
/// for content that vanished (Review Focus 5): an unknown channel or video.
class BackNotice extends StatelessWidget {
  const BackNotice({
    required this.icon,
    required this.message,
    this.title,
    super.key,
  });

  final IconData icon;
  final String? title;
  final String message;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final theme = Theme.of(context);
    final heading = title;
    return ListView(
      padding: const EdgeInsets.all(24),
      children: [
        Icon(icon, size: 48, color: theme.colorScheme.onSurfaceVariant),
        const SizedBox(height: 12),
        if (heading != null)
          Text(
            heading,
            textAlign: TextAlign.center,
            style: theme.textTheme.titleMedium,
          ),
        Text(message, textAlign: TextAlign.center),
        const SizedBox(height: 16),
        Center(
          child: FilledButton.tonal(
            onPressed: () =>
                context.canPop() ? context.pop() : context.go(AppRoutes.home),
            child: Text(l10n.users_channel_back),
          ),
        ),
      ],
    );
  }
}
