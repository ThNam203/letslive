/// Client-side checks on a picked profile or banner image, the same ones web
/// makes before uploading (`web/constant/image.ts`, settings `image-hover`
/// and `profile-banner`). The server checks again; these only spare the
/// upload.
library;

import 'package:letslive_mobile/core/validation/field_limits.dart';
import 'package:letslive_mobile/features/user/data/picked_image.dart';

/// Why a picked image is refused.
enum ImageProblem {
  /// Larger than the megabyte limit given to [checkFileSize].
  tooLarge,

  /// Not one of `FieldLimits.avatarUploadTypes`.
  unsupportedType,

  /// A side outside `FieldLimits.avatarMinDimension` ..
  /// `FieldLimits.avatarSourceMaxDimension`.
  dimensions,
}

/// MIME types by file extension, for pickers that do not report one.
const Map<String, String> _mimeByExtension = {
  'jpg': 'image/jpeg',
  'jpeg': 'image/jpeg',
  'png': 'image/png',
  'gif': 'image/gif',
  'webp': 'image/webp',
};

/// The MIME type for [fileName] from its extension, or null if unknown.
String? mimeTypeOfName(String fileName) {
  final dot = fileName.lastIndexOf('.');
  if (dot < 0) return null;
  return _mimeByExtension[fileName.substring(dot + 1).toLowerCase()];
}

/// [ImageProblem.tooLarge] when [image] is over [maxMb] megabytes.
ImageProblem? checkFileSize(PickedImage image, int maxMb) =>
    image.size > maxMb * 1024 * 1024 ? ImageProblem.tooLarge : null;

/// [ImageProblem.unsupportedType] unless the server can decode [image] as a
/// profile picture (web re-encodes to JPEG first, mobile uploads as is).
ImageProblem? checkAvatarType(PickedImage image) =>
    FieldLimits.avatarUploadTypes.contains(image.mimeType)
    ? null
    : ImageProblem.unsupportedType;

/// [ImageProblem.dimensions] when a side of [width] x [height] is outside the
/// range web accepts for a profile picture.
ImageProblem? checkAvatarDimensions(int width, int height) {
  bool inRange(int side) =>
      side >= FieldLimits.avatarMinDimension &&
      side <= FieldLimits.avatarSourceMaxDimension;
  return inRange(width) && inRange(height) ? null : ImageProblem.dimensions;
}
