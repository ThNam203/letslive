import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/l10n/api_message.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/network/api_load.dart';

/// One titled block of the home feed that loads on its own: a spinner while
/// loading, an error with a retry button when it failed, [empty] (or nothing
/// when null) for an empty list, else [builder]'s content.
///
/// While reloading with data on screen the old content stays (no flicker on
/// pull-to-refresh); after an error the spinner shows so a retry gives
/// feedback.
class FeedSection<T> extends StatelessWidget {
  const FeedSection({
    required this.title,
    required this.value,
    required this.onRetry,
    required this.builder,
    this.empty,
    super.key,
  });

  final String title;
  final AsyncValue<List<T>> value;
  final VoidCallback onRetry;
  final Widget Function(BuildContext context, List<T> items) builder;

  /// Shown for an empty list; null hides the whole section (like web's
  /// channel lists).
  final Widget? empty;

  @override
  Widget build(BuildContext context) {
    final Widget body;
    final items = value.value;
    if (value.hasError && !value.isLoading) {
      body = _SectionError(error: value.error!, onRetry: onRetry);
    } else if (items == null) {
      body = const Padding(
        padding: EdgeInsets.all(24),
        child: Center(child: CircularProgressIndicator()),
      );
    } else if (items.isEmpty) {
      final emptyBody = empty;
      if (emptyBody == null) return const SizedBox.shrink();
      body = emptyBody;
    } else {
      body = builder(context, items);
    }
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            child: Text(
              title,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: Theme.of(context).textTheme.titleLarge,
            ),
          ),
          const SizedBox(height: 8),
          body,
        ],
      ),
    );
  }
}

/// Title and description of an empty section (web's empty states).
class EmptyNotice extends StatelessWidget {
  const EmptyNotice({
    required this.icon,
    required this.title,
    required this.description,
    super.key,
  });

  final IconData icon;
  final String title;
  final String description;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Padding(
      padding: const EdgeInsets.all(24),
      child: Column(
        children: [
          Icon(icon, size: 48, color: theme.colorScheme.onSurfaceVariant),
          const SizedBox(height: 12),
          Text(
            title,
            textAlign: TextAlign.center,
            style: theme.textTheme.titleMedium,
          ),
          const SizedBox(height: 4),
          Text(description, textAlign: TextAlign.center),
        ],
      ),
    );
  }
}

class _SectionError extends StatelessWidget {
  const _SectionError({required this.error, required this.onRetry});

  final Object error;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final failed = error;
    return Padding(
      padding: const EdgeInsets.all(16),
      child: Column(
        children: [
          Text(
            failed is ApiLoadException
                ? apiErrorMessage(l10n, failed.response)
                : l10n.error_general_description,
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 8),
          FilledButton(onPressed: onRetry, child: Text(l10n.auth_retry)),
        ],
      ),
    );
  }
}
