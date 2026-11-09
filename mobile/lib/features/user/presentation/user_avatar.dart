import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';

/// Round avatar: the picture when there is one and it loads, else the first
/// letter of [name] (web `UserAvatar`). [name] may be null while the user is
/// still loading.
class UserAvatar extends StatelessWidget {
  const UserAvatar({this.pictureUrl, this.name, this.radius = 20, super.key});

  /// Absolute URL of the avatar; null shows the initial.
  final String? pictureUrl;

  /// Source of the initial; null while the user is loading.
  final String? name;

  /// Circle radius in logical pixels.
  final double radius;

  @override
  Widget build(BuildContext context) {
    final url = pictureUrl;
    final initial = (name ?? '').characters.firstOrNull?.toUpperCase();
    return CircleAvatar(
      radius: radius,
      foregroundImage: url == null ? null : CachedNetworkImageProvider(url),
      // The initial stays visible when the picture fails to load.
      onForegroundImageError: url == null ? null : (_, _) {},
      child: initial == null ? Icon(Icons.person, size: radius) : Text(initial),
    );
  }
}
