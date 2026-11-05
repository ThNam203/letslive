/// Snack bars that tell the user an action failed: a refused request, or a
/// photo library that would not open.
library;

import 'package:flutter/material.dart';
import 'package:letslive_mobile/core/l10n/api_message.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/network/api_response.dart';

/// Tells the user why an action (react, comment, like, ...) failed, in a
/// snack bar. Does nothing when [failure] is null (the action worked).
void showActionError(BuildContext context, ApiResponse<Object?>? failure) {
  if (failure == null || !context.mounted) return;
  final message = apiErrorMessage(AppLocalizations.of(context), failure);
  ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(message)));
}

/// Thrown by the photo / video pickers when the picker or the chosen file
/// cannot be opened (e.g. photo access denied). Cancelling is not a failure:
/// the pickers answer "nothing chosen" then.
class PickerException implements Exception {
  const PickerException(this.cause);

  /// The platform error behind it, for the log.
  final Object cause;

  @override
  String toString() => 'PickerException: $cause';
}

/// Runs a picker; when it fails ([PickerException]) says "Could not open
/// your photos" in a snack bar and answers null, so a tap never silently
/// does nothing.
Future<T?> pickWithFeedback<T>(
  BuildContext context,
  Future<T> Function() pick,
) async {
  final messenger = ScaffoldMessenger.of(context);
  final message = AppLocalizations.of(context).common_photo_library_error;
  try {
    return await pick();
  } on PickerException {
    messenger.showSnackBar(SnackBar(content: Text(message)));
    return null;
  }
}
