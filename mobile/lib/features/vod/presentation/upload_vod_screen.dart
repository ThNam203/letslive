import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:forui/forui.dart';
import 'package:image_picker/image_picker.dart';

import '../../../core/constants/field_limits.dart';
import '../../../l10n/app_localizations.dart';
import 'upload_progress_dialog.dart';

class UploadVodScreen extends ConsumerStatefulWidget {
  const UploadVodScreen({super.key});

  @override
  ConsumerState<UploadVodScreen> createState() => _UploadVodScreenState();
}

class _UploadVodScreenState extends ConsumerState<UploadVodScreen> {
  final _titleController = TextEditingController();
  final _descriptionController = TextEditingController();
  final _formKey = GlobalKey<FormState>();
  String _visibility = 'public';
  File? _selectedFile;
  String? _fileName;
  int? _fileSize;

  @override
  void dispose() {
    _titleController.dispose();
    _descriptionController.dispose();
    super.dispose();
  }

  Future<void> _pickVideo() async {
    final picker = ImagePicker();
    final video = await picker.pickVideo(source: ImageSource.gallery);
    if (video == null) return;

    final file = File(video.path);
    final stat = await file.stat();

    setState(() {
      _selectedFile = file;
      _fileName = video.name;
      _fileSize = stat.size;
    });
  }

  void _showUploadProgress() {
    if (_selectedFile == null) return;
    if (!_formKey.currentState!.validate()) return;

    final title = _titleController.text.trim();

    showFDialog(
      context: context,
      builder: (dialogContext, style, animation) {
        return UploadProgressDialog(
          animation: animation,
          videoFile: _selectedFile!,
          title: title,
          description: _descriptionController.text.trim(),
          visibility: _visibility,
        );
      },
    ).then((result) {
      if (result == true && mounted) {
        Navigator.of(context).pop(true);
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final colors = context.theme.colors;
    final typography = context.theme.typography;

    return FScaffold(
      header: FHeader.nested(
        title: Text(l10n.uploadVideoTitle),
      ),
      child: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Form(
          key: _formKey,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              // Video picker
              GestureDetector(
                onTap: _pickVideo,
                child: Container(
                  height: 200,
                  decoration: BoxDecoration(
                    color: colors.muted,
                    borderRadius: BorderRadius.circular(12),
                    border: Border.all(
                      color: colors.border,
                    ),
                  ),
                  child: _selectedFile != null
                      ? Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Icon(
                              FIcons.video,
                              size: 48,
                              color: colors.primary,
                            ),
                            const SizedBox(height: 8),
                            Padding(
                              padding:
                                  const EdgeInsets.symmetric(horizontal: 16),
                              child: Text(
                                _fileName ?? l10n.uploadVideoSelected,
                                style: typography.sm,
                                textAlign: TextAlign.center,
                                maxLines: 2,
                                overflow: TextOverflow.ellipsis,
                              ),
                            ),
                            if (_fileSize != null) ...[
                              const SizedBox(height: 4),
                              Text(
                                _formatBytes(_fileSize!),
                                style: typography.xs.copyWith(
                                  color: colors.mutedForeground,
                                ),
                              ),
                            ],
                            const SizedBox(height: 8),
                            Text(
                              l10n.uploadTapToChange,
                              style: typography.xs.copyWith(
                                color: colors.primary,
                              ),
                            ),
                          ],
                        )
                      : Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Icon(
                              FIcons.cloudUpload,
                              size: 48,
                              color: colors.mutedForeground,
                            ),
                            const SizedBox(height: 8),
                            Text(
                              l10n.uploadSelectVideo,
                              style: typography.base.copyWith(
                                color: colors.mutedForeground,
                              ),
                            ),
                            const SizedBox(height: 4),
                            Text(
                              l10n.uploadSupportedFormats,
                              style: typography.xs.copyWith(
                                color: colors.mutedForeground,
                              ),
                            ),
                          ],
                        ),
                ),
              ),
              const SizedBox(height: 24),

              // Title
              FTextFormField(
                control: FTextFieldControl.managed(
                  controller: _titleController,
                ),
                label: Text(l10n.uploadTitle),
                hint: l10n.uploadTitleHint,
                maxLength: FieldLimits.vodTitleMaxLength,
                autovalidateMode: AutovalidateMode.onUserInteraction,
                validator: (value) {
                  if (value == null || value.trim().isEmpty) {
                    return l10n.errorTitleRequired;
                  }
                  return null;
                },
              ),
              const SizedBox(height: 16),

              // Description
              FTextFormField(
                control: FTextFieldControl.managed(
                  controller: _descriptionController,
                ),
                label: Text(l10n.uploadDescription),
                hint: l10n.uploadDescriptionHint,
                maxLength: FieldLimits.vodDescriptionMaxLength,
                maxLines: 3,
              ),
              const SizedBox(height: 16),

              // Visibility
              Row(
                children: [
                  Text(
                    l10n.settingsVodsVisibility,
                    style:
                        typography.sm.copyWith(fontWeight: FontWeight.w600),
                  ),
                  const Spacer(),
                  if (_visibility == 'public')
                    FButton(
                      onPress: () {},
                      prefix: Icon(FIcons.eye,
                          size: 14, color: colors.primaryForeground),
                      child: Text(l10n.settingsVodsPublic),
                    )
                  else
                    FButton(
                      variant: FButtonVariant.outline,
                      onPress: () {
                        setState(() => _visibility = 'public');
                      },
                      prefix:
                          Icon(FIcons.eye, size: 14, color: colors.primary),
                      child: Text(l10n.settingsVodsPublic),
                    ),
                  const SizedBox(width: 8),
                  if (_visibility == 'private')
                    FButton(
                      onPress: () {},
                      prefix: Icon(FIcons.eyeOff,
                          size: 14, color: colors.primaryForeground),
                      child: Text(l10n.settingsVodsPrivate),
                    )
                  else
                    FButton(
                      variant: FButtonVariant.outline,
                      onPress: () {
                        setState(() => _visibility = 'private');
                      },
                      prefix: Icon(FIcons.eyeOff,
                          size: 14, color: colors.mutedForeground),
                      child: Text(l10n.settingsVodsPrivate),
                    ),
                ],
              ),
              const SizedBox(height: 32),

              // Upload button
              FButton(
                onPress: _selectedFile != null ? _showUploadProgress : null,
                prefix: const Icon(FIcons.upload, size: 16),
                child: Text(l10n.uploadButton),
              ),
            ],
          ),
        ),
      ),
    );
  }

  String _formatBytes(int bytes) {
    if (bytes < 1024) return '$bytes B';
    if (bytes < 1048576) return '${(bytes / 1024).toStringAsFixed(1)} KB';
    if (bytes < 1073741824) {
      return '${(bytes / 1048576).toStringAsFixed(1)} MB';
    }
    return '${(bytes / 1073741824).toStringAsFixed(1)} GB';
  }
}
