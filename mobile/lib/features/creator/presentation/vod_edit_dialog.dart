/// The edit dialog of one of the creator's videos (web `VODEditCard`'s
/// dialog): title, description, public switch and a new thumbnail.
///
/// Like web, a new thumbnail is uploaded first (`/upload-file`) and its path
/// saved with the other fields; if either step fails the dialog stays open
/// with everything the user entered, so a retry needs no re-typing.
library;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/l10n/api_message.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/core/validation/field_limits.dart';
import 'package:letslive_mobile/core/widgets/action_feedback.dart';
import 'package:letslive_mobile/features/auth/presentation/auth_widgets.dart';
import 'package:letslive_mobile/features/creator/creator_providers.dart';
import 'package:letslive_mobile/features/home/data/vod.dart';
import 'package:letslive_mobile/features/settings/domain/image_checks.dart';
import 'package:letslive_mobile/features/settings/settings_providers.dart';
import 'package:letslive_mobile/features/user/data/picked_image.dart';
import 'package:letslive_mobile/features/user/user_providers.dart';

/// Opens the edit dialog for [vod]; the list reloads after a save.
Future<void> showVodEditDialog(BuildContext context, Vod vod) =>
    showDialog<void>(
      context: context,
      builder: (context) => _VodEditDialog(vod: vod),
    );

class _VodEditDialog extends ConsumerStatefulWidget {
  const _VodEditDialog({required this.vod});

  final Vod vod;

  @override
  ConsumerState<_VodEditDialog> createState() => _VodEditDialogState();
}

class _VodEditDialogState extends ConsumerState<_VodEditDialog> {
  late final TextEditingController _title;
  late final TextEditingController _description;
  late bool _isPublic;
  PickedImage? _thumbnail;
  bool _busy = false;
  ApiResponse<Object?>? _failure;
  String? _titleError;
  String? _thumbnailError;

  @override
  void initState() {
    super.initState();
    _title = TextEditingController(text: widget.vod.title);
    _description = TextEditingController(text: widget.vod.description ?? '');
    _isPublic = widget.vod.visibility == 'public';
  }

  @override
  void dispose() {
    _title.dispose();
    _description.dispose();
    super.dispose();
  }

  Future<void> _pickThumbnail() async {
    final l10n = AppLocalizations.of(context);
    final image = await pickWithFeedback(context, ref.read(imagePickProvider));
    if (image == null || !mounted) return;
    const maxMb = FieldLimits.generalUploadMaxFileMb;
    setState(() {
      if (checkFileSize(image, maxMb) != null) {
        _thumbnailError = l10n.settings_file_size_exceeds(maxMb);
      } else {
        _thumbnail = image;
        _thumbnailError = null;
      }
    });
  }

  Future<void> _save() async {
    final l10n = AppLocalizations.of(context);
    final title = _title.text.trim();
    if (title.isEmpty) {
      return setState(() => _titleError = l10n.settings_upload_error_no_title);
    }
    setState(() {
      _busy = true;
      _failure = null;
      _titleError = null;
    });
    String? thumbnailPath;
    final thumbnail = _thumbnail;
    if (thumbnail != null) {
      final uploaded = await ref
          .read(uploadFileRepositoryProvider)
          .upload(thumbnail);
      if (!mounted) return;
      if (!uploaded.success) {
        return setState(() {
          _busy = false;
          _failure = uploaded;
        });
      }
      thumbnailPath = uploaded.data;
    }
    final saved = await ref
        .read(myVodsRepositoryProvider)
        .update(
          widget.vod.id,
          title: title,
          description: _description.text,
          isPublic: _isPublic,
          thumbnailPath: thumbnailPath,
        );
    if (!mounted) return;
    if (!saved.success) {
      return setState(() {
        _busy = false;
        _failure = saved;
      });
    }
    ref.invalidate(myVodsProvider);
    final messenger = ScaffoldMessenger.of(context);
    Navigator.pop(context);
    messenger.showSnackBar(
      SnackBar(content: Text(l10n.settings_vods_edit_dialog_update_success)),
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final failure = _failure;
    final thumbnail = _thumbnail;
    return AlertDialog(
      title: Text(l10n.settings_vods_edit_dialog_title),
      content: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(l10n.settings_vods_edit_dialog_description),
            const SizedBox(height: 12),
            if (thumbnail != null)
              AspectRatio(
                aspectRatio: 16 / 9,
                child: Image.memory(thumbnail.bytes, fit: BoxFit.cover),
              ),
            OutlinedButton.icon(
              key: const ValueKey('pick-vod-thumbnail'),
              onPressed: _busy ? null : _pickThumbnail,
              icon: const Icon(Icons.image_outlined),
              label: Text(l10n.settings_vods_edit_dialog_change_thumbnail),
            ),
            if (_thumbnailError != null) FormErrorText(_thumbnailError),
            TextField(
              key: const ValueKey('vod-title'),
              controller: _title,
              enabled: !_busy,
              maxLength: FieldLimits.vodTitleMax,
              decoration: InputDecoration(
                labelText: l10n.settings_vods_edit_dialog_title_label,
                errorText: _titleError,
              ),
            ),
            TextField(
              key: const ValueKey('vod-description'),
              controller: _description,
              enabled: !_busy,
              maxLength: FieldLimits.vodDescriptionMax,
              minLines: 2,
              maxLines: 5,
              decoration: InputDecoration(
                labelText: l10n.settings_vods_edit_dialog_description_label,
              ),
            ),
            SwitchListTile(
              key: const ValueKey('vod-public'),
              contentPadding: EdgeInsets.zero,
              title: Text(l10n.settings_vods_edit_dialog_public),
              value: _isPublic,
              onChanged: _busy ? null : (v) => setState(() => _isPublic = v),
            ),
            if (failure != null) FormErrorText(apiErrorMessage(l10n, failure)),
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: _busy ? null : () => Navigator.pop(context),
          child: Text(l10n.settings_vods_edit_dialog_cancel),
        ),
        SubmitButton(
          key: const ValueKey('save-vod'),
          expanded: false,
          busy: _busy,
          label: l10n.settings_vods_edit_dialog_save_changes,
          onPressed: _save,
        ),
      ],
    );
  }
}
