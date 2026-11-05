import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/network/api_response.dart';

/// The text to show for a failed [response], like web's
/// `t("api-response:" + key)`: the localized `api_response_<key>` when the
/// app knows the key, else the server's own message.
///
/// gen-l10n has no lookup by string, so known keys are listed here; add a
/// key when a feature starts showing it. Transport failures use the generic
/// fetch error; a malformed body the generic error (its message is
/// technical).
String apiErrorMessage(AppLocalizations l10n, ApiResponse<Object?> response) {
  switch (response.failure) {
    case null:
      break;
    case ApiFailureKind.malformedResponse:
      return l10n.api_response_default_error;
    case ApiFailureKind.timeout ||
        ApiFailureKind.network ||
        ApiFailureKind.cancelled:
      return l10n.fetch_error_client_fetch_error;
  }
  final known = switch (response.key) {
    'res_err_account_frozen' => l10n.api_response_res_err_account_frozen,
    'res_err_account_not_found' => l10n.api_response_res_err_account_not_found,
    'res_err_auth_already_exists' =>
      l10n.api_response_res_err_auth_already_exists,
    'res_err_auth_not_found' => l10n.api_response_res_err_auth_not_found,
    'res_err_captcha_failed' => l10n.api_response_res_err_captcha_failed,
    'res_err_email_or_password_incorrect' =>
      l10n.api_response_res_err_email_or_password_incorrect,
    'res_err_failed_to_create_sign_up_otp' =>
      l10n.api_response_res_err_failed_to_create_sign_up_otp,
    'res_err_failed_to_send_verification' =>
      l10n.api_response_res_err_failed_to_send_verification,
    'res_err_image_dimensions_out_of_range' =>
      l10n.api_response_res_err_image_dimensions_out_of_range,
    'res_err_image_too_large' => l10n.api_response_res_err_image_too_large,
    'res_err_invalid_image' => l10n.api_response_res_err_invalid_image,
    'res_err_password_not_match' =>
      l10n.api_response_res_err_password_not_match,
    'res_err_internal_server' => l10n.api_response_res_err_internal_server,
    'res_err_invalid_input' => l10n.api_response_res_err_invalid_input,
    'res_err_invalid_payload' => l10n.api_response_res_err_invalid_payload,
    'res_err_sign_up_otp_already_used' =>
      l10n.api_response_res_err_sign_up_otp_already_used,
    'res_err_sign_up_otp_expired' =>
      l10n.api_response_res_err_sign_up_otp_expired,
    'res_err_sign_up_otp_not_found' =>
      l10n.api_response_res_err_sign_up_otp_not_found,
    'res_err_unauthorized' => l10n.api_response_res_err_unauthorized,
    'res_err_user_not_found' => l10n.api_response_res_err_user_not_found,
    'res_err_username_taken' => l10n.api_response_res_err_username_taken,
    'res_err_vod_not_found' => l10n.api_response_res_err_vod_not_found,
    'res_err_vod_view_threshold' =>
      l10n.api_response_res_err_vod_view_threshold,
    'res_err_vod_comment_not_found' =>
      l10n.api_response_res_err_vod_comment_not_found,
    'res_err_vod_comment_create_failed' =>
      l10n.api_response_res_err_vod_comment_create_failed,
    'res_err_vod_comment_update_failed' =>
      l10n.api_response_res_err_vod_comment_update_failed,
    'res_err_vod_comment_delete_failed' =>
      l10n.api_response_res_err_vod_comment_delete_failed,
    'res_err_vod_comment_already_liked' =>
      l10n.api_response_res_err_vod_comment_already_liked,
    'res_err_vod_comment_not_liked' =>
      l10n.api_response_res_err_vod_comment_not_liked,
    _ => null,
  };
  if (known != null) return known;
  return response.message.isNotEmpty
      ? response.message
      : l10n.api_response_default_error;
}
