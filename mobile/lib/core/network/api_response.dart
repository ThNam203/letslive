class Meta {
  final int? page;
  final int? pageSize;
  final int? total;

  const Meta({this.page, this.pageSize, this.total});

  factory Meta.fromJson(Map<String, dynamic> json) {
    return Meta(
      page: json['page'] as int?,
      pageSize: json['page_size'] as int?,
      total: json['total'] as int?,
    );
  }
}

class ApiResponse<T> {
  final String requestId;
  final bool success;
  final int statusCode;
  final int code;
  final String key;
  final String message;
  final T? data;
  final Meta? meta;
  final List<Map<String, dynamic>>? errorDetails;

  const ApiResponse({
    required this.requestId,
    required this.success,
    required this.statusCode,
    required this.code,
    required this.key,
    required this.message,
    this.data,
    this.meta,
    this.errorDetails,
  });

  factory ApiResponse.fromJson(
    Map<String, dynamic> json,
    int statusCode, {
    T Function(dynamic)? fromJsonT,
  }) {
    return ApiResponse<T>(
      requestId: json['requestId'] as String? ?? '',
      success: json['success'] as bool? ?? false,
      statusCode: statusCode,
      code: json['code'] as int? ?? 0,
      key: json['key'] as String? ?? '',
      message: json['message'] as String? ?? '',
      data: json['data'] != null && fromJsonT != null
          ? fromJsonT(json['data'])
          : json['data'] as T?,
      meta: json['meta'] != null
          ? Meta.fromJson(json['meta'] as Map<String, dynamic>)
          : null,
      errorDetails: (json['errorDetails'] as List<dynamic>?)
          ?.map((e) => e as Map<String, dynamic>)
          .toList(),
    );
  }
}

/// Business-level API codes matching the web app.
abstract final class ApiCode {
  // General / Auth (200xx)
  static const resErrInvalidInput = 20000;
  static const resErrInvalidPayload = 20001;
  static const resErrAuthAlreadyExists = 20002;
  static const resErrCaptchaFailed = 20003;
  static const resErrPasswordNotMatch = 20004;
  static const resErrUnauthorized = 20005;
  static const resErrSignUpOtpExpired = 20006;
  static const resErrEmailOrPasswordIncorrect = 20007;
  static const resErrForbidden = 20008;
  static const resErrAuthNotFound = 20009;
  static const resErrRefreshTokenNotFound = 20010;
  static const resErrSignUpOtpNotFound = 20011;
  static const resErrRouteNotFound = 20012;
  static const resErrSignUpOtpAlreadyUsed = 20013;
  static const resErrFailedToCreateSignUpOtp = 20014;
  static const resErrDatabaseQuery = 20015;
  static const resErrDatabaseIssue = 20016;
  static const resErrInternalServer = 20017;
  static const resErrFailedToSendVerification = 20018;

  // User (300xx)
  static const resErrUserNotFound = 30000;
  static const resErrImageTooLarge = 30001;
  static const resErrNotificationNotFound = 30002;

  // Livestream / VOD (400xx)
  static const resErrLivestreamUpdateAfterEnded = 40000;
  static const resErrLivestreamNotFound = 40001;
  static const resErrVodNotFound = 40002;
  static const resErrEndAlreadyEndedLivestream = 40003;
  static const resErrQueryScanFailed = 40004;
  static const resErrLivestreamCreateFailed = 40005;
  static const resErrLivestreamUpdateFailed = 40006;
  static const resErrVodCreateFailed = 40007;
  static const resErrVodUpdateFailed = 40008;
  static const resErrVodCommentNotFound = 40009;
  static const resErrVodCommentCreateFailed = 40010;
  static const resErrVodCommentAlreadyLiked = 40011;
  static const resErrVodCommentNotLiked = 40012;
  static const resErrVodCommentDeleteFailed = 40013;

  // DM / Conversations (500xx)
  static const resErrRoomNotFound = 50018;
  static const resErrConversationNotFound = 50019;
  static const resErrDmAlreadyExists = 50020;
  static const resErrNotParticipant = 50021;
  static const resErrInsufficientRole = 50022;
  static const resErrDmMessageNotFound = 50023;
  static const resErrCannotMessageSelf = 50024;
  static const resErrTooManyParticipants = 50025;
}
