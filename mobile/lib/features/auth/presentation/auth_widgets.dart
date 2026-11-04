/// Building blocks shared by the login, signup and account-setup screens.
library;

import 'dart:developer' as developer;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_sign_in/google_sign_in.dart';
import 'package:letslive_mobile/core/l10n/api_message.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/features/auth/auth_providers.dart';
import 'package:letslive_mobile/features/auth/domain/auth_validators.dart';

/// The localized message for a validation [error], or null.
String? authFieldErrorText(AppLocalizations l10n, AuthFieldError? error) =>
    switch (error) {
      null => null,
      AuthFieldError.emailRequired => l10n.error_email_required,
      // Web has no translation for "too long"; the input caps the length, so
      // this is unreachable from the form.
      AuthFieldError.emailInvalid ||
      AuthFieldError.emailTooLong => l10n.error_email_invalid,
      AuthFieldError.passwordRequired => l10n.error_password_required,
      AuthFieldError.passwordTooShort => l10n.error_password_too_short(
        passwordMinLength,
      ),
      AuthFieldError.passwordTooLong => l10n.error_password_too_long(
        passwordMaxLength,
      ),
      AuthFieldError.passwordMissingLowercase =>
        l10n.error_password_missing_lowercase,
      AuthFieldError.passwordMissingUppercase =>
        l10n.error_password_missing_uppercase,
      AuthFieldError.passwordMissingSpecial =>
        l10n.error_password_missing_special,
      AuthFieldError.confirmPasswordRequired =>
        l10n.error_confirm_password_required,
      AuthFieldError.passwordsDoNotMatch => l10n.error_passwords_do_not_match,
      AuthFieldError.usernameRequired => l10n.error_username_required,
      AuthFieldError.usernameTooShort => l10n.error_username_too_short,
      AuthFieldError.usernameTooLong => l10n.error_username_too_long,
    };

/// Page frame of the auth screens (web `app/(auth)/layout.tsx`): brand,
/// title, subtitle, then the form, scrollable for small screens.
class AuthScaffold extends StatelessWidget {
  const AuthScaffold({
    required this.title,
    required this.subtitle,
    required this.children,
    super.key,
  });

  final String title;
  final String subtitle;
  final List<Widget> children;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Scaffold(
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(24),
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 600),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Text(
                    AppLocalizations.of(context).common_app_title,
                    style: theme.textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 8),
                  Text(
                    title,
                    style: theme.textTheme.headlineSmall?.copyWith(
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(subtitle),
                  const SizedBox(height: 16),
                  ...children,
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}

/// A password input with a show/hide toggle.
class PasswordField extends StatefulWidget {
  const PasswordField({
    required this.fieldKey,
    required this.controller,
    required this.label,
    required this.errorText,
    required this.autofillHint,
    super.key,
  });

  /// Key of the inner [TextField] (used by tests).
  final Key fieldKey;
  final TextEditingController controller;
  final String label;
  final String? errorText;
  final String autofillHint;

  @override
  State<PasswordField> createState() => _PasswordFieldState();
}

class _PasswordFieldState extends State<PasswordField> {
  bool _hidden = true;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return TextField(
      key: widget.fieldKey,
      controller: widget.controller,
      obscureText: _hidden,
      maxLength: passwordMaxLength,
      autofillHints: [widget.autofillHint],
      decoration: InputDecoration(
        labelText: widget.label,
        errorText: widget.errorText,
        counterText: '',
        prefixIcon: const Icon(Icons.lock_outline),
        suffixIcon: IconButton(
          tooltip: _hidden ? l10n.auth_show_password : l10n.auth_hide_password,
          icon: Icon(_hidden ? Icons.visibility : Icons.visibility_off),
          onPressed: () => setState(() => _hidden = !_hidden),
        ),
      ),
    );
  }
}

/// Primary button showing a spinner (and ignoring taps) while [busy].
/// Full width unless [expanded] is false (e.g. in dialog actions).
class SubmitButton extends StatelessWidget {
  const SubmitButton({
    required this.label,
    required this.busy,
    required this.onPressed,
    this.expanded = true,
    super.key,
  });

  final String label;
  final bool busy;
  final VoidCallback onPressed;
  final bool expanded;

  @override
  Widget build(BuildContext context) => FilledButton(
    style: FilledButton.styleFrom(
      minimumSize: expanded ? const Size.fromHeight(48) : null,
    ),
    onPressed: busy ? null : onPressed,
    child: busy
        ? const SizedBox.square(
            dimension: 20,
            child: CircularProgressIndicator(strokeWidth: 2),
          )
        : Text(label.toUpperCase()),
  );
}

/// A request failure shown above the submit button; submitting again is the
/// retry, as on web.
class FormErrorText extends StatelessWidget {
  const FormErrorText(this.message, {super.key});

  final String? message;

  @override
  Widget build(BuildContext context) {
    final text = message;
    if (text == null) return const SizedBox.shrink();
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Text(
        text,
        style: TextStyle(color: Theme.of(context).colorScheme.error),
      ),
    );
  }
}

/// "Google" button plus the "or" divider; renders nothing when Google
/// sign-in is not configured. Errors go to [onError]; success needs no
/// callback because the router follows the session.
class GoogleSignInSection extends ConsumerStatefulWidget {
  const GoogleSignInSection({required this.onError, super.key});

  final ValueChanged<String?> onError;

  @override
  ConsumerState<GoogleSignInSection> createState() =>
      _GoogleSignInSectionState();
}

class _GoogleSignInSectionState extends ConsumerState<GoogleSignInSection> {
  bool _busy = false;

  Future<void> _signIn() async {
    final google = ref.read(googleSignInServiceProvider);
    if (google == null) return;
    final l10n = AppLocalizations.of(context);
    setState(() => _busy = true);
    widget.onError(null);
    try {
      final idToken = await google.obtainIdToken();
      // Null: the user closed the sheet. Unmounted: `ref` is no longer usable.
      if (idToken == null || !mounted) return;
      final res = await ref
          .read(sessionProvider.notifier)
          .loginWithGoogle(idToken);
      if (!res.success) widget.onError(apiErrorMessage(l10n, res));
    } on Exception catch (e, stack) {
      // SDK descriptions are developer text (e.g. "SHA-1 not registered"),
      // so they are logged and the user sees a generic, localized message.
      final detail = e is GoogleSignInException
          ? '${e.code.name}: ${e.description}'
          : '$e';
      developer.log(
        'Google sign-in failed: $detail',
        name: 'auth',
        error: e,
        stackTrace: stack,
      );
      widget.onError(l10n.api_response_default_error);
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    if (ref.watch(googleSignInServiceProvider) == null) {
      return const SizedBox.shrink();
    }
    final l10n = AppLocalizations.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        OutlinedButton.icon(
          key: const ValueKey('google-sign-in'),
          style: OutlinedButton.styleFrom(
            minimumSize: const Size.fromHeight(48),
          ),
          onPressed: _busy ? null : _signIn,
          icon: const Icon(Icons.g_mobiledata, size: 32),
          // Web shows the brand name untranslated.
          label: const Text('Google'),
        ),
        Padding(
          padding: const EdgeInsets.symmetric(vertical: 16),
          child: Row(
            children: [
              const Expanded(child: Divider()),
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 16),
                child: Text(l10n.common_or),
              ),
              const Expanded(child: Divider()),
            ],
          ),
        ),
      ],
    );
  }
}
