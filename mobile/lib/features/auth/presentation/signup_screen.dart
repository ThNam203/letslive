import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:letslive_mobile/core/l10n/api_message.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/router/app_router.dart';
import 'package:letslive_mobile/core/validation/field_limits.dart';
import 'package:letslive_mobile/features/auth/auth_providers.dart';
import 'package:letslive_mobile/features/auth/domain/auth_validators.dart';
import 'package:letslive_mobile/features/auth/presentation/auth_widgets.dart';
import 'package:letslive_mobile/features/auth/presentation/otp_dialog.dart';

/// Signup with email verification (web `app/(auth)/signup/page.tsx` and
/// `components/forms/SignupForm.tsx`): validate the form, mail a code
/// (`POST /auth/verify-email`), then create the account with the code in
/// [OtpDialog]. Google sign-in is offered too.
class SignupScreen extends ConsumerStatefulWidget {
  const SignupScreen({super.key});

  @override
  ConsumerState<SignupScreen> createState() => _SignupScreenState();
}

class _SignupScreenState extends ConsumerState<SignupScreen> {
  final _email = TextEditingController();
  final _username = TextEditingController();
  final _password = TextEditingController();
  final _confirmPassword = TextEditingController();
  AuthFieldError? _emailError;
  AuthFieldError? _usernameError;
  AuthFieldError? _passwordError;
  AuthFieldError? _confirmPasswordError;
  String? _requestError;
  bool _busy = false;

  @override
  void dispose() {
    _email.dispose();
    _username.dispose();
    _password.dispose();
    _confirmPassword.dispose();
    super.dispose();
  }

  bool _validate() {
    setState(() {
      _emailError = validateEmail(_email.text);
      _usernameError = validateUsername(_username.text);
      _passwordError = validateSignupPassword(_password.text);
      _confirmPasswordError = validateConfirmPassword(
        _confirmPassword.text,
        _password.text,
      );
      _requestError = null;
    });
    return _emailError == null &&
        _usernameError == null &&
        _passwordError == null &&
        _confirmPasswordError == null;
  }

  /// Mails a code; returns the error message or null.
  Future<String?> _sendCode() async {
    final l10n = AppLocalizations.of(context);
    final res = await ref
        .read(authRepositoryProvider)
        .requestEmailVerification(_email.text);
    return res.success ? null : apiErrorMessage(l10n, res);
  }

  Future<void> _submit() async {
    if (!_validate()) return;
    setState(() => _busy = true);
    final error = await _sendCode();
    if (!mounted) return;
    setState(() {
      _busy = false;
      _requestError = error;
    });
    if (error != null) return;

    final l10n = AppLocalizations.of(context);
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(l10n.api_response_res_succ_sent_verification_email),
      ),
    );
    await showDialog<void>(
      context: context,
      barrierDismissible: false,
      builder: (_) => OtpDialog(
        email: _email.text,
        onVerify: _createAccount,
        onResend: _sendCode,
      ),
    );
  }

  /// Creates the account; returns the error message or null. On success the
  /// router moves on by itself.
  Future<String?> _createAccount(String code) async {
    final l10n = AppLocalizations.of(context);
    final messenger = ScaffoldMessenger.of(context);
    final res = await ref
        .read(sessionProvider.notifier)
        .signup(
          email: _email.text,
          username: _username.text,
          password: _password.text,
          otpCode: code,
        );
    if (!res.success) return apiErrorMessage(l10n, res);
    messenger.showSnackBar(
      SnackBar(content: Text(l10n.auth_account_created_success)),
    );
    return null;
  }

  void _showGoogleError(String? message) {
    if (mounted) setState(() => _requestError = message);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return AuthScaffold(
      title: l10n.auth_signup_title,
      subtitle: l10n.auth_signup_subtitle,
      children: [
        GoogleSignInSection(onError: _showGoogleError),
        AutofillGroup(
          child: Column(
            children: [
              TextField(
                key: const ValueKey('email'),
                controller: _email,
                keyboardType: TextInputType.emailAddress,
                autofillHints: const [AutofillHints.email],
                maxLength: FieldLimits.emailMax,
                decoration: InputDecoration(
                  labelText: l10n.auth_email,
                  errorText: authFieldErrorText(l10n, _emailError),
                  counterText: '',
                  prefixIcon: const Icon(Icons.email_outlined),
                ),
              ),
              const SizedBox(height: 16),
              TextField(
                key: const ValueKey('username'),
                controller: _username,
                autofillHints: const [AutofillHints.newUsername],
                maxLength: FieldLimits.usernameMax,
                decoration: InputDecoration(
                  labelText: l10n.common_username,
                  errorText: authFieldErrorText(l10n, _usernameError),
                  counterText: '',
                  prefixIcon: const Icon(Icons.person_outline),
                ),
              ),
              const SizedBox(height: 16),
              PasswordField(
                fieldKey: const ValueKey('password'),
                controller: _password,
                label: l10n.auth_password,
                errorText: authFieldErrorText(l10n, _passwordError),
                autofillHint: AutofillHints.newPassword,
              ),
              const SizedBox(height: 16),
              PasswordField(
                fieldKey: const ValueKey('confirm-password'),
                controller: _confirmPassword,
                label: l10n.auth_confirm_password,
                errorText: authFieldErrorText(l10n, _confirmPasswordError),
                autofillHint: AutofillHints.newPassword,
              ),
            ],
          ),
        ),
        const SizedBox(height: 16),
        FormErrorText(_requestError),
        SubmitButton(
          key: const ValueKey('signup-submit'),
          label: l10n.auth_signup,
          busy: _busy,
          onPressed: _submit,
        ),
        const SizedBox(height: 16),
        Row(
          mainAxisAlignment: MainAxisAlignment.end,
          children: [
            Text(l10n.auth_have_account),
            TextButton(
              key: const ValueKey('go-login'),
              onPressed: () => context.go(AppRoutes.login),
              child: Text(l10n.auth_login),
            ),
          ],
        ),
      ],
    );
  }
}
