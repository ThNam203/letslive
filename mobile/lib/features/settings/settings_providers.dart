/// Hooks of the settings screens that reach the device: the photo library and
/// the image decoder. Widget tests override them with fakes.
library;

import 'dart:typed_data';
import 'dart:ui' as ui;

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:image_picker/image_picker.dart';
import 'package:letslive_mobile/features/settings/domain/image_checks.dart';
import 'package:letslive_mobile/features/user/data/picked_image.dart';

/// Lets the user choose one image from the photo library; null when they
/// cancel.
typedef ImagePick = Future<PickedImage?> Function();

/// The photo library picker. On Android it uses the system picker (no
/// permission needed); iOS asks through `NSPhotoLibraryUsageDescription`.
final imagePickProvider = Provider<ImagePick>(
  (ref) => () async {
    final file = await ImagePicker().pickImage(source: ImageSource.gallery);
    if (file == null) return null;
    return PickedImage(
      name: file.name,
      bytes: await file.readAsBytes(),
      mimeType: file.mimeType ?? mimeTypeOfName(file.name),
    );
  },
);

/// Width and height of an encoded image, or null if it cannot be decoded.
typedef ImageSizeReader = Future<({int width, int height})?> Function(
  Uint8List bytes,
);

/// Reads the size with the engine's own decoder.
final imageSizeReaderProvider = Provider<ImageSizeReader>(
  (ref) => (bytes) async {
    try {
      final codec = await ui.instantiateImageCodec(bytes);
      final frame = await codec.getNextFrame();
      final size = (width: frame.image.width, height: frame.image.height);
      frame.image.dispose();
      codec.dispose();
      return size;
    } on Exception {
      return null;
    }
  },
);
