import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/l10n/api_message.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/features/auth/auth_providers.dart';
import 'package:letslive_mobile/features/auth/domain/auth_validators.dart';
import 'package:letslive_mobile/features/auth/presentation/auth_widgets.dart';

/// Username picker for sessions without one (Google signups; web
/// `app/(auth)/account-setup/page.tsx`). The router keeps the user here until
/// it succeeds. Unlike web there is no address bar to leave by, so the
/// screen also offers logout.
class AccountSetupScreen extends ConsumerStatefulWidget {
  const AccountSetupScreen({super.key});

  @override
  ConsumerState<AccountSetupScreen> createState() => _AccountSetupScreenState();
}

class _AccountSetupScreenState extends ConsumerState<AccountSetupScreen> {
  final _username = TextEditingController();
  String? _error;
  bool _busy = false;

  @override
  void dispose() {
    _username.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    final l10n = AppLocalizations.of(context);
    final username = _username.text.trim();
    final invalid = validateAccountSetupUsername(username);
    if (invalid != null) {
      setState(() => _error = authFieldErrorText(l10n, invalid));
      return;
    }
    setState(() {
      _busy = true;
      _error = null;
    });
    final res = await ref
        .read(sessionProvider.notifier)
        .completeAccountSetup(username);
    if (!mounted) return;
    setState(() {
      _busy = false;
      if (!res.success) _error = apiErrorMessage(l10n, res);
    });
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return AuthScaffold(
      title: l10n.auth_account_setup_title,
      subtitle: l10n.auth_account_setup_subtitle,
      children: [
        TextField(
          key: const ValueKey('username'),
          controller: _username,
          enabled: !_busy,
          maxLength: usernameMaxLength,
          autofillHints: const [AutofillHints.newUsername],
          onChanged: (_) => setState(() => _error = null),
          decoration: InputDecoration(
            labelText: l10n.common_username,
            errorText: _error,
            counterText: '',
            prefixIcon: const Icon(Icons.person_outline),
          ),
        ),
        const SizedBox(height: 16),
        SubmitButton(
          key: const ValueKey('setup-submit'),
          label: l10n.auth_account_setup_submit,
          busy: _busy,
          onPressed: _submit,
        ),
        const SizedBox(height: 8),
        TextButton(
          key: const ValueKey('setup-logout'),
          onPressed: _busy
              ? null
              : () => ref.read(sessionProvider.notifier).logout(),
          child: Text(l10n.auth_logout),
        ),
      ],
    );
  }
}
