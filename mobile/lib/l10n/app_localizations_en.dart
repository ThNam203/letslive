// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for English (`en`).
class AppLocalizationsEn extends AppLocalizations {
  AppLocalizationsEn([String locale = 'en']) : super(locale);

  @override
  String get appTitle => 'Let\'s Live';

  @override
  String get loading => 'Loading...';

  @override
  String get setting => 'Setting';

  @override
  String get save => 'Save';

  @override
  String get cancel => 'Cancel';

  @override
  String get or => 'or';

  @override
  String get goHome => 'Go home';

  @override
  String get pageNotFound => 'We couldn\'t find the page you were looking for';

  @override
  String get saveChanges => 'Save changes';

  @override
  String get username => 'Username';

  @override
  String get livestreams => 'Livestreams';

  @override
  String get channels => 'Channels';

  @override
  String get videos => 'Videos';

  @override
  String get home => 'Home';

  @override
  String get searchUsers => 'Search users';

  @override
  String get gotIt => 'Got it';

  @override
  String get howToLivestream => 'How to livestream';

  @override
  String get startYourLivestream => 'Start your livestream';

  @override
  String get liveStreaming => 'Live Streaming';

  @override
  String get follow => 'Follow';

  @override
  String get unfollow => 'Unfollow';

  @override
  String followersWithCount(int count) {
    return 'Followers: $count';
  }

  @override
  String get joined => 'Joined';

  @override
  String get showMore => 'Show more';

  @override
  String get pleaseWaitWhileLoading => 'Please wait while we load the content';

  @override
  String get noLivestreams => 'No livestreams';

  @override
  String get noLivestreamsDescription =>
      'There is currently no one streaming, check back later or explore our video on demand content';

  @override
  String get noVideos => 'No Videos Available';

  @override
  String get noVideosDescription =>
      'There are currently no videos available. Check back later for new content';

  @override
  String get searching => 'Searching...';

  @override
  String get noUsersFound => 'No users found';

  @override
  String get noDescription => 'No description';

  @override
  String get add => 'Add';

  @override
  String get bio => 'Bio';

  @override
  String startedAt(String time) {
    return 'Started $time';
  }

  @override
  String get live => 'Live';

  @override
  String get following => 'Following';

  @override
  String get recommended => 'Recommended';

  @override
  String timeSecondsAgo(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$count seconds ago',
      one: '$count second ago',
    );
    return '$_temp0';
  }

  @override
  String timeMinutesAgo(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$count minutes ago',
      one: '$count minute ago',
    );
    return '$_temp0';
  }

  @override
  String timeHoursAgo(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$count hours ago',
      one: '$count hour ago',
    );
    return '$_temp0';
  }

  @override
  String timeDaysAgo(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$count days ago',
      one: '$count day ago',
    );
    return '$_temp0';
  }

  @override
  String timeWeeksAgo(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$count weeks ago',
      one: '$count week ago',
    );
    return '$_temp0';
  }

  @override
  String timeMonthsAgo(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$count months ago',
      one: '$count month ago',
    );
    return '$_temp0';
  }

  @override
  String timeYearsAgo(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$count years ago',
      one: '$count year ago',
    );
    return '$_temp0';
  }

  @override
  String get authLogin => 'Log in';

  @override
  String get authLoginTitle => 'Welcome back!';

  @override
  String get authLoginSubtitle => 'Gain access to the world right now.';

  @override
  String get authEmail => 'Email';

  @override
  String get authPassword => 'Password';

  @override
  String get authForgotPassword => 'Forgot password?';

  @override
  String get authSignup => 'Sign up';

  @override
  String get authSignupTitle => 'Welcome! Sign up for a new world?';

  @override
  String get authSignupSubtitle => 'Choose a method below to begin';

  @override
  String get authLogout => 'Log out';

  @override
  String get authNoAccount => 'Don\'t have an account?';

  @override
  String get authHaveAccount => 'Already have an account?';

  @override
  String get authConfirmPassword => 'Confirm password';

  @override
  String get authAccountCreatedSuccess => 'Account created successfully';

  @override
  String get authEnterVerificationCode => 'Enter Verification Code';

  @override
  String get authOtpDialogDescriptionPart1 => 'A 6-digit code has been sent to';

  @override
  String get authOtpDialogDescriptionPart2 =>
      '. Please enter it below to verify your email address.';

  @override
  String get authVerifyOtp => 'Verify OTP';

  @override
  String get authResendOtp => 'Resend OTP';

  @override
  String get authSendingOtp => 'Sending OTP...';

  @override
  String authOtpResendCountDown(int countdown) {
    return 'Resend in ${countdown}s';
  }

  @override
  String get authContinueWithGoogle => 'Continue with Google';

  @override
  String get authGoogleSignInCancelled => 'Google sign-in was cancelled';

  @override
  String get authGoogleSignInFailed => 'Google sign-in failed';

  @override
  String get authEmailHint => 'Enter your email';

  @override
  String get authPasswordHint => 'Enter your password';

  @override
  String get authUsernameHint => 'Choose a username';

  @override
  String get authCreatePasswordHint => 'Create a password';

  @override
  String get authConfirmPasswordHint => 'Re-enter your password';

  @override
  String get errorGeneralTitle => 'Oops!';

  @override
  String get errorGeneralDescription =>
      'Something has gone wrong, please try again later';

  @override
  String get errorEmailRequired => 'Email is required';

  @override
  String get errorEmailInvalid => 'Email is invalid';

  @override
  String get errorPasswordRequired => 'Password is required';

  @override
  String errorPasswordTooShort(int minLength) {
    return 'Password must be at least $minLength characters';
  }

  @override
  String errorPasswordTooLong(int maxLength) {
    return 'Password must be at most $maxLength characters';
  }

  @override
  String get errorPasswordMissingLowercase =>
      'Password must contain at least one lowercase letter';

  @override
  String get errorPasswordMissingUppercase =>
      'Password must contain at least one uppercase letter';

  @override
  String get errorPasswordMissingSpecial =>
      'Password must contain at least one special character';

  @override
  String errorEmailTooLong(int maxLength) {
    return 'Email must be at most $maxLength characters';
  }

  @override
  String get errorUsernameRequired => 'Username is required';

  @override
  String get errorTitleRequired => 'Title is required';

  @override
  String get errorUsernameTooShort => 'Username must be >= 6 characters';

  @override
  String errorUsernameTooLong(int maxLength) {
    return 'Username must be at most $maxLength characters';
  }

  @override
  String get errorConfirmPasswordRequired => 'Please confirm your password';

  @override
  String get errorPasswordsDoNotMatch => 'Passwords do not match';

  @override
  String get errorNewPasswordMustBeDifferent =>
      'New password must be different from current password';

  @override
  String get errorOtpRequired => 'Please enter OTP code';

  @override
  String get errorOtpSendFail => 'Failed to send OTP';

  @override
  String get apiResErrInvalidInput => 'Input invalid.';

  @override
  String get apiResErrInvalidPayload => 'Payload invalid.';

  @override
  String get apiResErrAuthAlreadyExists => 'Email has already been registered.';

  @override
  String get apiResErrCaptchaFailed =>
      'Failed to verify CAPTCHA, please try again.';

  @override
  String get apiResErrPasswordNotMatch => 'Old password does not match.';

  @override
  String get apiResErrUnauthorized => 'Unauthorized.';

  @override
  String get apiResErrSignUpOtpExpired =>
      'OTP code has expired, please issue a new one.';

  @override
  String get apiResErrEmailOrPasswordIncorrect =>
      'Username or password incorrect.';

  @override
  String get apiResErrForbidden => 'Forbidden.';

  @override
  String get apiResErrAuthNotFound => 'Authentication credentials not found.';

  @override
  String get apiResErrRefreshTokenNotFound => 'Refresh token not found.';

  @override
  String get apiResErrSignUpOtpNotFound => 'OTP code not found.';

  @override
  String get apiResErrRouteNotFound => 'Requested endpoint not found.';

  @override
  String get apiResErrSignUpOtpAlreadyUsed => 'The OTP has already been used.';

  @override
  String get apiResErrFailedToCreateSignUpOtp =>
      'Failed to generate the OTP, please try again later.';

  @override
  String get apiResErrDatabaseQuery =>
      'Error querying database, please try again.';

  @override
  String get apiResErrDatabaseIssue => 'Database issue, please try again.';

  @override
  String get apiResErrInternalServer => 'Something went wrong.';

  @override
  String get apiResErrFailedToSendVerification =>
      'Failed to send email verification, please try again later.';

  @override
  String get apiResErrUserNotFound => 'User not found.';

  @override
  String get apiResErrImageTooLarge => 'Image exceeds 10mb limit.';

  @override
  String get apiResErrLivestreamUpdateAfterEnded =>
      'Failed to update, the livestream has ended.';

  @override
  String get apiResErrLivestreamNotFound => 'Livestream not found.';

  @override
  String get apiResErrVodNotFound => 'VOD not found.';

  @override
  String get apiResErrEndAlreadyEndedLivestream =>
      'The livestream has already been ended.';

  @override
  String get apiResErrVodCommentNotFound => 'Comment not found.';

  @override
  String get apiResErrVodCommentCreateFailed => 'Failed to create comment.';

  @override
  String get apiResErrVodCommentAlreadyLiked => 'Comment already liked.';

  @override
  String get apiResErrVodCommentNotLiked => 'Comment has not been liked.';

  @override
  String get apiResErrVodCommentDeleteFailed => 'Failed to delete comment.';

  @override
  String get apiResSuccSentVerificationEmail =>
      'Verification email sent, please check your inbox';

  @override
  String get apiResSuccOk => 'Success';

  @override
  String get apiResSuccLogin => 'Login successfully';

  @override
  String get apiResSuccSignUp => 'Sign up successfully';

  @override
  String get apiDefaultError => 'Something went wrong. Please try again.';

  @override
  String get fetchError => 'Failed to fetch data, please try again.';

  @override
  String get themeLight => 'Light';

  @override
  String get themeDark => 'Dark';

  @override
  String get themeSystem => 'System';

  @override
  String get settingsTitle => 'Settings';

  @override
  String get settingsNeedToLogin =>
      'You need to log in to configure your settings';

  @override
  String get settingsNavProfile => 'Profile';

  @override
  String get settingsNavSecurity => 'Security';

  @override
  String get settingsNavStream => 'Stream';

  @override
  String get settingsNavVods => 'VODs';

  @override
  String get settingsProfileTitle => 'Profile Settings';

  @override
  String get settingsProfileDescription =>
      'Change identifying details for your account';

  @override
  String get settingsProfileUsername => 'Username';

  @override
  String get settingsProfileDisplayName => 'Display Name';

  @override
  String get settingsProfileBio => 'Bio';

  @override
  String get settingsProfileUpdateSuccess =>
      'Profile information updated successfully!';

  @override
  String get settingsProfileUpdateBackground => 'Update background';

  @override
  String get settingsProfileUpdateProfilePicture => 'Update profile picture';

  @override
  String get settingsSocialMediaTitle => 'Edit Social Media Links';

  @override
  String get settingsSocialMediaDescription =>
      'Manage links to your social media profiles';

  @override
  String get settingsSocialMediaInvalidUrl => 'Invalid link, please try again';

  @override
  String get settingsSocialMediaErrEmptyUrl => 'Link can\'t be empty';

  @override
  String get settingsThemesTitle => 'Theme';

  @override
  String get settingsThemesDescription =>
      'Customize the look and feel of the app';

  @override
  String get settingsLanguageTitle => 'Language';

  @override
  String get settingsLanguageDescription => 'Select your preferred language';

  @override
  String get settingsSaveChanges => 'Save Changes';

  @override
  String settingsFileSizeExceeds(int size) {
    return 'File size exceeds $size MB';
  }

  @override
  String get settingsSecurityApiKeyLabel => 'Your API Key';

  @override
  String get settingsSecurityApiKeyGenerate => 'Generate New API Key';

  @override
  String get settingsSecurityApiKeyCopied => 'API Key copied to clipboard';

  @override
  String get settingsSecurityContactTitle => 'Contact';

  @override
  String get settingsSecurityContactDescription =>
      'Where we send important messages about your account';

  @override
  String get settingsSecurityContactEmail => 'Email';

  @override
  String get settingsSecurityContactPhone => 'Phone Number';

  @override
  String get settingsSecurityContactAddPhone => 'Add a number';

  @override
  String get settingsSecurityContactPhoneInvalid =>
      'Invalid phone number, please try again';

  @override
  String get settingsSecurityTitle => 'Security';

  @override
  String get settingsSecurityDescription => 'Keep your account safe and sound';

  @override
  String get settingsSecurityPasswordTitle => 'Password';

  @override
  String get settingsSecurityPasswordChange => 'Change password';

  @override
  String get settingsSecurityPasswordDialogTitle => 'Change Password';

  @override
  String get settingsSecurityPasswordLocalDescription =>
      'Improve your security with a strong password.';

  @override
  String get settingsSecurityPasswordUpdatedSuccess =>
      'Password updated successfully';

  @override
  String get settingsSecurityPasswordFormCurrentLabel => 'Current Password';

  @override
  String get settingsSecurityPasswordFormCurrentPlaceholder =>
      'Enter your current password';

  @override
  String get settingsSecurityPasswordFormNewLabel => 'New Password';

  @override
  String get settingsSecurityPasswordFormNewPlaceholder =>
      'Enter your new password';

  @override
  String get settingsSecurityPasswordFormConfirmLabel => 'Confirm New Password';

  @override
  String get settingsSecurityPasswordFormConfirmPlaceholder =>
      'Confirm your new password';

  @override
  String get settingsSecurityPasswordFormSubmit => 'Confirm';

  @override
  String get messagesTitle => 'Messages';

  @override
  String get messagesSelectConversation => 'Select a conversation';

  @override
  String get messagesLoginRequired => 'Please log in to view messages.';

  @override
  String get messagesUnknown => 'Unknown';

  @override
  String get messagesGroup => 'Group';

  @override
  String get messagesOnline => 'Online';

  @override
  String get messagesOffline => 'Offline';

  @override
  String messagesMembersCount(int count) {
    return '$count members';
  }

  @override
  String get messagesNoConversationsYet => 'No conversations yet';

  @override
  String get messagesLoadMore => 'Load more';

  @override
  String get messagesNoMessagesYet => 'No messages yet';

  @override
  String get messagesNewConversation => 'New Conversation';

  @override
  String get messagesDirectMessage => 'Direct Message';

  @override
  String get messagesGroupNamePlaceholder => 'Group name (optional)';

  @override
  String get messagesSearchUsersPlaceholder => 'Search users...';

  @override
  String get messagesSearching => 'Searching...';

  @override
  String get messagesCreating => 'Creating...';

  @override
  String get messagesCreate => 'Create';

  @override
  String get messagesSentAnImage => 'Sent an image';

  @override
  String messagesSentImagesCount(int count) {
    return 'Sent $count images';
  }

  @override
  String get messagesUploadFailed => 'Failed to upload files';

  @override
  String get messagesPlaceholderTypeMessage => 'Type a message...';

  @override
  String get messagesMessageDeleted => 'This message was deleted';

  @override
  String get messagesBeginningOfConversation => 'Beginning of conversation';

  @override
  String messagesTypingOne(String name) {
    return '$name is typing...';
  }

  @override
  String get notificationsTitle => 'Notifications';

  @override
  String get notificationsMarkAllAsRead => 'Mark all as read';

  @override
  String get notificationsLoading => 'Loading...';

  @override
  String get notificationsNoNotifications => 'No notifications';

  @override
  String get notificationsViewAll => 'View all notifications';

  @override
  String get notificationsNoNotificationsYet => 'No notifications yet.';

  @override
  String get notificationsPleaseLogIn => 'Please log in to view notifications.';

  @override
  String get notificationsView => 'View';

  @override
  String get notificationsMarkAsRead => 'Mark as read';

  @override
  String get notificationsDelete => 'Delete';

  @override
  String get notificationsLoadMore => 'Load more';

  @override
  String get commentsTitle => 'Comments';

  @override
  String get commentsWriteComment => 'Write a comment...';

  @override
  String get commentsWriteReply => 'Write a reply...';

  @override
  String get commentsPost => 'Post';

  @override
  String get commentsReply => 'Reply';

  @override
  String get commentsDelete => 'Delete';

  @override
  String get commentsDeletedComment => 'This comment has been deleted.';

  @override
  String commentsViewReplies(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: 'View $count replies',
      one: 'View $count reply',
    );
    return '$_temp0';
  }

  @override
  String get commentsLoadMoreReplies => 'Load more replies';

  @override
  String get commentsNoComments => 'No comments yet. Be the first to comment!';

  @override
  String get commentsLoginToComment => 'Log in to leave a comment.';

  @override
  String get commentsOwner => 'Owner';

  @override
  String get commentsYou => 'You';

  @override
  String get commentsDeleteConfirmTitle => 'Delete comment?';

  @override
  String get commentsDeleteConfirmDescription =>
      'This action cannot be undone.';

  @override
  String get commentsLike => 'Like';

  @override
  String get commentsUnlike => 'Unlike';

  @override
  String commentsCharRemaining(int count) {
    return '$count characters remaining';
  }

  @override
  String get usersChatTitle => 'Chat';

  @override
  String get usersChatJoined => 'joined the chat';

  @override
  String get usersChatLeft => 'left the chat';

  @override
  String get usersChatPlaceholderLogin => 'Login to start messaging';

  @override
  String get usersChatPlaceholderTyping => 'Type a message...';

  @override
  String get usersOffline => 'Offline';

  @override
  String get usersProfileAbout => 'About';

  @override
  String usersProfileFollowers(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: 'followers',
      one: 'follower',
    );
    return '$_temp0';
  }

  @override
  String get usersProfileJoinedPrefix => 'Joined';

  @override
  String get usersProfileRecentStreams => 'Recent Streams';

  @override
  String get accessibilityOpenMenu => 'Open menu';

  @override
  String get navHome => 'Home';

  @override
  String get navMessages => 'Messages';

  @override
  String get navNotifications => 'Notifications';

  @override
  String get navSettings => 'Settings';

  @override
  String get homeTabLivestreams => 'Live Now';

  @override
  String get homeTabVods => 'Videos';

  @override
  String get homeNoContent => 'Nothing here yet';

  @override
  String get homeNoContentDescription => 'Check back later for new content';

  @override
  String homeViewerCount(int count) {
    return '$count viewers';
  }

  @override
  String homeViewCount(int count) {
    return '$count views';
  }

  @override
  String get settingsStreamTitle => 'Stream Settings';

  @override
  String get settingsStreamDescription => 'Configure your livestream settings';

  @override
  String get settingsStreamStreamTitle => 'Stream Title';

  @override
  String get settingsStreamStreamDescription => 'Stream Description';

  @override
  String get settingsStreamThumbnail => 'Thumbnail';

  @override
  String get settingsStreamChangeThumbnail => 'Change thumbnail';

  @override
  String get settingsStreamUpdateSuccess =>
      'Livestream information updated successfully!';

  @override
  String get settingsVodsTitle => 'Your VODs';

  @override
  String get settingsVodsDescription => 'Manage your videos on demand';

  @override
  String get settingsVodsNoVods => 'No VODs yet';

  @override
  String get settingsVodsNoVodsDescription =>
      'Your recorded livestreams will appear here';

  @override
  String get settingsVodsEditTitle => 'Edit VOD';

  @override
  String get settingsVodsDeleteTitle => 'Delete VOD';

  @override
  String get settingsVodsDeleteConfirm =>
      'Are you sure you want to delete this VOD? This action cannot be undone.';

  @override
  String get settingsVodsDeleteSuccess => 'VOD deleted successfully';

  @override
  String get settingsVodsUpdateSuccess => 'VOD updated successfully';

  @override
  String get settingsVodsVisibility => 'Visibility';

  @override
  String get settingsVodsPublic => 'Public';

  @override
  String get settingsVodsPrivate => 'Private';

  @override
  String get retry => 'Retry';

  @override
  String get delete => 'Delete';

  @override
  String get confirm => 'Confirm';

  @override
  String get edit => 'Edit';

  @override
  String get copyToClipboard => 'Copy to clipboard';

  @override
  String get livestreamWatchLive => 'Watch Live';

  @override
  String get livestreamVideoError => 'Unable to load video stream';

  @override
  String get vodWatchVideo => 'Watch';

  @override
  String get vodVideoError => 'Unable to load video';

  @override
  String get uploadVideoTitle => 'Upload Video';

  @override
  String get uploadSelectVideo => 'Select a video';

  @override
  String get uploadVideoSelected => 'Video selected';

  @override
  String get uploadTapToChange => 'Tap to change';

  @override
  String get uploadSupportedFormats => 'MP4, MOV, AVI, MKV, WebM';

  @override
  String get uploadTitle => 'Title';

  @override
  String get uploadTitleHint => 'Enter video title';

  @override
  String get uploadDescription => 'Description';

  @override
  String get uploadDescriptionHint => 'Enter video description';

  @override
  String get uploadButton => 'Upload';

  @override
  String get uploadUploading => 'Uploading Video';

  @override
  String get uploadProcessing => 'Processing';

  @override
  String get uploadProcessingOnServer => 'Processing on server...';

  @override
  String get uploadComplete => 'Upload complete!';

  @override
  String get uploadDone => 'Done';

  @override
  String get uploadFailed => 'Upload Failed';

  @override
  String get uploadFailedGeneric => 'Upload failed';

  @override
  String get uploadDismiss => 'Dismiss';

  @override
  String get uploadClose => 'Close';

  @override
  String uploadManagerUploading(int count) {
    return 'Uploading $count video(s)';
  }

  @override
  String uploadManagerFailed(int count) {
    return '$count upload(s) failed';
  }

  @override
  String uploadManagerComplete(int count) {
    return '$count upload(s) complete';
  }

  @override
  String get uploadManagerClear => 'Clear';

  @override
  String get uploadManagerWaiting => 'Waiting...';

  @override
  String get uploadManagerQueued => 'Added to upload queue';
}
