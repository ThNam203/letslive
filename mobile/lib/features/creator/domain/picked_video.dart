import 'package:letslive_mobile/core/validation/field_limits.dart';

/// A video the user chose on the device. Only the path is kept: the upload
/// streams the file from disk, so a 100 MB video is never held in memory.
class PickedVideo {
  const PickedVideo({
    required this.path,
    required this.name,
    required this.size,
  });

  /// Local file path.
  final String path;

  /// File name sent as the multipart filename.
  final String name;

  /// File size in bytes.
  final int size;
}

/// Why a picked video is refused before any upload (web's checks in
/// `settings/upload/page.tsx`).
enum VideoProblem {
  /// The extension is not in `FieldLimits.vodUploadExtensions`.
  unsupportedFormat,

  /// Larger than `FieldLimits.vodUploadMaxMb`.
  tooLarge,
}

/// The reason [video] cannot be uploaded, or null when it can. Mirrors web:
/// the extension is checked first, then the size. The server checks again.
VideoProblem? checkVideo(PickedVideo video) {
  final dot = video.name.lastIndexOf('.');
  final extension = dot < 0 ? '' : video.name.substring(dot).toLowerCase();
  if (!FieldLimits.vodUploadExtensions.contains(extension)) {
    return VideoProblem.unsupportedFormat;
  }
  if (video.size > FieldLimits.vodUploadMaxMb * 1024 * 1024) {
    return VideoProblem.tooLarge;
  }
  return null;
}

/// The file name without its extension, cut to the title limit: the title
/// suggested for a picked video (web does the same).
String suggestedTitle(String fileName) {
  final dot = fileName.lastIndexOf('.');
  final base = dot <= 0 ? fileName : fileName.substring(0, dot);
  return base.length <= FieldLimits.vodTitleMax
      ? base
      : base.substring(0, FieldLimits.vodTitleMax);
}
