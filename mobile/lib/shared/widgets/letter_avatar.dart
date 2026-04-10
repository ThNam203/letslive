import 'package:flutter/widgets.dart';
import 'package:forui/forui.dart';

/// Circular muted background with centered letter text.
class LetterAvatar extends StatelessWidget {
  final String text;
  final TextStyle textStyle;
  final double size;

  const LetterAvatar({
    super.key,
    required this.text,
    required this.textStyle,
    this.size = 28,
  });

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;

    return ClipOval(
      child: Container(
        width: size,
        height: size,
        color: colors.muted,
        alignment: Alignment.center,
        child: Text(text, style: textStyle),
      ),
    );
  }
}
