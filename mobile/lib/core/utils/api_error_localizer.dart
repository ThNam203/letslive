import 'package:flutter/widgets.dart';

import '../../l10n/app_localizations.dart';

/// Maps an API response `key` (snake_case) to the corresponding
/// localized message string from AppLocalizations.
String getLocalizedApiMessage(BuildContext context, String apiKey) {
  final l10n = AppLocalizations.of(context);
  final getter = _apiKeyMap[apiKey];
  if (getter != null) return getter(l10n);
  return l10n.apiDefaultError;
}

final Map<String, String Function(AppLocalizations)> _apiKeyMap = {
  // Auth errors
  'res_err_invalid_input': (l) => l.apiResErrInvalidInput,
  'res_err_invalid_payload': (l) => l.apiResErrInvalidPayload,
  'res_err_auth_already_exists': (l) => l.apiResErrAuthAlreadyExists,
  'res_err_captcha_failed': (l) => l.apiResErrCaptchaFailed,
  'res_err_password_not_match': (l) => l.apiResErrPasswordNotMatch,
  'res_err_unauthorized': (l) => l.apiResErrUnauthorized,
  'res_err_sign_up_otp_expired': (l) => l.apiResErrSignUpOtpExpired,
  'res_err_email_or_password_incorrect': (l) =>
      l.apiResErrEmailOrPasswordIncorrect,
  'res_err_forbidden': (l) => l.apiResErrForbidden,
  'res_err_auth_not_found': (l) => l.apiResErrAuthNotFound,
  'res_err_refresh_token_not_found': (l) => l.apiResErrRefreshTokenNotFound,
  'res_err_sign_up_otp_not_found': (l) => l.apiResErrSignUpOtpNotFound,
  'res_err_route_not_found': (l) => l.apiResErrRouteNotFound,
  'res_err_sign_up_otp_already_used': (l) => l.apiResErrSignUpOtpAlreadyUsed,
  'res_err_failed_to_create_sign_up_otp': (l) =>
      l.apiResErrFailedToCreateSignUpOtp,
  'res_err_database_query': (l) => l.apiResErrDatabaseQuery,
  'res_err_database_issue': (l) => l.apiResErrDatabaseIssue,
  'res_err_internal_server': (l) => l.apiResErrInternalServer,
  'res_err_failed_to_send_verification': (l) =>
      l.apiResErrFailedToSendVerification,
  // User errors
  'res_err_user_not_found': (l) => l.apiResErrUserNotFound,
  'res_err_image_too_large': (l) => l.apiResErrImageTooLarge,
  // Livestream / VOD errors
  'res_err_livestream_update_after_ended': (l) =>
      l.apiResErrLivestreamUpdateAfterEnded,
  'res_err_livestream_not_found': (l) => l.apiResErrLivestreamNotFound,
  'res_err_vod_not_found': (l) => l.apiResErrVodNotFound,
  'res_err_end_already_ended_livestream': (l) =>
      l.apiResErrEndAlreadyEndedLivestream,
  'res_err_vod_comment_not_found': (l) => l.apiResErrVodCommentNotFound,
  'res_err_vod_comment_create_failed': (l) =>
      l.apiResErrVodCommentCreateFailed,
  'res_err_vod_comment_already_liked': (l) =>
      l.apiResErrVodCommentAlreadyLiked,
  'res_err_vod_comment_not_liked': (l) => l.apiResErrVodCommentNotLiked,
  'res_err_vod_comment_delete_failed': (l) =>
      l.apiResErrVodCommentDeleteFailed,
  // Success keys
  'res_succ_sent_verification_email': (l) =>
      l.apiResSuccSentVerificationEmail,
  'res_succ_ok': (l) => l.apiResSuccOk,
  'res_succ_login': (l) => l.apiResSuccLogin,
  'res_succ_sign_up': (l) => l.apiResSuccSignUp,
};
