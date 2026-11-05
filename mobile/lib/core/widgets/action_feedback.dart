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
