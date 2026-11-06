import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/features/vod/domain/timestamps.dart';

/// What timestamps in texts under a video can do: seek the player. Provided
/// above the video's description and comments (web `VODPlaybackProvider`);
/// without it [TimestampedText] shows plain text.
class VodPlayback extends InheritedWidget {
  const VodPlayback({
    required this.seekTo,
    required this.duration,
    required super.child,
    super.key,
  });

  /// Seeks the player to a position, in seconds, and plays.
  final void Function(int seconds) seekTo;

  /// Length of the video in seconds; later timestamps are not links.
  final int duration;

  /// The nearest [VodPlayback], or null outside a video.
  static VodPlayback? maybeOf(BuildContext context) =>
      context.dependOnInheritedWidgetOfExactType<VodPlayback>();

  @override
  bool updateShouldNotify(VodPlayback old) =>
      seekTo != old.seekTo || duration != old.duration;
}

/// Text with "1:23"-style timestamps as tappable links that seek the video
/// (web `TimestampedText`). Wraps like any [Text], so long Vietnamese and
/// emoji text never overflows.
class TimestampedText extends StatefulWidget {
  const TimestampedText(this.text, {this.style, this.maxLines, super.key});

  /// The text to show.
  final String text;

  /// Lines to cut the text at (with an ellipsis); null shows all of it.
  final int? maxLines;

  /// Text style; defaults to the ambient one.
  final TextStyle? style;

  @override
  State<TimestampedText> createState() => _TimestampedTextState();
}

class _TimestampedTextState extends State<TimestampedText> {
  /// Tap recognizers of the current build. Recognizers must be disposed, so
  /// each build replaces the previous ones.
  final List<TapGestureRecognizer> _recognizers = [];

  void _disposeRecognizers() {
    for (final recognizer in _recognizers) {
      recognizer.dispose();
    }
    _recognizers.clear();
  }

  @override
  void dispose() {
    _disposeRecognizers();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    _disposeRecognizers();
    final playback = VodPlayback.maybeOf(context);
    if (playback == null) {
      return Text(
        widget.text,
        style: widget.style,
        maxLines: widget.maxLines,
        overflow: widget.maxLines == null ? null : TextOverflow.ellipsis,
      );
    }

    final l10n = AppLocalizations.of(context);
    final linkStyle = TextStyle(
      color: Theme.of(context).colorScheme.primary,
      fontWeight: FontWeight.w500,
    );
    return Text.rich(
      TextSpan(
        children: [
          for (final segment in splitTimestamps(
            widget.text,
            maxSeconds: playback.duration,
          ))
            switch (segment) {
              PlainSegment(:final text) => TextSpan(text: text),
              TimestampSegment(:final text, :final seconds) => TextSpan(
                text: text,
                style: linkStyle,
                semanticsLabel: l10n.common_vod_seek_to(text),
                recognizer: _recognizer(() => playback.seekTo(seconds)),
              ),
            },
        ],
      ),
      style: widget.style,
      maxLines: widget.maxLines,
      overflow: widget.maxLines == null ? null : TextOverflow.ellipsis,
    );
  }

  TapGestureRecognizer _recognizer(VoidCallback onTap) {
    final recognizer = TapGestureRecognizer()..onTap = onTap;
    _recognizers.add(recognizer);
    return recognizer;
  }
}
