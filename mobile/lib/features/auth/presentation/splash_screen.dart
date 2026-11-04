import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/l10n/api_message.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/features/auth/auth_providers.dart';

/// Shown while the session loads at startup, so the app never flashes the
/// login screen for a signed-in user. If `GET /user/me` failed (server down,
/// no network) it shows the error and a retry.
class SplashScreen extends ConsumerWidget {
  const SplashScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final session = ref.watch(sessionProvider);
    // While retrying, Riverpod keeps the old error next to the loading flag.
    final error = session.isLoading ? null : session.error;
    return Scaffold(
      body: SafeArea(
        child: Center(
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(
                  l10n.common_app_title,
                  style: Theme.of(context).textTheme.headlineSmall,
                ),
                const SizedBox(height: 24),
                if (error == null)
                  const CircularProgressIndicator()
                else ...[
                  Text(
                    l10n.error_general_title,
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  const SizedBox(height: 8),
                  Text(
                    error is SessionLoadException
                        ? apiErrorMessage(l10n, error.response)
                        : l10n.error_general_description,
                    textAlign: TextAlign.center,
                  ),
                  const SizedBox(height: 16),
                  FilledButton(
                    key: const ValueKey('splash-retry'),
                    onPressed: () => ref.invalidate(sessionProvider),
                    // Web has no generic "retry" key; this is its "Try again".
                    child: Text(l10n.common_search_retry),
                  ),
                ],
              ],
            ),
          ),
        ),
      ),
    );
  }
}
