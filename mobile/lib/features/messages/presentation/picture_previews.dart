import 'package:flutter/material.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/features/user/data/picked_image.dart';

/// A picture waiting in the composer and, once uploaded, the path the server
/// stored it under (null until then, or after a failed upload).
class PendingPicture {
  const PendingPicture(this.image, [this.path]);

  final PickedImage image;
  final String? path;
}

/// The row of thumbnails above the message box, each with a remove button
/// (web `MessageInput` previews). While the message is being sent they dim,
/// and a picture not yet uploaded shows a spinner.
class PicturePreviews extends StatelessWidget {
  const PicturePreviews({
    required this.pictures,
    required this.sending,
    required this.onRemove,
    super.key,
  });

  final List<PendingPicture> pictures;
  final bool sending;
  final void Function(PendingPicture picture) onRemove;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return SizedBox(
      height: 76,
      child: ListView.separated(
        scrollDirection: Axis.horizontal,
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
        itemCount: pictures.length,
        separatorBuilder: (_, _) => const SizedBox(width: 8),
        itemBuilder: (context, index) {
          final picture = pictures[index];
          return Stack(
            key: ValueKey('picture-$index'),
            children: [
              ClipRRect(
                borderRadius: BorderRadius.circular(8),
                child: Opacity(
                  opacity: sending ? 0.5 : 1,
                  child: Image.memory(
                    picture.image.bytes,
                    width: 64,
                    height: 64,
                    fit: BoxFit.cover,
                  ),
                ),
              ),
              if (sending && picture.path == null)
                const Positioned.fill(
                  child: Center(
                    child: SizedBox.square(
                      dimension: 20,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    ),
                  ),
                )
              else if (!sending)
                Positioned(
                  top: -8,
                  right: -8,
                  child: IconButton(
                    key: ValueKey('remove-picture-$index'),
                    tooltip: l10n.messages_remove_image,
                    iconSize: 16,
                    visualDensity: VisualDensity.compact,
                    icon: const Icon(Icons.cancel),
                    onPressed: () => onRemove(picture),
                  ),
                ),
            ],
          );
        },
      ),
    );
  }
}
