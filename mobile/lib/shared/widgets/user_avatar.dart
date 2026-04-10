import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/widgets.dart';
import 'package:forui/forui.dart';

/// Reusable circular avatar with optional network image.
class UserAvatar extends StatelessWidget {
  final String? imageUrl;
  final double size;
  final Widget? fallback;
  final String? fallbackText;
  final TextStyle? textStyle;

  const UserAvatar({
    super.key,
    this.imageUrl,
    this.size = 40,
    this.fallback,
    this.fallbackText,
    this.textStyle,
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
        child: _buildChild(),
      ),
    );
  }

  Widget _buildChild() {
    if (imageUrl != null && imageUrl!.isNotEmpty) {
      return Image(
        image: CachedNetworkImageProvider(imageUrl!),
        width: size,
        height: size,
        fit: BoxFit.cover,
      );
    }

    if (fallback != null) return fallback!;

    if (fallbackText != null && fallbackText!.isNotEmpty) {
      return Text(fallbackText!, style: textStyle);
    }

    return const SizedBox.shrink();
  }
}
