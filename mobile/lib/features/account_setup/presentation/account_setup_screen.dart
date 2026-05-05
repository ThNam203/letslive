import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:forui/forui.dart';
import 'package:go_router/go_router.dart';

import '../../../core/constants/field_limits.dart';
import '../../../core/utils/api_error_localizer.dart';
import '../../../l10n/app_localizations.dart';
import '../../../providers.dart';

class AccountSetupScreen extends ConsumerStatefulWidget {
  const AccountSetupScreen({super.key});

  @override
  ConsumerState<AccountSetupScreen> createState() =>
      _AccountSetupScreenState();
}

class _AccountSetupScreenState extends ConsumerState<AccountSetupScreen> {
  final _formKey = GlobalKey<FormState>();
  final _usernameController = TextEditingController();
  bool _isSaving = false;

  @override
  void dispose() {
    _usernameController.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() => _isSaving = true);

    try {
      final userRepo = ref.read(userRepositoryProvider);
      final response = await userRepo.updateProfile(
        username: _usernameController.text.trim(),
      );

      if (!mounted) return;

      if (response.success && response.data != null) {
        ref.read(currentUserProvider.notifier).setUser(response.data);
        context.go('/');
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
      if (mounted) setState(() => _isSaving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final typography = context.theme.typography;
    final l10n = AppLocalizations.of(context);

    return FScaffold(
      child: SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 32),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Text(
                l10n.accountSetupTitle,
                style: typography.xl2.copyWith(fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 8),
              Text(
                l10n.accountSetupSubtitle,
                style: typography.sm.copyWith(
                  color: context.theme.colors.mutedForeground,
                ),
              ),
              const SizedBox(height: 32),
              Form(
                key: _formKey,
                child: FTextFormField(
                  control: FTextFieldControl.managed(
                    controller: _usernameController,
                  ),
                  label: Text(l10n.settingsProfileUsername),
                  hint: l10n.authUsernameHint,
                  maxLength: FieldLimits.usernameMaxLength,
                  textInputAction: TextInputAction.done,
                  validator: (value) {
                    final trimmed = value?.trim() ?? '';
                    if (trimmed.isEmpty) return l10n.errorUsernameRequired;
                    if (trimmed.length < FieldLimits.usernameMinLength) {
                      return l10n.errorUsernameTooShort;
                    }
                    return null;
                  },
                ),
              ),
              const SizedBox(height: 24),
              FButton(
                onPress: _isSaving ? null : _submit,
                child: _isSaving
                    ? const SizedBox(
                        height: 20,
                        width: 20,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : Text(l10n.accountSetupSubmit),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
