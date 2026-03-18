import 'package:cached_network_image/cached_network_image.dart';
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:forui/forui.dart';
import 'package:image_picker/image_picker.dart';

import '../../../core/config/app_config.dart';
import '../../../core/constants/field_limits.dart';
import '../../../core/utils/api_error_localizer.dart';
import '../../../l10n/app_localizations.dart';
import '../../../providers.dart';

class ProfileSettingsScreen extends ConsumerStatefulWidget {
  const ProfileSettingsScreen({super.key});

  @override
  ConsumerState<ProfileSettingsScreen> createState() =>
      _ProfileSettingsScreenState();
}

class _ProfileSettingsScreenState extends ConsumerState<ProfileSettingsScreen> {
  final _formKey = GlobalKey<FormState>();
  late final TextEditingController _displayNameController;
  late final TextEditingController _bioController;
  bool _isSaving = false;
  bool _isUploadingProfile = false;
  bool _isUploadingBackground = false;

  @override
  void initState() {
    super.initState();
    final user = ref.read(currentUserProvider);
    _displayNameController = TextEditingController(
      text: user?.displayName ?? '',
    );
    _bioController = TextEditingController(text: user?.bio ?? '');
  }

  @override
  void dispose() {
    _displayNameController.dispose();
    _bioController.dispose();
    super.dispose();
  }

  Future<void> _saveProfile() async {
    if (!_formKey.currentState!.validate()) return;

    setState(() => _isSaving = true);

    try {
      final userRepo = ref.read(userRepositoryProvider);
      final response = await userRepo.updateProfile(
        displayName: _displayNameController.text.trim(),
        bio: _bioController.text.trim(),
      );

      if (!mounted) return;

      if (response.success && response.data != null) {
        ref.read(currentUserProvider.notifier).setUser(response.data);
        final l10n = AppLocalizations.of(context);
        showFToast(
          context: context,
          title: Text(l10n.settingsProfileUpdateSuccess),
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
      if (mounted) setState(() => _isSaving = false);
    }
  }

  Future<void> _pickAndUploadProfilePicture() async {
    final picker = ImagePicker();
    final image = await picker.pickImage(
      source: ImageSource.gallery,
      maxWidth: 1024,
      maxHeight: 1024,
    );
    if (image == null) return;

    setState(() => _isUploadingProfile = true);

    try {
      final userRepo = ref.read(userRepositoryProvider);
      final response = await userRepo.updateProfilePicture(image.path);

      if (!mounted) return;

      if (response.success && response.data != null) {
        ref.read(currentUserProvider.notifier).setUser(response.data);
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
      if (mounted) setState(() => _isUploadingProfile = false);
    }
  }

  Future<void> _pickAndUploadBackgroundPicture() async {
    final picker = ImagePicker();
    final image = await picker.pickImage(
      source: ImageSource.gallery,
      maxWidth: 2048,
      maxHeight: 1024,
    );
    if (image == null) return;

    setState(() => _isUploadingBackground = true);

    try {
      final userRepo = ref.read(userRepositoryProvider);
      final response = await userRepo.updateBackgroundPicture(image.path);

      if (!mounted) return;

      if (response.success && response.data != null) {
        ref.read(currentUserProvider.notifier).setUser(response.data);
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
      if (mounted) setState(() => _isUploadingBackground = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;
    final l10n = AppLocalizations.of(context);
    final user = ref.watch(currentUserProvider);

    return FScaffold(
      header: FHeader.nested(title: Text(l10n.settingsProfileTitle)),
      child: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          // Background picture
          GestureDetector(
            onTap: _isUploadingBackground
                ? null
                : _pickAndUploadBackgroundPicture,
            child: ClipRRect(
              borderRadius: BorderRadius.circular(12),
              child: SizedBox(
                height: 120,
                width: double.infinity,
                child: Stack(
                  fit: StackFit.expand,
                  children: [
                    if (user?.backgroundPicture != null)
                      CachedNetworkImage(
                        imageUrl:
                            '${AppConfig.apiUrl}/${user!.backgroundPicture}',
                        fit: BoxFit.cover,
                        errorWidget: (_, _, _) =>
                            ColoredBox(color: colors.muted),
                      )
                    else
                      ColoredBox(color: colors.muted),
                    Container(color: Colors.black.withValues(alpha: 0.3)),
                    Center(
                      child: _isUploadingBackground
                          ? const CircularProgressIndicator(strokeWidth: 2)
                          : Column(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                const Icon(
                                  FIcons.camera,
                                  color: Colors.white,
                                  size: 24,
                                ),
                                const SizedBox(height: 4),
                                Text(
                                  l10n.settingsProfileUpdateBackground,
                                  style: typography.xs.copyWith(
                                    color: Colors.white,
                                  ),
                                ),
                              ],
                            ),
                    ),
                  ],
                ),
              ),
            ),
          ),
          const SizedBox(height: 16),

          // Profile picture
          Center(
            child: GestureDetector(
              onTap: _isUploadingProfile ? null : _pickAndUploadProfilePicture,
              child: Stack(
                children: [
                  CircleAvatar(
                    radius: 48,
                    backgroundImage: user?.profilePicture != null
                        ? CachedNetworkImageProvider(
                            '${AppConfig.apiUrl}/${user!.profilePicture}',
                          )
                        : null,
                    child: user?.profilePicture == null
                        ? const Icon(FIcons.user, size: 32)
                        : null,
                  ),
                  if (_isUploadingProfile)
                    const Positioned.fill(
                      child: Center(
                        child: CircularProgressIndicator(strokeWidth: 2),
                      ),
                    )
                  else
                    Positioned(
                      bottom: 0,
                      right: 0,
                      child: Container(
                        padding: const EdgeInsets.all(4),
                        decoration: BoxDecoration(
                          color: colors.primary,
                          shape: BoxShape.circle,
                        ),
                        child: Icon(
                          FIcons.camera,
                          size: 16,
                          color: colors.primaryForeground,
                        ),
                      ),
                    ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 24),

          // Form fields
          Form(
            key: _formKey,
            child: Column(
              children: [
                FTextFormField(
                  control: FTextFieldControl.managed(
                    controller: _displayNameController,
                  ),
                  label: Text(l10n.settingsProfileDisplayName),
                  hint: l10n.settingsProfileDisplayName,
                  maxLength: FieldLimits.displayNameMaxLength,
                  textInputAction: TextInputAction.next,
                ),
                const SizedBox(height: 16),
                FTextFormField(
                  control: FTextFieldControl.managed(
                    controller: _bioController,
                  ),
                  label: Text(l10n.settingsProfileBio),
                  hint: l10n.settingsProfileBio,
                  maxLength: FieldLimits.bioMaxLength,
                  maxLines: 4,
                  textInputAction: TextInputAction.newline,
                ),
                const SizedBox(height: 24),
                FButton(
                  onPress: _isSaving ? null : _saveProfile,
                  child: _isSaving
                      ? const SizedBox(
                          height: 20,
                          width: 20,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : Text(l10n.settingsSaveChanges),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
