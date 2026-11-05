import 'dart:typed_data';

/// An image the user chose from the photo library, read into memory (at most
/// 10 MB, see `FieldLimits`). Keeps `image_picker` types out of the rest of
/// the app and out of tests.
class PickedImage {
  const PickedImage({required this.name, required this.bytes, this.mimeType});

  /// File name sent as the multipart filename.
  final String name;
  final Uint8List bytes;

  /// MIME type reported by the picker, e.g. `image/jpeg`; null if unknown.
  final String? mimeType;

  /// File size in bytes.
  int get size => bytes.length;
}
