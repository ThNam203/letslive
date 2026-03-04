import 'dart:io';

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

class StreamSettingsScreen extends ConsumerStatefulWidget {
  const StreamSettingsScreen({super.key});

  @override
  ConsumerState<StreamSettingsScreen> createState() =>
      _StreamSettingsScreenState();
}

class _StreamSettingsScreenState extends ConsumerState<StreamSettingsScreen> {
  final _formKey = GlobalKey<FormState>();
  late final TextEditingController _titleController;
  late final TextEditingController _descriptionController;
  bool _isSaving = false;
  String? _newThumbnailPath;

  @override
  void initState() {
    super.initState();
    final user = ref.read(currentUserProvider);
    _titleController =
        TextEditingController(text: user?.livestreamInformation.title ?? '');
    _descriptionController = TextEditingController(
        text: user?.livestreamInformation.description ?? '');
  }

  @override
  void dispose() {
    _titleController.dispose();
    _descriptionController.dispose();
    super.dispose();
  }

  Future<void> _pickThumbnail() async {
    final picker = ImagePicker();
    final image = await picker.pickImage(
      source: ImageSource.gallery,
      maxWidth: 1920,
      maxHeight: 1080,
    );
    if (image != null) {
      setState(() => _newThumbnailPath = image.path);
    }
  }

  Future<void> _save() async {
    if (!_formKey.currentState!.validate()) return;

    setState(() => _isSaving = true);

    try {
      final user = ref.read(currentUserProvider);
      final userRepo = ref.read(userRepositoryProvider);
      final response = await userRepo.updateLivestreamInformation(
        title: _titleController.text.trim(),
        description: _descriptionController.text.trim(),
        thumbnailFilePath: _newThumbnailPath,
        thumbnailUrl: _newThumbnailPath == null
            ? user?.livestreamInformation.thumbnailUrl
            : null,
      );

      if (!mounted) return;

      if (response.success && response.data != null) {
        ref.read(currentUserProvider.notifier).setUser(response.data);
        setState(() => _newThumbnailPath = null);
        final l10n = AppLocalizations.of(context);
        showFToast(
          context: context,
          title: Text(l10n.settingsStreamUpdateSuccess),
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

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;
    final l10n = AppLocalizations.of(context);
    final user = ref.watch(currentUserProvider);
    final existingThumbnail = user?.livestreamInformation.thumbnailUrl;

    return FScaffold(
      header: FHeader.nested(
        title: Text(l10n.settingsStreamTitle),
      ),
      child: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          Text(
            l10n.settingsStreamDescription,
            style: typography.sm.copyWith(color: colors.mutedForeground),
          ),
          const SizedBox(height: 24),

          // Thumbnail preview
          Text(
            l10n.settingsStreamThumbnail,
            style: typography.sm.copyWith(fontWeight: FontWeight.w600),
          ),
          const SizedBox(height: 8),
          GestureDetector(
            onTap: _pickThumbnail,
            child: ClipRRect(
              borderRadius: BorderRadius.circular(12),
              child: AspectRatio(
                aspectRatio: 16 / 9,
                child: Stack(
                  fit: StackFit.expand,
                  children: [
                    if (_newThumbnailPath != null)
                      Image.file(
                        File(_newThumbnailPath!),
                        fit: BoxFit.cover,
                        errorBuilder: (_, _, _) =>
                            ColoredBox(color: colors.muted),
                      )
                    else if (existingThumbnail != null)
                      CachedNetworkImage(
                        imageUrl: '${AppConfig.apiUrl}/$existingThumbnail',
                        fit: BoxFit.cover,
                        errorWidget: (_, _, _) =>
                            ColoredBox(color: colors.muted),
                      )
                    else
                      ColoredBox(color: colors.muted),
                    Container(color: Colors.black.withValues(alpha: 0.2)),
                    Center(
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          const Icon(FIcons.camera,
                              color: Colors.white, size: 32),
                          const SizedBox(height: 8),
                          Text(
                            l10n.settingsStreamChangeThumbnail,
                            style:
                                typography.sm.copyWith(color: Colors.white),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
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
                    controller: _titleController,
                  ),
                  label: Text(l10n.settingsStreamStreamTitle),
                  hint: l10n.settingsStreamStreamTitle,
                  maxLength: FieldLimits.streamTitleMaxLength,
                  textInputAction: TextInputAction.next,
                ),
                const SizedBox(height: 16),
                FTextFormField(
                  control: FTextFieldControl.managed(
                    controller: _descriptionController,
                  ),
                  label: Text(l10n.settingsStreamStreamDescription),
                  hint: l10n.settingsStreamStreamDescription,
                  maxLength: FieldLimits.streamDescriptionMaxLength,
                  maxLines: 4,
                  textInputAction: TextInputAction.newline,
                ),
                const SizedBox(height: 24),
                FButton(
                  onPress: _isSaving ? null : _save,
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
