import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/chat_parser/emote.dart';
import 'package:letslive_mobile/core/l10n/api_message.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/core/validation/field_limits.dart';
import 'package:letslive_mobile/features/messages/domain/dm_rules.dart';
import 'package:letslive_mobile/features/messages/messages_providers.dart';
import 'package:letslive_mobile/features/messages/presentation/picture_previews.dart';
import 'package:letslive_mobile/features/settings/domain/image_checks.dart';
import 'package:letslive_mobile/features/settings/settings_providers.dart';
import 'package:letslive_mobile/features/user/user_providers.dart';
import 'package:letslive_mobile/features/watch/chat/presentation/emote_picker.dart';

/// The box under the messages (web `MessageInput`): attach button, emote
/// button, text field and send button.
///
/// - Pictures (up to `FieldLimits.dmImagesMax`, 10 MB each) are chosen from
///   the photo library and shown as removable thumbnails. Send uploads them
///   (`/upload-file`), then sends one picture message; the caption is the
///   typed text, or web's English "Sent an image" / "Sent N images", which
///   the chat service recognises to push a picture notice. A picture that
///   uploaded is not uploaded again when the user retries a failed send; if
///   only some pictures upload, the message goes out with those, like web.
/// - The others see "typing..." from the first keystroke until the user
///   stops for [dmTypingStopDelay] or sends (web's timing).
/// - The text is cleared only after the server accepted the message, and only
///   if the field still holds what was sent. A failure keeps the text and
///   shows the reason with a retry, so a message is never lost.
/// - The field is capped at the server's limit, counted in UTF-16 units
///   (an emoji is 2) like web's `maxLength`.
class MessageComposer extends ConsumerStatefulWidget {
  const MessageComposer({required this.conversationId, super.key});

  final String conversationId;

  @override
  ConsumerState<MessageComposer> createState() => _MessageComposerState();
}

class _MessageComposerState extends ConsumerState<MessageComposer> {
  final _text = TextEditingController();
  Timer? _stopTyping;
  bool _isTyping = false;
  bool _sending = false;
  ApiResponse<Object?>? _failure;

  /// Pictures waiting to be sent; [PendingPicture.path] is set once uploaded.
  List<PendingPicture> _attachments = const [];

  /// A picture-specific problem (too big, upload failed) shown above the box.
  String? _pictureProblem;

  DmThreadNotifier get _thread =>
      ref.read(dmThreadProvider(widget.conversationId).notifier);

  @override
  void dispose() {
    _stopTyping?.cancel();
    _text.dispose();
    super.dispose();
  }

  void _onChanged(String value) {
    if (_failure != null) setState(() => _failure = null);
    if (value.isEmpty) {
      _endTyping();
      return;
    }
    if (!_isTyping) {
      _isTyping = true;
      _thread.setTyping(isTyping: true);
    }
    _stopTyping?.cancel();
    _stopTyping = Timer(dmTypingStopDelay, _endTyping);
  }

  void _endTyping() {
    _stopTyping?.cancel();
    if (!_isTyping) return;
    _isTyping = false;
    _thread.setTyping(isTyping: false);
  }

  Future<void> _pickEmote() async {
    final emote = await showEmotePicker(context);
    if (emote == null || !mounted) return;
    final result = insertEmote(_text.text, _text.selection.baseOffset, emote);
    _text.value = TextEditingValue(
      text: _cap(result.text),
      selection: TextSelection.collapsed(offset: result.cursor),
    );
    _onChanged(_text.text);
  }

  Future<void> _pickPictures() async {
    final l10n = AppLocalizations.of(context);
    final room = FieldLimits.dmImagesMax - _attachments.length;
    if (room <= 0) return;
    final picked = await ref.read(imagesPickProvider)(room);
    if (!mounted || picked.isEmpty) return;
    const maxMb = FieldLimits.generalUploadMaxFileMb;
    String? problem;
    final accepted = <PendingPicture>[];
    for (final image in picked) {
      if (checkFileSize(image, maxMb) != null) {
        problem = l10n.messages_file_exceeds_limit(image.name, maxMb);
      } else {
        accepted.add(PendingPicture(image));
      }
    }
    setState(() {
      _attachments = [..._attachments, ...accepted];
      _pictureProblem = problem;
    });
  }

  void _removePicture(PendingPicture attachment) => setState(() {
    _attachments = [
      for (final a in _attachments)
        if (!identical(a, attachment)) a,
    ];
    _pictureProblem = null;
  });

  /// Uploads the pictures that have no path yet, in parallel (web does the
  /// same); a failed one stays without a path.
  Future<List<PendingPicture>> _uploadPictures() {
    final repo = ref.read(uploadFileRepositoryProvider);
    return Future.wait([
      for (final a in _attachments)
        if (a.path != null)
          Future.value(a)
        else
          repo.upload(a.image).then((r) => PendingPicture(a.image, r.data)),
    ]);
  }

  Future<void> _submit() async {
    final raw = _text.text;
    final text = raw.trim();
    if (_sending || (text.isEmpty && _attachments.isEmpty)) return;
    _endTyping();
    setState(() {
      _sending = true;
      _failure = null;
      _pictureProblem = null;
    });
    if (_attachments.isNotEmpty) return _submitWithPictures(raw, text);
    final response = await _thread.send(text);
    if (!mounted) return;
    setState(() {
      _sending = false;
      if (!response.success) _failure = response;
    });
    if (response.success && _text.text == raw) _text.clear();
  }

  Future<void> _submitWithPictures(String raw, String text) async {
    final l10n = AppLocalizations.of(context);
    final attempted = _attachments.length;
    final uploaded = await _uploadPictures();
    if (!mounted) return;
    final paths = [for (final a in uploaded) ?a.path];
    if (paths.isEmpty) {
      return setState(() {
        _sending = false;
        _pictureProblem = l10n.messages_upload_failed;
      });
    }
    setState(() => _attachments = uploaded);
    final caption = text.isNotEmpty
        ? text
        : paths.length == 1
        ? 'Sent an image'
        : 'Sent ${paths.length} images';
    final response = await _thread.send(caption, imageUrls: paths);
    if (!mounted) return;
    setState(() {
      _sending = false;
      if (!response.success) {
        _failure = response;
      } else {
        _attachments = const [];
        if (paths.length < attempted) {
          _pictureProblem = l10n.messages_upload_some_failed;
        }
      }
    });
    if (response.success && _text.text == raw) _text.clear();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final failure = _failure;
    return SafeArea(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          if (failure != null)
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 12),
              child: Wrap(
                crossAxisAlignment: WrapCrossAlignment.center,
                spacing: 8,
                children: [
                  Text(
                    '${l10n.chat_send_failed} ${apiErrorMessage(l10n, failure)}',
                    style: TextStyle(
                      color: Theme.of(context).colorScheme.error,
                    ),
                  ),
                  TextButton(
                    onPressed: _submit,
                    child: Text(l10n.chat_send_retry),
                  ),
                ],
              ),
            ),
          if (_pictureProblem case final problem?)
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 12),
              child: Text(
                problem,
                style: TextStyle(color: Theme.of(context).colorScheme.error),
              ),
            ),
          if (_attachments.isNotEmpty)
            PicturePreviews(
              pictures: _attachments,
              sending: _sending,
              onRemove: _removePicture,
            ),
          Padding(
            padding: const EdgeInsets.fromLTRB(8, 4, 8, 8),
            child: Row(
              children: [
                IconButton(
                  key: const ValueKey('attach-pictures'),
                  tooltip: l10n.messages_attach_images,
                  icon: const Icon(Icons.attach_file),
                  onPressed: _sending ? null : _pickPictures,
                ),
                IconButton(
                  tooltip: l10n.chat_emotes_action,
                  icon: const Icon(Icons.emoji_emotions_outlined),
                  onPressed: _pickEmote,
                ),
                Expanded(
                  child: TextField(
                    key: const ValueKey('message-input'),
                    controller: _text,
                    inputFormatters: [dmMessageLimitFormatter],
                    minLines: 1,
                    maxLines: 4,
                    textInputAction: TextInputAction.newline,
                    decoration: InputDecoration(
                      hintText: l10n.messages_placeholder_type_message,
                      isDense: true,
                      border: const OutlineInputBorder(),
                    ),
                    onChanged: _onChanged,
                  ),
                ),
                IconButton(
                  key: const ValueKey('send-message'),
                  tooltip: l10n.chat_send_action,
                  icon: const Icon(Icons.send),
                  onPressed: _sending ? null : _submit,
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

/// [text] cut to [FieldLimits.dmMessageMax] UTF-16 units without splitting a
/// surrogate pair.
String _cap(String text) {
  if (text.length <= FieldLimits.dmMessageMax) return text;
  var end = FieldLimits.dmMessageMax;
  final last = text.codeUnitAt(end - 1);
  if (last >= 0xD800 && last <= 0xDBFF) end--;
  return text.substring(0, end);
}

/// Caps a field at the server's DM limit (see `_cap`); shared with the edit
/// dialog so both count UTF-16 units like the server.
final TextInputFormatter dmMessageLimitFormatter =
    TextInputFormatter.withFunction((old, value) {
      final cut = _cap(value.text);
      if (cut == value.text) return value;
      return TextEditingValue(
        text: cut,
        selection: TextSelection.collapsed(offset: cut.length),
      );
    });
