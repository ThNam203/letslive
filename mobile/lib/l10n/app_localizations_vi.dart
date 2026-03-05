// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Vietnamese (`vi`).
class AppLocalizationsVi extends AppLocalizations {
  AppLocalizationsVi([String locale = 'vi']) : super(locale);

  @override
  String get appTitle => 'Let\'s Live';

  @override
  String get loading => 'Đang tải...';

  @override
  String get setting => 'Cài đặt';

  @override
  String get save => 'Lưu';

  @override
  String get cancel => 'Hủy';

  @override
  String get or => 'hoặc';

  @override
  String get goHome => 'Về trang chủ';

  @override
  String get pageNotFound => 'Không tìm thấy trang';

  @override
  String get saveChanges => 'Lưu thay đổi';

  @override
  String get username => 'Tên người dùng';

  @override
  String get livestreams => 'Livestreams';

  @override
  String get channels => 'Kênh';

  @override
  String get videos => 'Videos';

  @override
  String get home => 'Trang chủ';

  @override
  String get searchUsers => 'Tìm kiếm người dùng';

  @override
  String get gotIt => 'Đã hiểu';

  @override
  String get howToLivestream => 'Cách livestream';

  @override
  String get startYourLivestream => 'Bắt đầu livestream của bạn';

  @override
  String get liveStreaming => 'Live Streaming';

  @override
  String get follow => 'Theo dõi';

  @override
  String get unfollow => 'Bỏ theo dõi';

  @override
  String followersWithCount(int count) {
    return 'Người theo dõi: $count';
  }

  @override
  String get joined => 'Tham gia vào';

  @override
  String get showMore => 'Xem thêm';

  @override
  String get pleaseWaitWhileLoading => 'Vui lòng chờ đợi việc tải dữ liệu';

  @override
  String get noLivestreams => 'Không có livestreams';

  @override
  String get noLivestreamsDescription =>
      'Hiện tại không có ai livestream, vui lòng quay lại sau hoặc xem thử những nội dung videos có sẵn';

  @override
  String get noVideos => 'Không có videos';

  @override
  String get noVideosDescription =>
      'Hiện không có videos nào có sẵn, vui lòng trở lại sau để xem các content mới nhất';

  @override
  String get searching => 'Đang tìm kiếm...';

  @override
  String get noUsersFound => 'Không tìm thấy người dùng';

  @override
  String get noDescription => 'Không có mô tả';

  @override
  String get add => 'Thêm';

  @override
  String get bio => 'Tiểu sử';

  @override
  String startedAt(String time) {
    return 'Bắt đầu $time trước';
  }

  @override
  String get live => 'Trực tiếp';

  @override
  String get following => 'Đang theo dõi';

  @override
  String get recommended => 'Gợi ý';

  @override
  String timeSecondsAgo(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$count giây',
    );
    return '$_temp0';
  }

  @override
  String timeMinutesAgo(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$count phút',
    );
    return '$_temp0';
  }

  @override
  String timeHoursAgo(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$count giờ',
    );
    return '$_temp0';
  }

  @override
  String timeDaysAgo(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$count ngày',
    );
    return '$_temp0';
  }

  @override
  String timeWeeksAgo(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$count tuần',
    );
    return '$_temp0';
  }

  @override
  String timeMonthsAgo(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$count tháng',
    );
    return '$_temp0';
  }

  @override
  String timeYearsAgo(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$count năm',
    );
    return '$_temp0';
  }

  @override
  String get authLogin => 'Đăng nhập';

  @override
  String get authLoginTitle => 'Chào mừng trở lại!';

  @override
  String get authLoginSubtitle => 'Truy cập thế giới ngay bây giờ.';

  @override
  String get authEmail => 'Email';

  @override
  String get authPassword => 'Mật khẩu';

  @override
  String get authForgotPassword => 'Quên mật khẩu?';

  @override
  String get authSignup => 'Đăng ký';

  @override
  String get authSignupTitle => 'Chào mừng! Đăng ký để khám phá thế giới mới?';

  @override
  String get authSignupSubtitle => 'Chọn một phương thức bên dưới để bắt đầu';

  @override
  String get authLogout => 'Đăng xuất';

  @override
  String get authNoAccount => 'Chưa có tài khoản?';

  @override
  String get authHaveAccount => 'Đã có tài khoản?';

  @override
  String get authConfirmPassword => 'Xác nhận mật khẩu';

  @override
  String get authAccountCreatedSuccess => 'Tạo tài khoản thành công';

  @override
  String get authEnterVerificationCode => 'Nhập mã xác minh';

  @override
  String get authOtpDialogDescriptionPart1 => 'Mã gồm 6 chữ số đã được gửi đến';

  @override
  String get authOtpDialogDescriptionPart2 =>
      '. Vui lòng nhập mã vào bên dưới để xác minh địa chỉ email của bạn.';

  @override
  String get authVerifyOtp => 'Xác minh OTP';

  @override
  String get authResendOtp => 'Gửi lại OTP';

  @override
  String get authSendingOtp => 'Đang gửi OTP...';

  @override
  String authOtpResendCountDown(int countdown) {
    return 'Gửi lại sau ${countdown}s';
  }

  @override
  String get errorGeneralTitle => 'Oops!';

  @override
  String get errorGeneralDescription => 'Đã xảy ra lỗi, vui lòng thử lại sau';

  @override
  String get errorEmailRequired => 'Email là bắt buộc';

  @override
  String get errorEmailInvalid => 'Email không hợp lệ';

  @override
  String get errorPasswordRequired => 'Mật khẩu là bắt buộc';

  @override
  String errorPasswordTooShort(int minLength) {
    return 'Mật khẩu phải có ít nhất $minLength ký tự';
  }

  @override
  String errorPasswordTooLong(int maxLength) {
    return 'Mật khẩu phải có tối đa $maxLength ký tự';
  }

  @override
  String get errorPasswordMissingLowercase =>
      'Mật khẩu phải chứa ít nhất một chữ cái thường';

  @override
  String get errorPasswordMissingUppercase =>
      'Mật khẩu phải chứa ít nhất một chữ cái hoa';

  @override
  String get errorPasswordMissingSpecial =>
      'Mật khẩu phải chứa ít nhất một ký tự đặc biệt';

  @override
  String errorEmailTooLong(int maxLength) {
    return 'Email phải có tối đa $maxLength ký tự';
  }

  @override
  String get errorUsernameRequired => 'Tên người dùng là bắt buộc';

  @override
  String get errorTitleRequired => 'Tiêu đề là bắt buộc';

  @override
  String get errorUsernameTooShort => 'Tên người dùng phải >= 6 ký tự';

  @override
  String errorUsernameTooLong(int maxLength) {
    return 'Tên người dùng phải có tối đa $maxLength ký tự';
  }

  @override
  String get errorConfirmPasswordRequired =>
      'Vui lòng xác nhận mật khẩu của bạn';

  @override
  String get errorPasswordsDoNotMatch => 'Mật khẩu không khớp';

  @override
  String get errorNewPasswordMustBeDifferent =>
      'Mật khẩu mới phải khác mật khẩu hiện tại';

  @override
  String get errorOtpRequired => 'Vui lòng nhập mã OTP';

  @override
  String get errorOtpSendFail => 'Có lỗi xảy ra khi gửi OTP';

  @override
  String get apiResErrInvalidInput => 'Dữ liệu không hợp lệ.';

  @override
  String get apiResErrInvalidPayload => 'Payload không hợp lệ.';

  @override
  String get apiResErrAuthAlreadyExists => 'Email đã được đăng ký.';

  @override
  String get apiResErrCaptchaFailed =>
      'Xác minh CAPTCHA thất bại, vui lòng thử lại.';

  @override
  String get apiResErrPasswordNotMatch => 'Mật khẩu cũ không khớp.';

  @override
  String get apiResErrUnauthorized => 'Chưa được xác thực.';

  @override
  String get apiResErrSignUpOtpExpired =>
      'Mã OTP đã hết hạn, vui lòng lấy mã mới.';

  @override
  String get apiResErrEmailOrPasswordIncorrect =>
      'Tên đăng nhập hoặc mật khẩu không đúng.';

  @override
  String get apiResErrForbidden => 'Không có quyền truy cập.';

  @override
  String get apiResErrAuthNotFound => 'Không tìm thấy thông tin xác thực.';

  @override
  String get apiResErrRefreshTokenNotFound => 'Không tìm thấy refresh token.';

  @override
  String get apiResErrSignUpOtpNotFound => 'Không tìm thấy mã OTP.';

  @override
  String get apiResErrRouteNotFound => 'Không tìm thấy endpoint yêu cầu.';

  @override
  String get apiResErrSignUpOtpAlreadyUsed => 'Mã OTP đã được sử dụng.';

  @override
  String get apiResErrFailedToCreateSignUpOtp =>
      'Không thể tạo OTP, vui lòng thử lại sau.';

  @override
  String get apiResErrDatabaseQuery =>
      'Lỗi truy vấn cơ sở dữ liệu, vui lòng thử lại.';

  @override
  String get apiResErrDatabaseIssue => 'Lỗi cơ sở dữ liệu, vui lòng thử lại.';

  @override
  String get apiResErrInternalServer => 'Đã xảy ra lỗi hệ thống.';

  @override
  String get apiResErrFailedToSendVerification =>
      'Gửi email xác minh thất bại, vui lòng thử lại sau.';

  @override
  String get apiResErrUserNotFound => 'Không tìm thấy người dùng.';

  @override
  String get apiResErrImageTooLarge => 'Ảnh vượt quá giới hạn 10mb.';

  @override
  String get apiResErrLivestreamUpdateAfterEnded =>
      'Không thể cập nhật, livestream đã kết thúc.';

  @override
  String get apiResErrLivestreamNotFound => 'Không tìm thấy livestream.';

  @override
  String get apiResErrVodNotFound => 'Không tìm thấy VOD.';

  @override
  String get apiResErrEndAlreadyEndedLivestream =>
      'Livestream đã kết thúc trước đó.';

  @override
  String get apiResErrVodCommentNotFound => 'Không tìm thấy bình luận.';

  @override
  String get apiResErrVodCommentCreateFailed => 'Không thể tạo bình luận.';

  @override
  String get apiResErrVodCommentAlreadyLiked => 'Bình luận đã được thích.';

  @override
  String get apiResErrVodCommentNotLiked => 'Bình luận chưa được thích.';

  @override
  String get apiResErrVodCommentDeleteFailed => 'Không thể xóa bình luận.';

  @override
  String get apiResSuccSentVerificationEmail =>
      'Email xác thực đã được gửi, vui lòng kiểm tra hộp thư';

  @override
  String get apiResSuccOk => 'Thành công';

  @override
  String get apiResSuccLogin => 'Đăng nhập thành công';

  @override
  String get apiResSuccSignUp => 'Đăng kí thành công';

  @override
  String get apiDefaultError => 'Đã xảy ra lỗi. Vui lòng thử lại.';

  @override
  String get fetchError =>
      'Có lỗi xảy ra khi tải dữ liệu, vui lòng thử lại sau.';

  @override
  String get themeLight => 'Sáng';

  @override
  String get themeDark => 'Tối';

  @override
  String get themeSystem => 'Hệ thống';

  @override
  String get settingsTitle => 'Cài đặt';

  @override
  String get settingsNeedToLogin => 'Bạn cần đăng nhập để điều chỉnh cài đặt';

  @override
  String get settingsNavProfile => 'Hồ sơ';

  @override
  String get settingsNavSecurity => 'Bảo mật';

  @override
  String get settingsNavStream => 'Stream';

  @override
  String get settingsNavVods => 'VODs';

  @override
  String get settingsProfileTitle => 'Cài đặt hồ sơ';

  @override
  String get settingsProfileDescription =>
      'Thay đổi thông tin nhận dạng cho tài khoản của bạn';

  @override
  String get settingsProfileUsername => 'Tên người dùng';

  @override
  String get settingsProfileDisplayName => 'Tên hiển thị';

  @override
  String get settingsProfileBio => 'Tiểu sử';

  @override
  String get settingsProfileUpdateSuccess =>
      'Cập nhật thông tin hồ sơ thành công!';

  @override
  String get settingsProfileUpdateBackground => 'Cập nhật ảnh nền';

  @override
  String get settingsProfileUpdateProfilePicture => 'Cập nhật ảnh đại diện';

  @override
  String get settingsSocialMediaTitle => 'Chỉnh sửa liên kết mạng xã hội';

  @override
  String get settingsSocialMediaDescription =>
      'Quản lý liên kết đến các hồ sơ mạng xã hội của bạn';

  @override
  String get settingsSocialMediaInvalidUrl =>
      'Đường dẫn không hợp lệ, vui lòng thử lại';

  @override
  String get settingsSocialMediaErrEmptyUrl => 'Đường dẫn không được để trống';

  @override
  String get settingsThemesTitle => 'Giao diện';

  @override
  String get settingsThemesDescription => 'Tùy chỉnh giao diện của ứng dụng';

  @override
  String get settingsLanguageTitle => 'Ngôn ngữ';

  @override
  String get settingsLanguageDescription => 'Chọn ngôn ngữ ưa thích của bạn';

  @override
  String get settingsSaveChanges => 'Lưu thay đổi';

  @override
  String settingsFileSizeExceeds(int size) {
    return 'Kích thước tệp vượt quá $size MB';
  }

  @override
  String get settingsSecurityApiKeyLabel => 'API Key của bạn';

  @override
  String get settingsSecurityApiKeyGenerate => 'Tạo API Key mới';

  @override
  String get settingsSecurityApiKeyCopied =>
      'Đã sao chép API Key vào clipboard';

  @override
  String get settingsSecurityContactTitle => 'Liên hệ';

  @override
  String get settingsSecurityContactDescription =>
      'Nơi chúng tôi gửi tin nhắn quan trọng về tài khoản của bạn';

  @override
  String get settingsSecurityContactEmail => 'Email';

  @override
  String get settingsSecurityContactPhone => 'Số điện thoại';

  @override
  String get settingsSecurityContactAddPhone => 'Thêm số điện thoại';

  @override
  String get settingsSecurityContactPhoneInvalid =>
      'Số điện thoại không hợp lệ, vui lòng thử lại';

  @override
  String get settingsSecurityTitle => 'Bảo mật';

  @override
  String get settingsSecurityDescription => 'Giữ tài khoản của bạn an toàn';

  @override
  String get settingsSecurityPasswordTitle => 'Mật khẩu';

  @override
  String get settingsSecurityPasswordChange => 'Thay đổi mật khẩu';

  @override
  String get settingsSecurityPasswordDialogTitle => 'Thay đổi mật khẩu';

  @override
  String get settingsSecurityPasswordLocalDescription =>
      'Tăng cường bảo mật với mật khẩu mạnh.';

  @override
  String get settingsSecurityPasswordUpdatedSuccess =>
      'Cập nhật mật khẩu thành công';

  @override
  String get settingsSecurityPasswordFormCurrentLabel => 'Mật khẩu hiện tại';

  @override
  String get settingsSecurityPasswordFormCurrentPlaceholder =>
      'Nhập mật khẩu hiện tại của bạn';

  @override
  String get settingsSecurityPasswordFormNewLabel => 'Mật khẩu mới';

  @override
  String get settingsSecurityPasswordFormNewPlaceholder => 'Nhập mật khẩu mới';

  @override
  String get settingsSecurityPasswordFormConfirmLabel =>
      'Xác nhận mật khẩu mới';

  @override
  String get settingsSecurityPasswordFormConfirmPlaceholder =>
      'Xác nhận mật khẩu mới của bạn';

  @override
  String get settingsSecurityPasswordFormSubmit => 'Xác nhận';

  @override
  String get messagesTitle => 'Tin nhắn';

  @override
  String get messagesSelectConversation => 'Chọn cuộc trò chuyện';

  @override
  String get messagesLoginRequired => 'Vui lòng đăng nhập để xem tin nhắn.';

  @override
  String get messagesUnknown => 'Ẩn danh';

  @override
  String get messagesGroup => 'Nhóm';

  @override
  String get messagesOnline => 'Đang trực tuyến';

  @override
  String get messagesOffline => 'Ngoại tuyến';

  @override
  String messagesMembersCount(int count) {
    return '$count thành viên';
  }

  @override
  String get messagesNoConversationsYet => 'Chưa có cuộc trò chuyện';

  @override
  String get messagesLoadMore => 'Tải thêm';

  @override
  String get messagesNoMessagesYet => 'Chưa có tin nhắn';

  @override
  String get messagesNewConversation => 'Cuộc trò chuyện mới';

  @override
  String get messagesDirectMessage => 'Tin nhắn trực tiếp';

  @override
  String get messagesGroupNamePlaceholder => 'Tên nhóm (tùy chọn)';

  @override
  String get messagesSearchUsersPlaceholder => 'Tìm người dùng...';

  @override
  String get messagesSearching => 'Đang tìm...';

  @override
  String get messagesCreating => 'Đang tạo...';

  @override
  String get messagesCreate => 'Tạo';

  @override
  String get messagesSentAnImage => 'Đã gửi một ảnh';

  @override
  String messagesSentImagesCount(int count) {
    return 'Đã gửi $count ảnh';
  }

  @override
  String get messagesUploadFailed => 'Tải lên thất bại';

  @override
  String get messagesPlaceholderTypeMessage => 'Nhập tin nhắn...';

  @override
  String get messagesMessageDeleted => 'Tin nhắn đã bị xóa';

  @override
  String get messagesBeginningOfConversation => 'Đầu cuộc trò chuyện';

  @override
  String messagesTypingOne(String name) {
    return '$name đang nhập...';
  }

  @override
  String get notificationsTitle => 'Thông báo';

  @override
  String get notificationsMarkAllAsRead => 'Đánh dấu tất cả đã đọc';

  @override
  String get notificationsLoading => 'Đang tải...';

  @override
  String get notificationsNoNotifications => 'Không có thông báo';

  @override
  String get notificationsViewAll => 'Xem tất cả thông báo';

  @override
  String get notificationsNoNotificationsYet => 'Chưa có thông báo.';

  @override
  String get notificationsPleaseLogIn => 'Vui lòng đăng nhập để xem thông báo.';

  @override
  String get notificationsView => 'Xem';

  @override
  String get notificationsMarkAsRead => 'Đánh dấu đã đọc';

  @override
  String get notificationsDelete => 'Xóa';

  @override
  String get notificationsLoadMore => 'Tải thêm';

  @override
  String get commentsTitle => 'Bình luận';

  @override
  String get commentsWriteComment => 'Viết bình luận...';

  @override
  String get commentsWriteReply => 'Viết trả lời...';

  @override
  String get commentsPost => 'Đăng';

  @override
  String get commentsReply => 'Trả lời';

  @override
  String get commentsDelete => 'Xóa';

  @override
  String get commentsDeletedComment => 'Bình luận này đã bị xóa.';

  @override
  String commentsViewReplies(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: 'Xem $count trả lời',
    );
    return '$_temp0';
  }

  @override
  String get commentsLoadMoreReplies => 'Tải thêm trả lời';

  @override
  String get commentsNoComments =>
      'Chưa có bình luận nào. Hãy là người đầu tiên bình luận!';

  @override
  String get commentsLoginToComment => 'Đăng nhập để bình luận.';

  @override
  String get commentsOwner => 'Chủ kênh';

  @override
  String get commentsYou => 'Bạn';

  @override
  String get commentsDeleteConfirmTitle => 'Xóa bình luận?';

  @override
  String get commentsDeleteConfirmDescription =>
      'Hành động này không thể hoàn tác.';

  @override
  String get commentsLike => 'Thích';

  @override
  String get commentsUnlike => 'Bỏ thích';

  @override
  String commentsCharRemaining(int count) {
    return 'Còn $count ký tự';
  }

  @override
  String get usersChatTitle => 'Trò chuyện';

  @override
  String get usersChatJoined => 'đã tham gia phòng chat';

  @override
  String get usersChatLeft => 'đã rời phòng chat';

  @override
  String get usersChatPlaceholderLogin => 'Đăng nhập để bắt đầu nhắn tin';

  @override
  String get usersChatPlaceholderTyping => 'Nhập tin nhắn...';

  @override
  String get usersOffline => 'Ngoại tuyến';

  @override
  String get usersProfileAbout => 'Giới thiệu';

  @override
  String usersProfileFollowers(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: 'người theo dõi',
    );
    return '$_temp0';
  }

  @override
  String get usersProfileJoinedPrefix => 'Tham gia';

  @override
  String get usersProfileRecentStreams => 'Luồng gần đây';

  @override
  String get accessibilityOpenMenu => 'Mở menu';

  @override
  String get navHome => 'Trang chủ';

  @override
  String get navMessages => 'Tin nhắn';

  @override
  String get navNotifications => 'Thông báo';

  @override
  String get navSettings => 'Cài đặt';

  @override
  String get homeTabLivestreams => 'Đang phát';

  @override
  String get homeTabVods => 'Videos';

  @override
  String get homeNoContent => 'Chưa có nội dung';

  @override
  String get homeNoContentDescription =>
      'Vui lòng quay lại sau để xem nội dung mới';

  @override
  String homeViewerCount(int count) {
    return '$count người xem';
  }

  @override
  String homeViewCount(int count) {
    return '$count lượt xem';
  }

  @override
  String get settingsStreamTitle => 'Cài đặt Stream';

  @override
  String get settingsStreamDescription => 'Cấu hình cài đặt livestream của bạn';

  @override
  String get settingsStreamStreamTitle => 'Tiêu đề Stream';

  @override
  String get settingsStreamStreamDescription => 'Mô tả Stream';

  @override
  String get settingsStreamThumbnail => 'Ảnh thu nhỏ';

  @override
  String get settingsStreamChangeThumbnail => 'Đổi ảnh thu nhỏ';

  @override
  String get settingsStreamUpdateSuccess =>
      'Cập nhật thông tin livestream thành công!';

  @override
  String get settingsVodsTitle => 'VODs của bạn';

  @override
  String get settingsVodsDescription => 'Quản lý video theo yêu cầu của bạn';

  @override
  String get settingsVodsNoVods => 'Chưa có VOD nào';

  @override
  String get settingsVodsNoVodsDescription =>
      'Các bản ghi livestream của bạn sẽ xuất hiện ở đây';

  @override
  String get settingsVodsEditTitle => 'Chỉnh sửa VOD';

  @override
  String get settingsVodsDeleteTitle => 'Xóa VOD';

  @override
  String get settingsVodsDeleteConfirm =>
      'Bạn có chắc muốn xóa VOD này? Hành động này không thể hoàn tác.';

  @override
  String get settingsVodsDeleteSuccess => 'Đã xóa VOD thành công';

  @override
  String get settingsVodsUpdateSuccess => 'Đã cập nhật VOD thành công';

  @override
  String get settingsVodsVisibility => 'Hiển thị';

  @override
  String get settingsVodsPublic => 'Công khai';

  @override
  String get settingsVodsPrivate => 'Riêng tư';

  @override
  String get retry => 'Thử lại';

  @override
  String get delete => 'Xóa';

  @override
  String get confirm => 'Xác nhận';

  @override
  String get edit => 'Sửa';

  @override
  String get copyToClipboard => 'Sao chép vào clipboard';

  @override
  String get livestreamWatchLive => 'Xem trực tiếp';

  @override
  String get livestreamVideoError => 'Không thể tải luồng video';

  @override
  String get vodWatchVideo => 'Xem';

  @override
  String get vodVideoError => 'Không thể tải video';
}
