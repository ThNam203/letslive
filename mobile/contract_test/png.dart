import 'dart:io';
import 'dart:typed_data';

/// A valid solid-colour PNG of [width] x [height] pixels, built without an
/// image library, for upload contract tests.
Uint8List solidPng(int width, int height) {
  final row = Uint8List(1 + width * 3); // filter byte 0, then RGB pixels
  for (var x = 0; x < width; x++) {
    row[1 + x * 3] = 200;
    row[2 + x * 3] = 60;
    row[3 + x * 3] = 90;
  }
  final raw = BytesBuilder();
  for (var y = 0; y < height; y++) {
    raw.add(row);
  }
  final header = ByteData(13)
    ..setUint32(0, width)
    ..setUint32(4, height)
    ..setUint8(8, 8) // bit depth
    ..setUint8(9, 2); // colour type: RGB
  return (BytesBuilder()
        ..add(const [0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A])
        ..add(_chunk('IHDR', header.buffer.asUint8List()))
        ..add(_chunk('IDAT', zlib.encode(raw.toBytes())))
        ..add(_chunk('IEND', const [])))
      .toBytes();
}

Uint8List _chunk(String type, List<int> data) {
  final typeBytes = type.codeUnits;
  final out = ByteData(12 + data.length)..setUint32(0, data.length);
  final bytes = out.buffer.asUint8List()
    ..setRange(4, 8, typeBytes)
    ..setRange(8, 8 + data.length, data);
  out.setUint32(8 + data.length, _crc32([...typeBytes, ...data]));
  return bytes;
}

int _crc32(List<int> bytes) {
  var crc = 0xFFFFFFFF;
  for (final byte in bytes) {
    crc ^= byte;
    for (var i = 0; i < 8; i++) {
      crc = (crc & 1) == 1 ? (crc >> 1) ^ 0xEDB88320 : crc >> 1;
    }
  }
  return crc ^ 0xFFFFFFFF;
}
