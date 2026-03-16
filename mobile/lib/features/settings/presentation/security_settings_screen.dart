import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:forui/forui.dart';

import '../../../core/constants/password.dart';
import '../../../core/utils/api_error_localizer.dart';
import '../../../l10n/app_localizations.dart';
import '../../../models/user.dart';
import '../../../providers.dart';

class SecuritySettingsScreen extends ConsumerStatefulWidget {
  const SecuritySettingsScreen({super.key});

  @override
  ConsumerState<SecuritySettingsScreen> createState() =>
      _SecuritySettingsScreenState();
}

class _SecuritySettingsScreenState
    extends ConsumerState<SecuritySettingsScreen> {
  bool _isGeneratingKey = false;

  Future<void> _generateApiKey() async {
    setState(() => _isGeneratingKey = true);

    try {
      final userRepo = ref.read(userRepositoryProvider);
      final response = await userRepo.generateApiKey();

      if (!mounted) return;

      if (response.success && response.data != null) {
        ref.read(currentUserProvider.notifier).setUser(response.data);
        final l10n = AppLocalizations.of(context);
        showFToast(
          context: context,
          title: Text(l10n.settingsSecurityApiKeyCopied),
          icon: const Icon(FIcons.check),
        );
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
      if (mounted) setState(() => _isGeneratingKey = false);
    }
  }

  void _copyApiKey() {
    final user = ref.read(currentUserProvider);
    if (user?.streamAPIKey != null) {
      Clipboard.setData(ClipboardData(text: user!.streamAPIKey!));
      final l10n = AppLocalizations.of(context);
      showFToast(
        context: context,
        title: Text(l10n.settingsSecurityApiKeyCopied),
        icon: const Icon(FIcons.check),
      );
    }
  }

  void _showChangePasswordDialog() {
    final currentPasswordController = TextEditingController();
    final newPasswordController = TextEditingController();
    final confirmPasswordController = TextEditingController();
    final formKey = GlobalKey<FormState>();
    var isSaving = false;

    showFDialog(
      context: context,
      builder: (dialogContext, style, animation) {
        return StatefulBuilder(
          builder: (context, setDialogState) {
            final l10n = AppLocalizations.of(context);

            Future<void> handleChangePassword() async {
              if (!formKey.currentState!.validate()) return;
              setDialogState(() => isSaving = true);

              try {
                final authRepo = ref.read(authRepositoryProvider);
                final response = await authRepo.changePassword(
                  oldPassword: currentPasswordController.text,
                  newPassword: newPasswordController.text,
                );

                if (!context.mounted) return;

                if (response.success) {
                  showFToast(
                    context: context,
                    title: Text(l10n.settingsSecurityPasswordUpdatedSuccess),
                    icon: const Icon(FIcons.check),
                  );
                  Navigator.of(context).pop();
                } else {
                  final errorMsg =
                      getLocalizedApiMessage(context, response.key);
                  showFToast(
                    context: context,
                    title: Text(errorMsg),
                    icon: const Icon(FIcons.circleAlert),
                  );
                }
              } on DioException catch (_) {
                if (context.mounted) {
                  showFToast(
                    context: context,
                    title: Text(l10n.fetchError),
                    icon: const Icon(FIcons.circleAlert),
                  );
                }
              } finally {
                if (context.mounted) setDialogState(() => isSaving = false);
              }
            }

            return FDialog(
              animation: animation,
              title: Text(l10n.settingsSecurityPasswordDialogTitle),
              body: Form(
                key: formKey,
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    FTextFormField.password(
                      control: FTextFieldControl.managed(
                        controller: currentPasswordController,
                      ),
                      label: Text(
                          l10n.settingsSecurityPasswordFormCurrentLabel),
                      hint: l10n
                          .settingsSecurityPasswordFormCurrentPlaceholder,
                      autovalidateMode: AutovalidateMode.onUserInteraction,
                      textInputAction: TextInputAction.next,
                      validator: (value) {
                        if (value == null || value.isEmpty) {
                          return l10n.errorPasswordRequired;
                        }
                        return null;
                      },
                    ),
                    const SizedBox(height: 12),
                    FTextFormField.password(
                      control: FTextFieldControl.managed(
                        controller: newPasswordController,
                      ),
                      label:
                          Text(l10n.settingsSecurityPasswordFormNewLabel),
                      hint: l10n
                          .settingsSecurityPasswordFormNewPlaceholder,
                      autovalidateMode: AutovalidateMode.onUserInteraction,
                      textInputAction: TextInputAction.next,
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
                        if (value == currentPasswordController.text) {
                          return l10n.errorNewPasswordMustBeDifferent;
                        }
                        return null;
                      },
                    ),
                    const SizedBox(height: 12),
                    FTextFormField.password(
                      control: FTextFieldControl.managed(
                        controller: confirmPasswordController,
                      ),
                      label: Text(
                          l10n.settingsSecurityPasswordFormConfirmLabel),
                      hint: l10n
                          .settingsSecurityPasswordFormConfirmPlaceholder,
                      autovalidateMode: AutovalidateMode.onUserInteraction,
                      textInputAction: TextInputAction.done,
                      validator: (value) {
                        if (value == null || value.isEmpty) {
                          return l10n.errorConfirmPasswordRequired;
                        }
                        if (value != newPasswordController.text) {
                          return l10n.errorPasswordsDoNotMatch;
                        }
                        return null;
                      },
                    ),
                  ],
                ),
              ),
              actions: [
                FButton(
                  variant: FButtonVariant.outline,
                  onPress: () => Navigator.of(context).pop(),
                  child: Text(l10n.cancel),
                ),
                FButton(
                  onPress: isSaving ? null : handleChangePassword,
                  child: isSaving
                      ? const SizedBox(
                          height: 16,
                          width: 16,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : Text(l10n.settingsSecurityPasswordFormSubmit),
                ),
              ],
            );
          },
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;
    final l10n = AppLocalizations.of(context);
    final user = ref.watch(currentUserProvider);

    return FScaffold(
      header: FHeader.nested(
        title: Text(l10n.settingsSecurityTitle),
      ),
      child: ListView(
        children: [
          // Contact section
          _SectionHeader(title: l10n.settingsSecurityContactTitle),
          FTile(
            prefix: const Icon(FIcons.mail),
            title: Text(l10n.settingsSecurityContactEmail),
            subtitle: Text(user?.email ?? ''),
          ),
          FTile(
            prefix: const Icon(FIcons.phone),
            title: Text(l10n.settingsSecurityContactPhone),
            subtitle: Text(user?.phoneNumber ?? l10n.settingsSecurityContactAddPhone),
          ),

          // Password section
          if (user?.authProvider == AuthProvider.local) ...[
            _SectionHeader(title: l10n.settingsSecurityPasswordTitle),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: Text(
                l10n.settingsSecurityPasswordLocalDescription,
                style: typography.sm.copyWith(color: colors.mutedForeground),
              ),
            ),
            const SizedBox(height: 8),
            FTile(
              prefix: const Icon(FIcons.lock),
              title: Text(l10n.settingsSecurityPasswordChange),
              suffix: const Icon(FIcons.chevronRight),
              onPress: _showChangePasswordDialog,
            ),
          ],

          // API Key section
          _SectionHeader(title: l10n.settingsSecurityApiKeyLabel),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            child: DecoratedBox(
              decoration: BoxDecoration(
                color: colors.muted,
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: colors.border, width: 0.5),
              ),
              child: Padding(
                padding:
                    const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                child: Row(
                  children: [
                    Expanded(
                      child: Text(
                        user?.streamAPIKey != null
                            ? '${'*' * 16}${user!.streamAPIKey!.substring(user.streamAPIKey!.length > 4 ? user.streamAPIKey!.length - 4 : 0)}'
                            : '—',
                        style: typography.sm
                            .copyWith(fontFamily: 'monospace'),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                    if (user?.streamAPIKey != null)
                      IconButton(
                        icon: const Icon(FIcons.copy, size: 18),
                        onPressed: _copyApiKey,
                        tooltip: l10n.copyToClipboard,
                      ),
                  ],
                ),
              ),
            ),
          ),
          const SizedBox(height: 12),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            child: FButton(
              variant: FButtonVariant.outline,
              onPress: _isGeneratingKey ? null : _generateApiKey,
              prefix: _isGeneratingKey
                  ? const SizedBox(
                      height: 16,
                      width: 16,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : const Icon(FIcons.refreshCw),
              child: Text(l10n.settingsSecurityApiKeyGenerate),
            ),
          ),
          const SizedBox(height: 24),
        ],
      ),
    );
  }
}

class _SectionHeader extends StatelessWidget {
  final String title;

  const _SectionHeader({required this.title});

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;

    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 24, 16, 8),
      child: Text(
        title,
        style: typography.sm.copyWith(
          color: colors.primary,
          fontWeight: FontWeight.w600,
        ),
      ),
    );
  }
}

