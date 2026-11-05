import 'dart:typed_data';

import 'package:dio/dio.dart';

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

  /// The image as a multipart file part, with its MIME type when known.
  /// Shared by every upload that sends a picked image.
  MultipartFile toMultipart() {
    final mime = mimeType;
    return MultipartFile.fromBytes(
      bytes,
      filename: name,
      contentType: mime == null ? null : DioMediaType.parse(mime),
    );
  }
}
