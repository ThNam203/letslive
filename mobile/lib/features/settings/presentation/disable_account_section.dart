/// "Disable your account", the last section of the profile settings (web
/// `disable-account-dialog.tsx`): a confirmation dialog, then
/// `PUT /user/me {status: disabled}`, then logout. Signing in again
/// reactivates the account (the dialog says so, as on web). A refusal keeps
/// the user signed in and says why.
library;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/widgets/action_feedback.dart';
import 'package:letslive_mobile/features/auth/auth_providers.dart';

/// `MeUser.status` of a disabled account (backend `UpdateUserRequestDTO`).
const String _disabledStatus = 'disabled';

/// The disable-account explanation and button.
class DisableAccountSection extends ConsumerStatefulWidget {
  const DisableAccountSection({super.key});

  @override
  ConsumerState<DisableAccountSection> createState() =>
      _DisableAccountSectionState();
}

class _DisableAccountSectionState extends ConsumerState<DisableAccountSection> {
  bool _busy = false;

  Future<void> _disable() async {
    final l10n = AppLocalizations.of(context);
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(l10n.settings_disable_dialog_title),
        content: Text(l10n.settings_disable_dialog_description),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(false),
            child: Text(l10n.common_cancel),
          ),
          FilledButton(
            onPressed: () => Navigator.of(context).pop(true),
            child: Text(l10n.settings_disable_dialog_confirm),
          ),
        ],
      ),
    );
    if (confirmed != true || !mounted) return;
    setState(() => _busy = true);
    final session = ref.read(sessionProvider.notifier);
    final res = await session.updateProfile(status: _disabledStatus);
    if (!res.success) {
      if (!mounted) return;
      setState(() => _busy = false);
      showActionError(context, res);
      return;
    }
    // The router shows the login screen once the session ends.
    await session.logout();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final theme = Theme.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(l10n.settings_disable_title, style: theme.textTheme.titleMedium),
        const SizedBox(height: 4),
        Text(l10n.settings_disable_description),
        const SizedBox(height: 8),
        Text(l10n.settings_disable_note, style: theme.textTheme.bodySmall),
        const SizedBox(height: 12),
        FilledButton(
          key: const ValueKey('disable-account'),
          style: FilledButton.styleFrom(
            backgroundColor: theme.colorScheme.error,
            foregroundColor: theme.colorScheme.onError,
          ),
          onPressed: _busy ? null : _disable,
          child: Text(
            l10n.settings_disable_button,
            textAlign: TextAlign.center,
          ),
        ),
      ],
    );
  }
}
