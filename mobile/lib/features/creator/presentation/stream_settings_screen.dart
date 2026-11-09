/// Stream settings (web `/settings/stream` and the API key tab of
/// `/settings/security`): the stream key and the information shown for the
/// next livestream.
///
/// The stream key is a secret. It is masked until the user reveals it, can be
/// copied, and is never logged or put in a message. Generating a new key asks
/// for confirmation first because the old key stops working at once.
library;

import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/l10n/api_message.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/validation/field_limits.dart';
import 'package:letslive_mobile/core/widgets/action_feedback.dart';
import 'package:letslive_mobile/core/widgets/feed_section.dart';
import 'package:letslive_mobile/features/auth/presentation/auth_widgets.dart';
import 'package:letslive_mobile/features/creator/creator_providers.dart';
import 'package:letslive_mobile/features/creator/data/stream_settings_repository.dart';
import 'package:letslive_mobile/features/settings/domain/image_checks.dart';
import 'package:letslive_mobile/features/settings/settings_providers.dart';
import 'package:letslive_mobile/features/user/data/picked_image.dart';

/// The stream settings screen.
class StreamSettingsScreen extends ConsumerWidget {
  const StreamSettingsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final settings = ref.watch(streamSettingsProvider);
    final body = feedStatusBody<StreamSettings>(
      value: settings.whenData((s) => [s]),
      onRetry: () => ref.invalidate(streamSettingsProvider),
    );
    return Scaffold(
      appBar: AppBar(title: Text(l10n.settings_navigation_stream)),
      body: body != null
          ? SingleChildScrollView(child: body)
          : ListView(
              padding: const EdgeInsets.all(16),
              children: [
                const StreamKeyCard(),
                const SizedBox(height: 24),
                StreamInfoForm(settings: settings.requireValue),
              ],
            ),
    );
  }
}

/// The masked stream key with show, copy and generate-new buttons.
class StreamKeyCard extends ConsumerStatefulWidget {
  const StreamKeyCard({super.key});

  @override
  ConsumerState<StreamKeyCard> createState() => _StreamKeyCardState();
}

class _StreamKeyCardState extends ConsumerState<StreamKeyCard> {
  bool _revealed = false;
  bool _busy = false;

  void _say(String message) {
    if (!mounted) return;
    ScaffoldMessenger.of(context)
        .showSnackBar(SnackBar(content: Text(message)));
  }

  Future<void> _copy(String key) async {
    final copied = AppLocalizations.of(context)
        .settings_security_api_key_copied;
    await Clipboard.setData(ClipboardData(text: key));
    _say(copied);
  }

  Future<void> _reset() async {
    final l10n = AppLocalizations.of(context);
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(l10n.settings_stream_key_reset_confirm_title),
        content: Text(l10n.settings_stream_key_reset_confirm_description),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(l10n.common_cancel),
          ),
          FilledButton(
            key: const ValueKey('confirm-reset-key'),
            onPressed: () => Navigator.pop(context, true),
            child: Text(l10n.settings_stream_key_reset_confirm_action),
          ),
        ],
      ),
    );
    if (confirmed != true || !mounted) return;
    setState(() => _busy = true);
    final failure = await ref.read(streamSettingsProvider.notifier).resetKey();
    if (!mounted) return;
    setState(() {
      _busy = false;
      // A new key starts hidden.
      if (failure == null) _revealed = false;
    });
    if (failure != null) _say(apiErrorMessage(l10n, failure));
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final theme = Theme.of(context);
    final key = ref.watch(streamSettingsProvider).requireValue.streamKey;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          l10n.settings_stream_key_title,
          style: theme.textTheme.titleMedium,
        ),
        Text(l10n.settings_stream_key_description),
        const SizedBox(height: 12),
        Row(
          children: [
            Expanded(
              child: Text(
                _revealed ? key : '•' * key.length,
                key: const ValueKey('stream-key-value'),
                style: const TextStyle(fontFamily: 'monospace'),
                // A long key wraps instead of overflowing.
                softWrap: true,
              ),
            ),
            IconButton(
              key: const ValueKey('toggle-key'),
              tooltip: _revealed
                  ? l10n.settings_stream_key_hide
                  : l10n.settings_stream_key_show,
              icon: Icon(_revealed ? Icons.visibility_off : Icons.visibility),
              onPressed: () => setState(() => _revealed = !_revealed),
            ),
            IconButton(
              key: const ValueKey('copy-key'),
              tooltip: l10n.settings_stream_key_copy,
              icon: const Icon(Icons.copy),
              onPressed: () => _copy(key),
            ),
          ],
        ),
        const SizedBox(height: 8),
        OutlinedButton.icon(
          key: const ValueKey('reset-key'),
          onPressed: _busy ? null : _reset,
          icon: _busy
              ? const SizedBox.square(
                  dimension: 16,
                  child: CircularProgressIndicator(strokeWidth: 2),
                )
              : const Icon(Icons.refresh),
          label: Text(l10n.settings_security_api_key_generate),
        ),
      ],
    );
  }
}

/// Title, description and thumbnail of the next livestream. Nothing is sent
/// until Confirm, like web.
class StreamInfoForm extends ConsumerStatefulWidget {
  const StreamInfoForm({required this.settings, super.key});

  /// What the server has saved.
  final StreamSettings settings;

  @override
  ConsumerState<StreamInfoForm> createState() => _StreamInfoFormState();
}

class _StreamInfoFormState extends ConsumerState<StreamInfoForm> {
  late final TextEditingController _title;
  late final TextEditingController _description;
  PickedImage? _pending;
  bool _thumbnailRemoved = false;
  double? _progress;
  bool _busy = false;

  @override
  void initState() {
    super.initState();
    _title = TextEditingController(text: widget.settings.title)
      ..addListener(_rebuild);
    _description = TextEditingController(text: widget.settings.description)
      ..addListener(_rebuild);
  }

  void _rebuild() => setState(() {});

  @override
  void dispose() {
    _title.dispose();
    _description.dispose();
    super.dispose();
  }

  bool get _hasChanges =>
      _title.text != widget.settings.title ||
      _description.text != widget.settings.description ||
      _pending != null ||
      _thumbnailRemoved;

  void _say(String message) {
    if (!mounted) return;
    ScaffoldMessenger.of(context)
        .showSnackBar(SnackBar(content: Text(message)));
  }

  Future<void> _pick() async {
    final l10n = AppLocalizations.of(context);
    final image = await pickWithFeedback(context, ref.read(imagePickProvider));
    if (image == null || !mounted) return;
    const maxMb = FieldLimits.livestreamThumbnailMaxFileMb;
    if (checkFileSize(image, maxMb) != null) {
      return _say(l10n.settings_file_size_exceeds(maxMb));
    }
    setState(() {
      _pending = image;
      _thumbnailRemoved = false;
    });
  }

  Future<void> _save() async {
    final l10n = AppLocalizations.of(context);
    setState(() => _busy = true);
    final response = await ref
        .read(streamSettingsRepositoryProvider)
        .update(
          title: _title.text,
          description: _description.text,
          thumbnail: _pending,
          keepThumbnailUrl: _thumbnailRemoved
              ? null
              : widget.settings.thumbnailUrl,
          onProgress: (sent, total) {
            if (mounted && total > 0) setState(() => _progress = sent / total);
          },
        );
    if (!mounted) return;
    final info = response.data;
    setState(() {
      _busy = false;
      _progress = null;
      if (response.success && info != null) {
        _pending = null;
        _thumbnailRemoved = false;
      }
    });
    if (response.success && info != null) {
      ref.read(streamSettingsProvider.notifier).applyInfo(info);
      _say(l10n.settings_stream_updated_success);
    } else {
      _say(apiErrorMessage(l10n, response));
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final theme = Theme.of(context);
    final saved = widget.settings.thumbnailUrl;
    final pending = _pending;
    final progress = _progress;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(l10n.settings_stream_title, style: theme.textTheme.titleMedium),
        Text(l10n.settings_stream_description),
        const SizedBox(height: 12),
        TextField(
          key: const ValueKey('stream-title'),
          controller: _title,
          enabled: !_busy,
          maxLength: FieldLimits.streamTitleMax,
          decoration: InputDecoration(
            labelText: l10n.settings_stream_title_label,
            helperText: l10n.settings_stream_title_description,
            helperMaxLines: 3,
          ),
        ),
        const SizedBox(height: 12),
        TextField(
          key: const ValueKey('stream-description'),
          controller: _description,
          enabled: !_busy,
          maxLength: FieldLimits.streamDescriptionMax,
          minLines: 3,
          maxLines: 6,
          decoration: InputDecoration(
            labelText: l10n.settings_stream_description_label,
          ),
        ),
        const SizedBox(height: 12),
        Text(l10n.settings_stream_thumbnail_label),
        Text(
          l10n.settings_stream_thumbnail_description,
          style: theme.textTheme.bodySmall,
        ),
        const SizedBox(height: 8),
        AspectRatio(
          aspectRatio: 16 / 9,
          child: ClipRRect(
            borderRadius: BorderRadius.circular(8),
            child: ColoredBox(
              color: theme.colorScheme.surfaceContainerHighest,
              child: pending != null
                  ? Image.memory(pending.bytes, fit: BoxFit.cover)
                  : saved == null || _thumbnailRemoved
                  ? null
                  : CachedNetworkImage(
                      imageUrl: saved,
                      fit: BoxFit.cover,
                      errorWidget: (_, _, _) => const SizedBox.shrink(),
                    ),
            ),
          ),
        ),
        Wrap(
          spacing: 8,
          children: [
            OutlinedButton.icon(
              key: const ValueKey('pick-thumbnail'),
              onPressed: _busy ? null : _pick,
              icon: const Icon(Icons.image_outlined),
              label: Text(l10n.settings_stream_thumbnail_hover),
            ),
            if (pending != null || (saved != null && !_thumbnailRemoved))
              TextButton.icon(
                key: const ValueKey('remove-thumbnail'),
                onPressed: _busy
                    ? null
                    : () => setState(() {
                        _pending = null;
                        _thumbnailRemoved = true;
                      }),
                icon: const Icon(Icons.close),
                label: Text(l10n.settings_stream_thumbnail_remove),
              ),
          ],
        ),
        const SizedBox(height: 16),
        if (progress != null) ...[
          Text(l10n.settings_upload_uploading),
          LinearProgressIndicator(value: progress),
          const SizedBox(height: 16),
        ],
        SubmitButton(
          key: const ValueKey('stream-save'),
          label: l10n.settings_stream_confirm_edit_button,
          busy: _busy,
          onPressed: _hasChanges ? _save : null,
        ),
      ],
    );
  }
}
