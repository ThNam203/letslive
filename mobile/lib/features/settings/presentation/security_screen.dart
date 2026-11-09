/// Account security (web `/settings/security`): the account email, the
/// phone number ([PhoneNumberTile]) and the change-password form. The form
/// validates like web's `changePassword.ts` and calls `PATCH /auth/password`;
/// a server refusal (e.g. the current password is wrong) shows under the
/// confirm field, as on web. Accounts that signed in with Google have no
/// password to change and see a note instead. The stream key is on the
/// Stream settings screen instead.
library;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/l10n/api_message.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/features/auth/auth_providers.dart';
import 'package:letslive_mobile/features/auth/domain/auth_validators.dart';
import 'package:letslive_mobile/features/auth/presentation/auth_widgets.dart';
import 'package:letslive_mobile/features/settings/presentation/phone_number_tile.dart';

/// `MeUser.authProvider` of accounts with a password.
const String _localProvider = 'local';

/// The security settings screen.
class SecurityScreen extends ConsumerWidget {
  const SecurityScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final user = ref.watch(sessionProvider).value?.user;
    return Scaffold(
      appBar: AppBar(title: Text(l10n.settings_navigation_security)),
      body: user == null
          ? const SizedBox.shrink()
          : ListView(
              padding: const EdgeInsets.all(16),
              children: [
                Text(
                  l10n.settings_security_contact_title,
                  style: Theme.of(context).textTheme.titleMedium,
                ),
                ListTile(
                  contentPadding: EdgeInsets.zero,
                  title: Text(l10n.settings_security_contact_email),
                  subtitle: Text(user.email),
                ),
                PhoneNumberTile(phoneNumber: user.phoneNumber),
                const Divider(),
                Text(
                  l10n.settings_security_security_password_title,
                  style: Theme.of(context).textTheme.titleMedium,
                ),
                const SizedBox(height: 8),
                if (user.authProvider == _localProvider)
                  const _ChangePasswordForm()
                else
                  Text(
                    l10n.settings_security_security_password_third_party_description,
                  ),
              ],
            ),
    );
  }
}

class _ChangePasswordForm extends ConsumerStatefulWidget {
  const _ChangePasswordForm();

  @override
  ConsumerState<_ChangePasswordForm> createState() =>
      _ChangePasswordFormState();
}

class _ChangePasswordFormState extends ConsumerState<_ChangePasswordForm> {
  final _current = TextEditingController();
  final _next = TextEditingController();
  final _confirm = TextEditingController();
  ChangePasswordErrors? _errors;
  String? _serverError;
  bool _busy = false;

  @override
  void dispose() {
    _current.dispose();
    _next.dispose();
    _confirm.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    final l10n = AppLocalizations.of(context);
    final messenger = ScaffoldMessenger.of(context);
    final errors = validateChangePassword(
      current: _current.text,
      next: _next.text,
      confirm: _confirm.text,
    );
    setState(() {
      _errors = errors;
      _serverError = null;
    });
    if (errors.current != null ||
        errors.next != null ||
        errors.confirm != null) {
      return;
    }
    setState(() => _busy = true);
    final res = await ref
        .read(authRepositoryProvider)
        .changePassword(oldPassword: _current.text, newPassword: _next.text);
    if (!mounted) return;
    setState(() {
      _busy = false;
      if (!res.success) {
        _serverError = apiErrorMessage(l10n, res);
      }
    });
    if (!res.success) return;
    _current.clear();
    _next.clear();
    _confirm.clear();
    messenger.showSnackBar(
      SnackBar(
        content: Text(l10n.settings_security_security_password_updated_success),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final errors = _errors;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        PasswordField(
          fieldKey: const ValueKey('current-password'),
          controller: _current,
          label: l10n.settings_security_security_password_form_current_label,
          errorText: authFieldErrorText(l10n, errors?.current),
          autofillHint: AutofillHints.password,
        ),
        const SizedBox(height: 12),
        PasswordField(
          fieldKey: const ValueKey('new-password'),
          controller: _next,
          label: l10n.settings_security_security_password_form_new_label,
          errorText: authFieldErrorText(l10n, errors?.next),
          autofillHint: AutofillHints.newPassword,
        ),
        const SizedBox(height: 12),
        PasswordField(
          fieldKey: const ValueKey('confirm-password'),
          controller: _confirm,
          label: l10n.settings_security_security_password_form_confirm_label,
          // Web shows a server refusal under this field too.
          errorText: _serverError ?? authFieldErrorText(l10n, errors?.confirm),
          autofillHint: AutofillHints.newPassword,
        ),
        const SizedBox(height: 16),
        SubmitButton(
          key: const ValueKey('password-submit'),
          label: l10n.settings_security_security_password_form_submit,
          busy: _busy,
          onPressed: _submit,
        ),
      ],
    );
  }
}
