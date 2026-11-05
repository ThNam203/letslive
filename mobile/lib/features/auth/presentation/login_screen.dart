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

/// Email/password and Google login (web `app/(auth)/login/page.tsx` and
/// `components/forms/LoginForm.tsx`). On success the router leaves this
/// screen on its own, following the session.
class LoginScreen extends ConsumerStatefulWidget {
  const LoginScreen({super.key});

  @override
  ConsumerState<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends ConsumerState<LoginScreen> {
  final _email = TextEditingController();
  final _password = TextEditingController();
  AuthFieldError? _emailError;
  AuthFieldError? _passwordError;
  String? _requestError;
  bool _busy = false;

  @override
  void dispose() {
    _email.dispose();
    _password.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    final email = _email.text;
    final password = _password.text;
    setState(() {
      _emailError = validateEmail(email);
      _passwordError = validateLoginPassword(password);
      _requestError = null;
    });
    if (_emailError != null || _passwordError != null) return;

    final l10n = AppLocalizations.of(context);
    setState(() => _busy = true);
    final res = await ref.read(sessionProvider.notifier).login(email, password);
    if (!mounted) return;
    setState(() {
      _busy = false;
      if (!res.success) _requestError = apiErrorMessage(l10n, res);
    });
  }

  void _showGoogleError(String? message) {
    if (mounted) setState(() => _requestError = message);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return AuthScaffold(
      title: l10n.auth_login_title,
      subtitle: l10n.auth_login_subtitle,
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
              PasswordField(
                fieldKey: const ValueKey('password'),
                controller: _password,
                label: l10n.auth_password,
                errorText: authFieldErrorText(l10n, _passwordError),
                autofillHint: AutofillHints.password,
              ),
            ],
          ),
        ),
        const SizedBox(height: 16),
        FormErrorText(_requestError),
        SubmitButton(
          key: const ValueKey('login-submit'),
          label: l10n.auth_login,
          busy: _busy,
          onPressed: _submit,
        ),
        const SizedBox(height: 16),
        Row(
          mainAxisAlignment: MainAxisAlignment.end,
          children: [
            Text(l10n.auth_no_account),
            TextButton(
              key: const ValueKey('go-signup'),
              onPressed: () => context.go(AppRoutes.signup),
              child: Text(l10n.auth_signup),
            ),
          ],
        ),
      ],
    );
  }
}
