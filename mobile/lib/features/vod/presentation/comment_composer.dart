import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';

/// Longest comment the server accepts (web `VOD_COMMENT_MAX_LENGTH`).
const int commentMaxLength = 2000;

/// Caps the text at [commentMaxLength] UTF-16 units, like web's
/// `maxLength` (an emoji counts 2). Pasted text is cut, not refused, and a
/// cut never splits an emoji.
final TextInputFormatter _lengthLimit = TextInputFormatter.withFunction((
  old,
  value,
) {
  final text = value.text;
  if (text.length <= commentMaxLength) return value;
  var end = commentMaxLength;
  final lastUnit = text.codeUnitAt(end - 1);
  if (lastUnit >= 0xD800 && lastUnit <= 0xDBFF) end--;
  return TextEditingValue(
    text: text.substring(0, end),
    selection: TextSelection.collapsed(offset: end),
  );
});

/// A box to write or edit a comment, with the characters left and a submit
/// button (web `CommentForm` / `CommentEditForm`).
///
/// [onSubmit] gets the trimmed text and returns whether it worked; the box
/// stays filled when it did not so nothing typed is lost. After a success a
/// new-comment box ([initialText] empty) clears itself.
class CommentComposer extends StatefulWidget {
  const CommentComposer({
    required this.hint,
    required this.submitLabel,
    required this.onSubmit,
    this.initialText = '',
    this.onCancel,
    this.autofocus = false,
    super.key,
  });

  /// Placeholder and accessibility label.
  final String hint;

  /// Text of the submit button.
  final String submitLabel;

  /// Sends the text; true when it was accepted.
  final Future<bool> Function(String text) onSubmit;

  /// Text to start with (editing a comment).
  final String initialText;

  /// Shows a cancel button when set. Submitting an edit without changes
  /// cancels too, like web.
  final VoidCallback? onCancel;

  /// Whether the box takes the keyboard focus when it appears (edit and
  /// reply boxes do).
  final bool autofocus;

  @override
  State<CommentComposer> createState() => _CommentComposerState();
}

class _CommentComposerState extends State<CommentComposer> {
  late final TextEditingController _controller = TextEditingController(
    text: widget.initialText,
  )..addListener(() => setState(() {}));
  bool _isSending = false;

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    final text = _controller.text.trim();
    if (text.isEmpty || _isSending) return;
    final cancel = widget.onCancel;
    if (cancel != null && text == widget.initialText) {
      cancel();
      return;
    }
    setState(() => _isSending = true);
    final ok = await widget.onSubmit(text);
    if (!mounted) return;
    setState(() => _isSending = false);
    if (ok && widget.initialText.isEmpty) _controller.clear();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final theme = Theme.of(context);
    final text = _controller.text;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        TextField(
          controller: _controller,
          autofocus: widget.autofocus,
          minLines: 2,
          maxLines: 6,
          keyboardType: TextInputType.multiline,
          textCapitalization: TextCapitalization.sentences,
          inputFormatters: [_lengthLimit],
          decoration: InputDecoration(
            hintText: widget.hint,
            border: const OutlineInputBorder(),
          ),
        ),
        const SizedBox(height: 4),
        Wrap(
          alignment: WrapAlignment.spaceBetween,
          crossAxisAlignment: WrapCrossAlignment.center,
          spacing: 8,
          children: [
            Text(
              text.isEmpty
                  ? ''
                  : l10n.comments_char_remaining(
                      commentMaxLength - text.length,
                    ),
              style: theme.textTheme.bodySmall?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
              ),
            ),
            Wrap(
              spacing: 8,
              children: [
                if (widget.onCancel != null)
                  TextButton(
                    onPressed: _isSending ? null : widget.onCancel,
                    child: Text(l10n.common_cancel),
                  ),
                FilledButton(
                  onPressed: text.trim().isEmpty || _isSending ? null : _submit,
                  child: Text(
                    _isSending ? l10n.common_loading : widget.submitLabel,
                  ),
                ),
              ],
            ),
          ],
        ),
      ],
    );
  }
}
