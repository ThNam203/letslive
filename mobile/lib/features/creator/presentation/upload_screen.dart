/// Upload a video (web `/settings/upload` plus its upload manager panel).
///
/// The form checks the file (format, 100 MB) before anything is sent, then
/// hands it to the app-wide `UploadManager`, so leaving this screen does not
/// cancel the upload. The list under the form shows every upload of this
/// session: progress, "processing on the server", done, or failed with Retry.
/// Uploads need the app in the foreground: keep it open while sending.
library;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/l10n/api_message.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/validation/field_limits.dart';
import 'package:letslive_mobile/core/widgets/action_feedback.dart';
import 'package:letslive_mobile/features/auth/presentation/auth_widgets.dart';
import 'package:letslive_mobile/features/creator/creator_providers.dart';
import 'package:letslive_mobile/features/creator/domain/format_bytes.dart';
import 'package:letslive_mobile/features/creator/domain/picked_video.dart';
import 'package:letslive_mobile/features/creator/upload_manager.dart';

/// The upload screen.
class UploadScreen extends ConsumerStatefulWidget {
  const UploadScreen({super.key});

  @override
  ConsumerState<UploadScreen> createState() => _UploadScreenState();
}

class _UploadScreenState extends ConsumerState<UploadScreen> {
  final _title = TextEditingController();
  final _description = TextEditingController();
  PickedVideo? _video;
  bool _isPublic = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _title.addListener(() => setState(() {}));
  }

  @override
  void dispose() {
    _title.dispose();
    _description.dispose();
    super.dispose();
  }

  Future<void> _pick() async {
    final l10n = AppLocalizations.of(context);
    final video = await pickWithFeedback(context, ref.read(videoPickProvider));
    if (video == null || !mounted) return;
    setState(() {
      switch (checkVideo(video)) {
        case VideoProblem.unsupportedFormat:
          _error = l10n.settings_upload_error_invalid_format;
        case VideoProblem.tooLarge:
          _error = l10n.settings_upload_error_file_too_large;
        case null:
          _error = null;
          _video = video;
          if (_title.text.isEmpty) _title.text = suggestedTitle(video.name);
      }
    });
  }

  void _submit() {
    final video = _video;
    if (video == null || _title.text.trim().isEmpty) return;
    ref
        .read(uploadManagerProvider.notifier)
        .enqueue(
          video,
          title: _title.text.trim(),
          description: _description.text.trim(),
          isPublic: _isPublic,
        );
    // Ready for the next video; this one is now in the list below.
    setState(() {
      _video = null;
      _title.clear();
      _description.clear();
      _isPublic = true;
    });
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final theme = Theme.of(context);
    final video = _video;
    final uploads = ref.watch(uploadManagerProvider);
    final queued = uploads.where((i) => i.status == UploadStatus.queued).length;
    return Scaffold(
      appBar: AppBar(title: Text(l10n.settings_upload_title)),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          Text(l10n.settings_upload_description),
          const SizedBox(height: 12),
          OutlinedButton.icon(
            key: const ValueKey('pick-video'),
            onPressed: _pick,
            icon: const Icon(Icons.video_file_outlined),
            label: Text(
              video == null
                  ? l10n.settings_upload_select_video
                  : l10n.settings_upload_change_video,
            ),
          ),
          if (video != null)
            Padding(
              padding: const EdgeInsets.only(top: 8),
              child: Text(
                l10n.settings_upload_selected_file(
                  video.name,
                  formatBytes(video.size),
                ),
                key: const ValueKey('selected-video'),
              ),
            )
          else
            Text(
              l10n.settings_upload_select_video_description,
              style: theme.textTheme.bodySmall,
            ),
          FormErrorText(_error),
          const SizedBox(height: 12),
          TextField(
            key: const ValueKey('upload-title'),
            controller: _title,
            maxLength: FieldLimits.vodTitleMax,
            decoration: InputDecoration(
              labelText: l10n.settings_upload_title_label,
              helperText: l10n.settings_upload_title_description,
              helperMaxLines: 3,
            ),
          ),
          TextField(
            key: const ValueKey('upload-description'),
            controller: _description,
            maxLength: FieldLimits.vodDescriptionMax,
            minLines: 3,
            maxLines: 6,
            decoration: InputDecoration(
              labelText: l10n.settings_upload_description_label,
            ),
          ),
          SwitchListTile(
            key: const ValueKey('upload-public'),
            contentPadding: EdgeInsets.zero,
            title: Text(l10n.settings_upload_visibility_label),
            value: _isPublic,
            onChanged: (value) => setState(() => _isPublic = value),
          ),
          Text(
            l10n.settings_upload_note_max_time,
            style: theme.textTheme.bodySmall,
          ),
          Text(
            l10n.settings_upload_keep_open_note,
            style: theme.textTheme.bodySmall,
          ),
          if (queued >= FieldLimits.vodUploadConcurrency)
            Text(
              l10n.settings_upload_error_max_concurrent(
                FieldLimits.vodUploadConcurrency,
              ),
              style: theme.textTheme.bodySmall,
            ),
          const SizedBox(height: 12),
          SubmitButton(
            key: const ValueKey('upload-submit'),
            label: l10n.settings_upload_upload_button,
            busy: false,
            onPressed: video != null && _title.text.trim().isNotEmpty
                ? _submit
                : null,
          ),
          if (uploads.isNotEmpty) ...[
            const Divider(height: 32),
            _UploadList(uploads: uploads),
          ],
        ],
      ),
    );
  }
}

/// The uploads of this session with a summary header (web's manager panel).
class _UploadList extends ConsumerWidget {
  const _UploadList({required this.uploads});

  final List<UploadItem> uploads;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    int count(bool Function(UploadStatus) test) =>
        uploads.where((i) => test(i.status)).length;
    final active = count(
      (s) => s == UploadStatus.queued || s == UploadStatus.uploading,
    );
    final failed = count((s) => s == UploadStatus.failed);
    final completed = count((s) => s == UploadStatus.completed);
    final header = active > 0
        ? l10n.settings_upload_manager_header_uploading(active)
        : failed > 0
        ? l10n.settings_upload_manager_header_failed(failed)
        : l10n.settings_upload_manager_header_complete(completed);
    final hasFinished = failed + completed > 0;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Expanded(
              child: Text(
                header,
                style: Theme.of(context).textTheme.titleMedium,
              ),
            ),
            if (hasFinished)
              TextButton(
                key: const ValueKey('clear-uploads'),
                onPressed: ref
                    .read(uploadManagerProvider.notifier)
                    .dismissFinished,
                child: Text(l10n.settings_upload_manager_clear),
              ),
          ],
        ),
        for (final item in uploads) _UploadTile(item: item),
      ],
    );
  }
}

class _UploadTile extends ConsumerWidget {
  const _UploadTile({required this.item});

  final UploadItem item;

  String _failureText(AppLocalizations l10n) {
    final failure = item.failure;
    return switch (failure?.kind) {
      UploadFailureKind.request when failure?.response != null =>
        apiErrorMessage(l10n, failure!.response!),
      UploadFailureKind.fileMissing => l10n.settings_upload_error_no_file,
      UploadFailureKind.processing => l10n.settings_upload_processing_failed,
      _ => l10n.settings_upload_manager_failed,
    };
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final theme = Theme.of(context);
    final manager = ref.read(uploadManagerProvider.notifier);
    final isActive =
        item.status == UploadStatus.queued ||
        item.status == UploadStatus.uploading;
    final line = switch (item.status) {
      UploadStatus.queued => l10n.settings_upload_manager_waiting,
      UploadStatus.uploading =>
        '${formatBytes(item.sent)} / ${formatBytes(item.total)} '
            '(${item.percent}%)',
      UploadStatus.processing => l10n.settings_upload_manager_processing,
      UploadStatus.completed => l10n.settings_upload_manager_complete,
      UploadStatus.failed => _failureText(l10n),
    };
    return ListTile(
      key: ValueKey('upload-${item.id}'),
      contentPadding: EdgeInsets.zero,
      title: Text(item.title, maxLines: 2, overflow: TextOverflow.ellipsis),
      subtitle: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (item.status == UploadStatus.uploading)
            LinearProgressIndicator(value: item.percent / 100),
          Text(
            line,
            style: item.status == UploadStatus.failed
                ? TextStyle(color: theme.colorScheme.error)
                : null,
          ),
        ],
      ),
      trailing: Wrap(
        crossAxisAlignment: WrapCrossAlignment.center,
        children: [
          if (item.status == UploadStatus.failed)
            TextButton(
              key: ValueKey('retry-${item.id}'),
              onPressed: () => manager.retry(item.id),
              child: Text(l10n.settings_upload_manager_retry),
            ),
          IconButton(
            key: ValueKey('close-${item.id}'),
            tooltip: isActive
                ? l10n.settings_upload_manager_cancel
                : l10n.settings_upload_manager_dismiss,
            icon: const Icon(Icons.close),
            onPressed: () =>
                isActive ? manager.cancel(item.id) : manager.dismiss(item.id),
          ),
        ],
      ),
    );
  }
}
