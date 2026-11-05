/// Hooks of the settings screens that reach the device: the photo library and
/// the image decoder. Widget tests override them with fakes.
library;

import 'dart:ui' as ui;

import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:image_picker/image_picker.dart';
import 'package:letslive_mobile/core/validation/field_limits.dart';
import 'package:letslive_mobile/core/widgets/action_feedback.dart';
import 'package:letslive_mobile/features/settings/domain/image_checks.dart';
import 'package:letslive_mobile/features/user/data/picked_image.dart';

/// Lets the user choose one image from the photo library; null when they
/// cancel. Throws [PickerException] when the picker fails.
typedef ImagePick = Future<PickedImage?> Function();

/// The photo library picker. On Android it uses the system picker (no
/// permission needed); iOS asks through `NSPhotoLibraryUsageDescription`.
final imagePickProvider = Provider<ImagePick>(
  (ref) => () async {
    try {
      final file = await ImagePicker().pickImage(source: ImageSource.gallery);
      if (file == null) return null;
      return PickedImage(
        name: file.name,
        bytes: await file.readAsBytes(),
        mimeType: file.mimeType ?? mimeTypeOfName(file.name),
      );
    } on PlatformException catch (e) {
      // The picker or the file could not be opened (e.g. permission denied).
      throw PickerException(e);
    }
  },
);

/// What the multi-image picker returned: the images to use, and the names of
/// those refused for being over the size limit (never read into memory).
typedef PickedImages = ({List<PickedImage> images, List<String> oversized});

/// Lets the user choose several images at once; both lists are empty when they
/// cancel. Throws [PickerException] when the picker fails. [limit] is the
/// most the caller can still use.
typedef ImagesPick = Future<PickedImages> Function(int limit);

/// The photo library picker for several images (direct-message pictures).
/// Files over `FieldLimits.generalUploadMaxFileMb` are only measured, not
/// read.
final imagesPickProvider = Provider<ImagesPick>(
  (ref) => (limit) async {
    try {
      // The plugin needs a limit of at least 2.
      final files = await ImagePicker().pickMultiImage(
        limit: limit < 2 ? 2 : limit,
      );
      const maxBytes = FieldLimits.generalUploadMaxFileMb * 1024 * 1024;
      final images = <PickedImage>[];
      final oversized = <String>[];
      for (final file in files.take(limit)) {
        if (await file.length() > maxBytes) {
          oversized.add(file.name);
          continue;
        }
        images.add(
          PickedImage(
            name: file.name,
            bytes: await file.readAsBytes(),
            mimeType: file.mimeType ?? mimeTypeOfName(file.name),
          ),
        );
      }
      return (images: images, oversized: oversized);
    } on PlatformException catch (e) {
      throw PickerException(e);
    }
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
