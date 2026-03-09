import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:forui/forui.dart';
import 'package:go_router/go_router.dart';
import 'package:google_sign_in/google_sign_in.dart';

import '../../../core/config/env.dart';
import '../../../core/constants/field_limits.dart';
import '../../../core/constants/password.dart';
import '../../../core/router/app_router.dart';
import '../../../core/utils/api_error_localizer.dart';
import '../../../l10n/app_localizations.dart';
import '../../../providers.dart';

class SignupScreen extends ConsumerStatefulWidget {
  const SignupScreen({super.key});

  @override
  ConsumerState<SignupScreen> createState() => _SignupScreenState();
}

class _SignupScreenState extends ConsumerState<SignupScreen> {
  final _formKey = GlobalKey<FormState>();
  final _emailController = TextEditingController();
  final _usernameController = TextEditingController();
  final _passwordController = TextEditingController();
  final _confirmPasswordController = TextEditingController();
  final _otpController = TextEditingController();
  bool _isLoading = false;
  bool _isGoogleLoading = false;
  Timer? _resendTimer;
  int _resendCountdown = 0;

  @override
  void dispose() {
    _resendTimer?.cancel();
    _emailController.dispose();
    _usernameController.dispose();
    _passwordController.dispose();
    _confirmPasswordController.dispose();
    _otpController.dispose();
    super.dispose();
  }

  void _startResendCountdown() {
    _resendCountdown = 60;
    _resendTimer?.cancel();
    _resendTimer = Timer.periodic(const Duration(seconds: 1), (timer) {
      if (_resendCountdown <= 1) {
        timer.cancel();
        setState(() => _resendCountdown = 0);
      } else {
        setState(() => _resendCountdown--);
      }
    });
  }

  Future<void> _handleSignup() async {
    if (!_formKey.currentState!.validate()) return;

    setState(() => _isLoading = true);

    try {
      final authRepo = ref.read(authRepositoryProvider);
      final response = await authRepo.requestVerification(
        email: _emailController.text.trim(),
      );

      if (!mounted) return;

      if (response.success) {
        _showOtpDialog();
      } else {
        final errorMsg = getLocalizedApiMessage(context, response.key);
        showFToast(
          context: context,
          title: Text(errorMsg),
          icon: const Icon(FIcons.circleAlert),
        );
      }
    } on DioException catch (_) {
      if (mounted) {
        final l10n = AppLocalizations.of(context);
        showFToast(
          context: context,
          title: Text(l10n.fetchError),
          icon: const Icon(FIcons.circleAlert),
        );
      }
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  Future<void> _handleGoogleSignIn() async {
    setState(() => _isGoogleLoading = true);

    try {
      final googleSignIn = GoogleSignIn(
        serverClientId: Env.googleOAuthServerClientId,
        scopes: ['email', 'profile'],
      );
      final account = await googleSignIn.signIn();
      if (account == null) {
        if (mounted) {
          final l10n = AppLocalizations.of(context);
          showFToast(
            context: context,
            title: Text(l10n.authGoogleSignInCancelled),
            icon: const Icon(FIcons.circleAlert),
          );
        }
        return;
      }

      final auth = await account.authentication;
      final idToken = auth.idToken;
      if (idToken == null) {
        if (mounted) {
          final l10n = AppLocalizations.of(context);
          showFToast(
            context: context,
            title: Text(l10n.authGoogleSignInFailed),
            icon: const Icon(FIcons.circleAlert),
          );
        }
        return;
      }

      final authRepo = ref.read(authRepositoryProvider);
      final response = await authRepo.loginWithGoogle(idToken: idToken);

      if (!mounted) return;

      if (response.success) {
        await ref.read(currentUserProvider.notifier).fetchMe();
        if (mounted) context.go(AppRoutes.home);
      } else {
        final errorMsg = getLocalizedApiMessage(context, response.key);
        showFToast(
          context: context,
          title: Text(errorMsg),
          icon: const Icon(FIcons.circleAlert),
        );
      }
    } on DioException catch (_) {
      if (mounted) {
        final l10n = AppLocalizations.of(context);
        showFToast(
          context: context,
          title: Text(l10n.authGoogleSignInFailed),
          icon: const Icon(FIcons.circleAlert),
        );
      }
    } catch (_) {
      if (mounted) {
        final l10n = AppLocalizations.of(context);
        showFToast(
          context: context,
          title: Text(l10n.authGoogleSignInFailed),
          icon: const Icon(FIcons.circleAlert),
        );
      }
    } finally {
      if (mounted) setState(() => _isGoogleLoading = false);
    }
  }

  void _showOtpDialog() {
    _otpController.clear();
    _startResendCountdown();

    showFDialog(
      context: context,
      barrierDismissible: false,
      builder: (dialogContext, style, animation) {
        bool isOtpSubmitting = false;
        String otpError = '';

        return StatefulBuilder(
          builder: (context, setDialogState) {
            final l10n = AppLocalizations.of(context);
            final colors = context.theme.colors;
            final typography = context.theme.typography;

            return FDialog(
              animation: animation,
              title: Text(l10n.authEnterVerificationCode),
              body: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text.rich(
                    TextSpan(children: [
                      TextSpan(text: l10n.authOtpDialogDescriptionPart1),
                      TextSpan(
                        text: ' ${_emailController.text.trim()} ',
                        style: const TextStyle(fontWeight: FontWeight.bold),
                      ),
                      TextSpan(text: l10n.authOtpDialogDescriptionPart2),
                    ]),
                    style: typography.sm,
                  ),
                  const SizedBox(height: 24),
                  TextField(
                    controller: _otpController,
                    keyboardType: TextInputType.number,
                    maxLength: 6,
                    textAlign: TextAlign.center,
                    inputFormatters: [
                      FilteringTextInputFormatter.digitsOnly,
                    ],
                    style: typography.xl2.copyWith(
                      letterSpacing: 12,
                      fontWeight: FontWeight.bold,
                    ),
                    decoration: InputDecoration(
                      counterText: '',
                      hintText: '------',
                      errorText: otpError.isNotEmpty ? otpError : null,
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(8),
                      ),
                      focusedBorder: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(8),
                        borderSide: BorderSide(color: colors.primary),
                      ),
                    ),
                    onChanged: (value) {
                      if (value.length == 6) {
                        _submitOtp(setDialogState, () {
                          isOtpSubmitting = true;
                        }, (v) {
                          isOtpSubmitting = v;
                        }, (v) {
                          otpError = v;
                        });
                      }
                    },
                  ),
                ],
              ),
              actions: [
                FButton(
                  onPress: isOtpSubmitting
                      ? null
                      : () => _submitOtp(setDialogState, () {
                            isOtpSubmitting = true;
                          }, (v) {
                            isOtpSubmitting = v;
                          }, (v) {
                            otpError = v;
                          }),
                  child: isOtpSubmitting
                      ? const SizedBox(
                          height: 20,
                          width: 20,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : Text(l10n.authVerifyOtp),
                ),
                FButton(
                  variant: FButtonVariant.ghost,
                  onPress: _resendCountdown > 0
                      ? null
                      : () => _resendOtp(setDialogState),
                  child: Text(
                    _resendCountdown > 0
                        ? l10n.authOtpResendCountDown(_resendCountdown)
                        : l10n.authResendOtp,
                  ),
                ),
              ],
            );
          },
        );
      },
    );
  }

  Future<void> _submitOtp(
    StateSetter setDialogState,
    VoidCallback onSubmitStart,
    ValueChanged<bool> setSubmitting,
    ValueChanged<String> setError,
  ) async {
    final otpCode = _otpController.text.trim();
    final l10n = AppLocalizations.of(context);

    if (otpCode.length != 6) {
      setDialogState(() => setError(l10n.errorOtpRequired));
      return;
    }

    setDialogState(() {
      onSubmitStart();
      setError('');
    });

    try {
      final authRepo = ref.read(authRepositoryProvider);
      final response = await authRepo.signup(
        email: _emailController.text.trim(),
        username: _usernameController.text.trim(),
        password: _passwordController.text,
        otpCode: otpCode,
      );

      if (!mounted) return;

      if (response.success) {
        Navigator.of(context, rootNavigator: true).pop();
        await ref.read(currentUserProvider.notifier).fetchMe();
        if (mounted) context.go(AppRoutes.home);
      } else {
        final errorMsg = getLocalizedApiMessage(context, response.key);
        setDialogState(() {
          setSubmitting(false);
          setError(errorMsg);
          _otpController.clear();
        });
      }
    } on DioException catch (_) {
      if (mounted) {
        setDialogState(() {
          setSubmitting(false);
          setError(l10n.fetchError);
        });
      }
    }
  }

  Future<void> _resendOtp(StateSetter setDialogState) async {
    try {
      final authRepo = ref.read(authRepositoryProvider);
      final response = await authRepo.requestVerification(
        email: _emailController.text.trim(),
      );

      if (!mounted) return;

      if (response.success) {
        _startResendCountdown();
        setDialogState(() {});
      } else {
        final errorMsg = getLocalizedApiMessage(context, response.key);
        showFToast(
          context: context,
          title: Text(errorMsg),
          icon: const Icon(FIcons.circleAlert),
        );
      }
    } on DioException catch (_) {
      if (mounted) {
        final l10n = AppLocalizations.of(context);
        showFToast(
          context: context,
          title: Text(l10n.errorOtpSendFail),
          icon: const Icon(FIcons.circleAlert),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;
    final l10n = AppLocalizations.of(context);

    return FScaffold(
      child: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.symmetric(horizontal: 24),
            child: Form(
              key: _formKey,
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Text(
                    l10n.appTitle,
                    style: typography.xl4.copyWith(
                      fontWeight: FontWeight.bold,
                      color: colors.primary,
                    ),
                    textAlign: TextAlign.center,
                  ),
                  const SizedBox(height: 8),
                  Text(
                    l10n.authSignupTitle,
                    style: typography.lg.copyWith(
                      color: colors.mutedForeground,
                    ),
                    textAlign: TextAlign.center,
                  ),
                  const SizedBox(height: 40),
                  FTextFormField(
                    control: FTextFieldControl.managed(
                      controller: _emailController,
                    ),
                    label: Text(l10n.authEmail),
                    hint: 'Enter your email',
                    keyboardType: TextInputType.emailAddress,
                    textInputAction: TextInputAction.next,
                    autovalidateMode: AutovalidateMode.onUserInteraction,
                    validator: (value) {
                      if (value == null || value.isEmpty) {
                        return l10n.errorEmailRequired;
                      }
                      if (value.length > FieldLimits.emailMaxLength) {
                        return l10n.errorEmailTooLong(
                            FieldLimits.emailMaxLength);
                      }
                      final emailRegex =
                          RegExp(r'^[^@\s]+@[^@\s]+\.[^@\s]+$');
                      if (!emailRegex.hasMatch(value)) {
                        return l10n.errorEmailInvalid;
                      }
                      return null;
                    },
                  ),
                  const SizedBox(height: 16),
                  FTextFormField(
                    control: FTextFieldControl.managed(
                      controller: _usernameController,
                    ),
                    label: Text(l10n.username),
                    hint: 'Choose a username',
                    textInputAction: TextInputAction.next,
                    autovalidateMode: AutovalidateMode.onUserInteraction,
                    validator: (value) {
                      if (value == null || value.isEmpty) {
                        return l10n.errorUsernameRequired;
                      }
                      if (value.length < 6) {
                        return l10n.errorUsernameTooShort;
                      }
                      if (value.length > FieldLimits.usernameMaxLength) {
                        return l10n.errorUsernameTooLong(
                            FieldLimits.usernameMaxLength);
                      }
                      return null;
                    },
                  ),
                  const SizedBox(height: 16),
                  FTextFormField.password(
                    control: FTextFieldControl.managed(
                      controller: _passwordController,
                    ),
                    label: Text(l10n.authPassword),
                    hint: 'Create a password',
                    textInputAction: TextInputAction.next,
                    autovalidateMode: AutovalidateMode.onUserInteraction,
                    validator: (value) {
                      if (value == null || value.isEmpty) {
                        return l10n.errorPasswordRequired;
                      }
                      if (value.length < PasswordConstants.minLength) {
                        return l10n.errorPasswordTooShort(
                            PasswordConstants.minLength);
                      }
                      if (value.length > PasswordConstants.maxLength) {
                        return l10n.errorPasswordTooLong(
                            PasswordConstants.maxLength);
                      }
                      if (!RegExp(r'[a-z]').hasMatch(value)) {
                        return l10n.errorPasswordMissingLowercase;
                      }
                      if (!RegExp(r'[A-Z]').hasMatch(value)) {
                        return l10n.errorPasswordMissingUppercase;
                      }
                      if (!RegExp(r'''[!@#$%^&*()\-_+=\[\]{};':"\\|,.<>/?]''')
                          .hasMatch(value)) {
                        return l10n.errorPasswordMissingSpecial;
                      }
                      return null;
                    },
                  ),
                  const SizedBox(height: 16),
                  FTextFormField.password(
                    control: FTextFieldControl.managed(
                      controller: _confirmPasswordController,
                    ),
                    label: Text(l10n.authConfirmPassword),
                    hint: 'Re-enter your password',
                    textInputAction: TextInputAction.done,
                    autovalidateMode: AutovalidateMode.onUserInteraction,
                    validator: (value) {
                      if (value == null || value.isEmpty) {
                        return l10n.errorConfirmPasswordRequired;
                      }
                      if (value != _passwordController.text) {
                        return l10n.errorPasswordsDoNotMatch;
                      }
                      return null;
                    },
                  ),
                  const SizedBox(height: 24),
                  FButton(
                    onPress: _isLoading ? null : _handleSignup,
                    child: _isLoading
                        ? const SizedBox(
                            height: 20,
                            width: 20,
                            child:
                                CircularProgressIndicator(strokeWidth: 2),
                          )
                        : Text(l10n.authSignup),
                  ),
                  const SizedBox(height: 12),
                  Row(
                    children: [
                      Expanded(child: Divider(color: colors.border)),
                      Padding(
                        padding: const EdgeInsets.symmetric(horizontal: 12),
                        child: Text(l10n.or,
                            style: typography.xs
                                .copyWith(color: colors.mutedForeground)),
                      ),
                      Expanded(child: Divider(color: colors.border)),
                    ],
                  ),
                  const SizedBox(height: 12),
                  FButton(
                    variant: FButtonVariant.outline,
                    onPress: _isGoogleLoading ? null : _handleGoogleSignIn,
                    child: _isGoogleLoading
                        ? const SizedBox(
                            height: 20,
                            width: 20,
                            child:
                                CircularProgressIndicator(strokeWidth: 2),
                          )
                        : Row(
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              const Icon(Icons.g_mobiledata, size: 24),
                              const SizedBox(width: 8),
                              Text(l10n.authContinueWithGoogle),
                            ],
                          ),
                  ),
                  const SizedBox(height: 16),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Text(
                        '${l10n.authHaveAccount} ',
                        style: typography.sm,
                      ),
                      GestureDetector(
                        onTap: () => context.go(AppRoutes.login),
                        child: Text(
                          l10n.authLogin,
                          style: typography.sm.copyWith(
                            color: colors.primary,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}
