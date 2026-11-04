import 'package:flutter/material.dart';
import 'package:letslive_mobile/features/search/domain/highlight.dart';

/// [text] with the parts that match [query] marked (web `Highlight`). Matching
/// ignores case and accents, see `findHighlightRanges`. Clipped to [maxLines]
/// so a very long text never overflows.
class HighlightedText extends StatelessWidget {
  const HighlightedText({
    required this.text,
    required this.query,
    this.style,
    this.maxLines = 1,
    super.key,
  });

  final String text;
  final String query;
  final TextStyle? style;
  final int maxLines;

  @override
  Widget build(BuildContext context) {
    final mark = TextStyle(
      backgroundColor: Colors.yellow.withValues(alpha: 0.4),
    );
    return Text.rich(
      TextSpan(
        children: [
          for (final segment in splitByRanges(
            text,
            findHighlightRanges(text, query),
          ))
            TextSpan(text: segment.text, style: segment.isMatch ? mark : null),
        ],
      ),
      style: style,
      maxLines: maxLines,
      overflow: TextOverflow.ellipsis,
    );
  }
}
