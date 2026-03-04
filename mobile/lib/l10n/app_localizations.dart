import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:intl/intl.dart' as intl;

import 'app_localizations_en.dart';
import 'app_localizations_vi.dart';

// ignore_for_file: type=lint

/// Callers can lookup localized strings with an instance of AppLocalizations
/// returned by `AppLocalizations.of(context)`.
///
/// Applications need to include `AppLocalizations.delegate()` in their app's
/// `localizationDelegates` list, and the locales they support in the app's
/// `supportedLocales` list. For example:
///
/// ```dart
/// import 'l10n/app_localizations.dart';
///
/// return MaterialApp(
///   localizationsDelegates: AppLocalizations.localizationsDelegates,
///   supportedLocales: AppLocalizations.supportedLocales,
///   home: MyApplicationHome(),
/// );
/// ```
///
/// ## Update pubspec.yaml
///
/// Please make sure to update your pubspec.yaml to include the following
/// packages:
///
/// ```yaml
/// dependencies:
///   # Internationalization support.
///   flutter_localizations:
///     sdk: flutter
///   intl: any # Use the pinned version from flutter_localizations
///
///   # Rest of dependencies
/// ```
///
/// ## iOS Applications
///
/// iOS applications define key application metadata, including supported
/// locales, in an Info.plist file that is built into the application bundle.
/// To configure the locales supported by your app, you’ll need to edit this
/// file.
///
/// First, open your project’s ios/Runner.xcworkspace Xcode workspace file.
/// Then, in the Project Navigator, open the Info.plist file under the Runner
/// project’s Runner folder.
///
/// Next, select the Information Property List item, select Add Item from the
/// Editor menu, then select Localizations from the pop-up menu.
///
/// Select and expand the newly-created Localizations item then, for each
/// locale your application supports, add a new item and select the locale
/// you wish to add from the pop-up menu in the Value field. This list should
/// be consistent with the languages listed in the AppLocalizations.supportedLocales
/// property.
abstract class AppLocalizations {
  AppLocalizations(String locale)
    : localeName = intl.Intl.canonicalizedLocale(locale.toString());

  final String localeName;

  static AppLocalizations of(BuildContext context) {
    return Localizations.of<AppLocalizations>(context, AppLocalizations)!;
  }

  static const LocalizationsDelegate<AppLocalizations> delegate =
      _AppLocalizationsDelegate();

  /// A list of this localizations delegate along with the default localizations
  /// delegates.
  ///
  /// Returns a list of localizations delegates containing this delegate along with
  /// GlobalMaterialLocalizations.delegate, GlobalCupertinoLocalizations.delegate,
  /// and GlobalWidgetsLocalizations.delegate.
  ///
  /// Additional delegates can be added by appending to this list in
  /// MaterialApp. This list does not have to be used at all if a custom list
  /// of delegates is preferred or required.
  static const List<LocalizationsDelegate<dynamic>> localizationsDelegates =
      <LocalizationsDelegate<dynamic>>[
        delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
      ];

  /// A list of this localizations delegate's supported locales.
  static const List<Locale> supportedLocales = <Locale>[
    Locale('en'),
    Locale('vi'),
  ];

  /// No description provided for @appTitle.
  ///
  /// In en, this message translates to:
  /// **'Let\'s Live'**
  String get appTitle;

  /// No description provided for @loading.
  ///
  /// In en, this message translates to:
  /// **'Loading...'**
  String get loading;

  /// No description provided for @setting.
  ///
  /// In en, this message translates to:
  /// **'Setting'**
  String get setting;

  /// No description provided for @save.
  ///
  /// In en, this message translates to:
  /// **'Save'**
  String get save;

  /// No description provided for @cancel.
  ///
  /// In en, this message translates to:
  /// **'Cancel'**
  String get cancel;

  /// No description provided for @or.
  ///
  /// In en, this message translates to:
  /// **'or'**
  String get or;

  /// No description provided for @goHome.
  ///
  /// In en, this message translates to:
  /// **'Go home'**
  String get goHome;

  /// No description provided for @pageNotFound.
  ///
  /// In en, this message translates to:
  /// **'We couldn\'t find the page you were looking for'**
  String get pageNotFound;

  /// No description provided for @saveChanges.
  ///
  /// In en, this message translates to:
  /// **'Save changes'**
  String get saveChanges;

  /// No description provided for @username.
  ///
  /// In en, this message translates to:
  /// **'Username'**
  String get username;

  /// No description provided for @livestreams.
  ///
  /// In en, this message translates to:
  /// **'Livestreams'**
  String get livestreams;

  /// No description provided for @channels.
  ///
  /// In en, this message translates to:
  /// **'Channels'**
  String get channels;

  /// No description provided for @videos.
  ///
  /// In en, this message translates to:
  /// **'Videos'**
  String get videos;

  /// No description provided for @home.
  ///
  /// In en, this message translates to:
  /// **'Home'**
  String get home;

  /// No description provided for @searchUsers.
  ///
  /// In en, this message translates to:
  /// **'Search users'**
  String get searchUsers;

  /// No description provided for @gotIt.
  ///
  /// In en, this message translates to:
  /// **'Got it'**
  String get gotIt;

  /// No description provided for @howToLivestream.
  ///
  /// In en, this message translates to:
  /// **'How to livestream'**
  String get howToLivestream;

  /// No description provided for @startYourLivestream.
  ///
  /// In en, this message translates to:
  /// **'Start your livestream'**
  String get startYourLivestream;

  /// No description provided for @liveStreaming.
  ///
  /// In en, this message translates to:
  /// **'Live Streaming'**
  String get liveStreaming;

  /// No description provided for @follow.
  ///
  /// In en, this message translates to:
  /// **'Follow'**
  String get follow;

  /// No description provided for @unfollow.
  ///
  /// In en, this message translates to:
  /// **'Unfollow'**
  String get unfollow;

  /// No description provided for @followersWithCount.
  ///
  /// In en, this message translates to:
  /// **'Followers: {count}'**
  String followersWithCount(int count);

  /// No description provided for @joined.
  ///
  /// In en, this message translates to:
  /// **'Joined'**
  String get joined;

  /// No description provided for @showMore.
  ///
  /// In en, this message translates to:
  /// **'Show more'**
  String get showMore;

  /// No description provided for @pleaseWaitWhileLoading.
  ///
  /// In en, this message translates to:
  /// **'Please wait while we load the content'**
  String get pleaseWaitWhileLoading;

  /// No description provided for @noLivestreams.
  ///
  /// In en, this message translates to:
  /// **'No livestreams'**
  String get noLivestreams;

  /// No description provided for @noLivestreamsDescription.
  ///
  /// In en, this message translates to:
  /// **'There is currently no one streaming, check back later or explore our video on demand content'**
  String get noLivestreamsDescription;

  /// No description provided for @noVideos.
  ///
  /// In en, this message translates to:
  /// **'No Videos Available'**
  String get noVideos;

  /// No description provided for @noVideosDescription.
  ///
  /// In en, this message translates to:
  /// **'There are currently no videos available. Check back later for new content'**
  String get noVideosDescription;

  /// No description provided for @searching.
  ///
  /// In en, this message translates to:
  /// **'Searching...'**
  String get searching;

  /// No description provided for @noUsersFound.
  ///
  /// In en, this message translates to:
  /// **'No users found'**
  String get noUsersFound;

  /// No description provided for @noDescription.
  ///
  /// In en, this message translates to:
  /// **'No description'**
  String get noDescription;

  /// No description provided for @add.
  ///
  /// In en, this message translates to:
  /// **'Add'**
  String get add;

  /// No description provided for @bio.
  ///
  /// In en, this message translates to:
  /// **'Bio'**
  String get bio;

  /// No description provided for @startedAt.
  ///
  /// In en, this message translates to:
  /// **'Started {time}'**
  String startedAt(String time);

  /// No description provided for @live.
  ///
  /// In en, this message translates to:
  /// **'Live'**
  String get live;

  /// No description provided for @following.
  ///
  /// In en, this message translates to:
  /// **'Following'**
  String get following;

  /// No description provided for @recommended.
  ///
  /// In en, this message translates to:
  /// **'Recommended'**
  String get recommended;

  /// No description provided for @timeSecondsAgo.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, one{{count} second ago} other{{count} seconds ago}}'**
  String timeSecondsAgo(int count);

  /// No description provided for @timeMinutesAgo.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, one{{count} minute ago} other{{count} minutes ago}}'**
  String timeMinutesAgo(int count);

  /// No description provided for @timeHoursAgo.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, one{{count} hour ago} other{{count} hours ago}}'**
  String timeHoursAgo(int count);

  /// No description provided for @timeDaysAgo.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, one{{count} day ago} other{{count} days ago}}'**
  String timeDaysAgo(int count);

  /// No description provided for @timeWeeksAgo.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, one{{count} week ago} other{{count} weeks ago}}'**
  String timeWeeksAgo(int count);

  /// No description provided for @timeMonthsAgo.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, one{{count} month ago} other{{count} months ago}}'**
  String timeMonthsAgo(int count);

  /// No description provided for @timeYearsAgo.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, one{{count} year ago} other{{count} years ago}}'**
  String timeYearsAgo(int count);

  /// No description provided for @authLogin.
  ///
  /// In en, this message translates to:
  /// **'Log in'**
  String get authLogin;

  /// No description provided for @authLoginTitle.
  ///
  /// In en, this message translates to:
  /// **'Welcome back!'**
  String get authLoginTitle;

  /// No description provided for @authLoginSubtitle.
  ///
  /// In en, this message translates to:
  /// **'Gain access to the world right now.'**
  String get authLoginSubtitle;

  /// No description provided for @authEmail.
  ///
  /// In en, this message translates to:
  /// **'Email'**
  String get authEmail;

  /// No description provided for @authPassword.
  ///
  /// In en, this message translates to:
  /// **'Password'**
  String get authPassword;

  /// No description provided for @authForgotPassword.
  ///
  /// In en, this message translates to:
  /// **'Forgot password?'**
  String get authForgotPassword;

  /// No description provided for @authSignup.
  ///
  /// In en, this message translates to:
  /// **'Sign up'**
  String get authSignup;

  /// No description provided for @authSignupTitle.
  ///
  /// In en, this message translates to:
  /// **'Welcome! Sign up for a new world?'**
  String get authSignupTitle;

  /// No description provided for @authSignupSubtitle.
  ///
  /// In en, this message translates to:
  /// **'Choose a method below to begin'**
  String get authSignupSubtitle;

  /// No description provided for @authLogout.
  ///
  /// In en, this message translates to:
  /// **'Log out'**
  String get authLogout;

  /// No description provided for @authNoAccount.
  ///
  /// In en, this message translates to:
  /// **'Don\'t have an account?'**
  String get authNoAccount;

  /// No description provided for @authHaveAccount.
  ///
  /// In en, this message translates to:
  /// **'Already have an account?'**
  String get authHaveAccount;

  /// No description provided for @authConfirmPassword.
  ///
  /// In en, this message translates to:
  /// **'Confirm password'**
  String get authConfirmPassword;

  /// No description provided for @authAccountCreatedSuccess.
  ///
  /// In en, this message translates to:
  /// **'Account created successfully'**
  String get authAccountCreatedSuccess;

  /// No description provided for @authEnterVerificationCode.
  ///
  /// In en, this message translates to:
  /// **'Enter Verification Code'**
  String get authEnterVerificationCode;

  /// No description provided for @authOtpDialogDescriptionPart1.
  ///
  /// In en, this message translates to:
  /// **'A 6-digit code has been sent to'**
  String get authOtpDialogDescriptionPart1;

  /// No description provided for @authOtpDialogDescriptionPart2.
  ///
  /// In en, this message translates to:
  /// **'. Please enter it below to verify your email address.'**
  String get authOtpDialogDescriptionPart2;

  /// No description provided for @authVerifyOtp.
  ///
  /// In en, this message translates to:
  /// **'Verify OTP'**
  String get authVerifyOtp;

  /// No description provided for @authResendOtp.
  ///
  /// In en, this message translates to:
  /// **'Resend OTP'**
  String get authResendOtp;

  /// No description provided for @authSendingOtp.
  ///
  /// In en, this message translates to:
  /// **'Sending OTP...'**
  String get authSendingOtp;

  /// No description provided for @authOtpResendCountDown.
  ///
  /// In en, this message translates to:
  /// **'Resend in {countdown}s'**
  String authOtpResendCountDown(int countdown);

  /// No description provided for @errorGeneralTitle.
  ///
  /// In en, this message translates to:
  /// **'Oops!'**
  String get errorGeneralTitle;

  /// No description provided for @errorGeneralDescription.
  ///
  /// In en, this message translates to:
  /// **'Something has gone wrong, please try again later'**
  String get errorGeneralDescription;

  /// No description provided for @errorEmailRequired.
  ///
  /// In en, this message translates to:
  /// **'Email is required'**
  String get errorEmailRequired;

  /// No description provided for @errorEmailInvalid.
  ///
  /// In en, this message translates to:
  /// **'Email is invalid'**
  String get errorEmailInvalid;

  /// No description provided for @errorPasswordRequired.
  ///
  /// In en, this message translates to:
  /// **'Password is required'**
  String get errorPasswordRequired;

  /// No description provided for @errorPasswordTooShort.
  ///
  /// In en, this message translates to:
  /// **'Password must be at least {minLength} characters'**
  String errorPasswordTooShort(int minLength);

  /// No description provided for @errorPasswordTooLong.
  ///
  /// In en, this message translates to:
  /// **'Password must be at most {maxLength} characters'**
  String errorPasswordTooLong(int maxLength);

  /// No description provided for @errorPasswordMissingLowercase.
  ///
  /// In en, this message translates to:
  /// **'Password must contain at least one lowercase letter'**
  String get errorPasswordMissingLowercase;

  /// No description provided for @errorPasswordMissingUppercase.
  ///
  /// In en, this message translates to:
  /// **'Password must contain at least one uppercase letter'**
  String get errorPasswordMissingUppercase;

  /// No description provided for @errorPasswordMissingSpecial.
  ///
  /// In en, this message translates to:
  /// **'Password must contain at least one special character'**
  String get errorPasswordMissingSpecial;

  /// No description provided for @errorUsernameRequired.
  ///
  /// In en, this message translates to:
  /// **'Username is required'**
  String get errorUsernameRequired;

  /// No description provided for @errorUsernameTooShort.
  ///
  /// In en, this message translates to:
  /// **'Username must be >= 6 characters'**
  String get errorUsernameTooShort;

  /// No description provided for @errorUsernameTooLong.
  ///
  /// In en, this message translates to:
  /// **'Username must be <= 20 characters'**
  String get errorUsernameTooLong;

  /// No description provided for @errorConfirmPasswordRequired.
  ///
  /// In en, this message translates to:
  /// **'Please confirm your password'**
  String get errorConfirmPasswordRequired;

  /// No description provided for @errorPasswordsDoNotMatch.
  ///
  /// In en, this message translates to:
  /// **'Passwords do not match'**
  String get errorPasswordsDoNotMatch;

  /// No description provided for @errorNewPasswordMustBeDifferent.
  ///
  /// In en, this message translates to:
  /// **'New password must be different from current password'**
  String get errorNewPasswordMustBeDifferent;

  /// No description provided for @errorOtpRequired.
  ///
  /// In en, this message translates to:
  /// **'Please enter OTP code'**
  String get errorOtpRequired;

  /// No description provided for @errorOtpSendFail.
  ///
  /// In en, this message translates to:
  /// **'Failed to send OTP'**
  String get errorOtpSendFail;

  /// No description provided for @apiResErrInvalidInput.
  ///
  /// In en, this message translates to:
  /// **'Input invalid.'**
  String get apiResErrInvalidInput;

  /// No description provided for @apiResErrInvalidPayload.
  ///
  /// In en, this message translates to:
  /// **'Payload invalid.'**
  String get apiResErrInvalidPayload;

  /// No description provided for @apiResErrAuthAlreadyExists.
  ///
  /// In en, this message translates to:
  /// **'Email has already been registered.'**
  String get apiResErrAuthAlreadyExists;

  /// No description provided for @apiResErrCaptchaFailed.
  ///
  /// In en, this message translates to:
  /// **'Failed to verify CAPTCHA, please try again.'**
  String get apiResErrCaptchaFailed;

  /// No description provided for @apiResErrPasswordNotMatch.
  ///
  /// In en, this message translates to:
  /// **'Old password does not match.'**
  String get apiResErrPasswordNotMatch;

  /// No description provided for @apiResErrUnauthorized.
  ///
  /// In en, this message translates to:
  /// **'Unauthorized.'**
  String get apiResErrUnauthorized;

  /// No description provided for @apiResErrSignUpOtpExpired.
  ///
  /// In en, this message translates to:
  /// **'OTP code has expired, please issue a new one.'**
  String get apiResErrSignUpOtpExpired;

  /// No description provided for @apiResErrEmailOrPasswordIncorrect.
  ///
  /// In en, this message translates to:
  /// **'Username or password incorrect.'**
  String get apiResErrEmailOrPasswordIncorrect;

  /// No description provided for @apiResErrForbidden.
  ///
  /// In en, this message translates to:
  /// **'Forbidden.'**
  String get apiResErrForbidden;

  /// No description provided for @apiResErrAuthNotFound.
  ///
  /// In en, this message translates to:
  /// **'Authentication credentials not found.'**
  String get apiResErrAuthNotFound;

  /// No description provided for @apiResErrRefreshTokenNotFound.
  ///
  /// In en, this message translates to:
  /// **'Refresh token not found.'**
  String get apiResErrRefreshTokenNotFound;

  /// No description provided for @apiResErrSignUpOtpNotFound.
  ///
  /// In en, this message translates to:
  /// **'OTP code not found.'**
  String get apiResErrSignUpOtpNotFound;

  /// No description provided for @apiResErrRouteNotFound.
  ///
  /// In en, this message translates to:
  /// **'Requested endpoint not found.'**
  String get apiResErrRouteNotFound;

  /// No description provided for @apiResErrSignUpOtpAlreadyUsed.
  ///
  /// In en, this message translates to:
  /// **'The OTP has already been used.'**
  String get apiResErrSignUpOtpAlreadyUsed;

  /// No description provided for @apiResErrFailedToCreateSignUpOtp.
  ///
  /// In en, this message translates to:
  /// **'Failed to generate the OTP, please try again later.'**
  String get apiResErrFailedToCreateSignUpOtp;

  /// No description provided for @apiResErrDatabaseQuery.
  ///
  /// In en, this message translates to:
  /// **'Error querying database, please try again.'**
  String get apiResErrDatabaseQuery;

  /// No description provided for @apiResErrDatabaseIssue.
  ///
  /// In en, this message translates to:
  /// **'Database issue, please try again.'**
  String get apiResErrDatabaseIssue;

  /// No description provided for @apiResErrInternalServer.
  ///
  /// In en, this message translates to:
  /// **'Something went wrong.'**
  String get apiResErrInternalServer;

  /// No description provided for @apiResErrFailedToSendVerification.
  ///
  /// In en, this message translates to:
  /// **'Failed to send email verification, please try again later.'**
  String get apiResErrFailedToSendVerification;

  /// No description provided for @apiResErrUserNotFound.
  ///
  /// In en, this message translates to:
  /// **'User not found.'**
  String get apiResErrUserNotFound;

  /// No description provided for @apiResErrImageTooLarge.
  ///
  /// In en, this message translates to:
  /// **'Image exceeds 10mb limit.'**
  String get apiResErrImageTooLarge;

  /// No description provided for @apiResErrLivestreamUpdateAfterEnded.
  ///
  /// In en, this message translates to:
  /// **'Failed to update, the livestream has ended.'**
  String get apiResErrLivestreamUpdateAfterEnded;

  /// No description provided for @apiResErrLivestreamNotFound.
  ///
  /// In en, this message translates to:
  /// **'Livestream not found.'**
  String get apiResErrLivestreamNotFound;

  /// No description provided for @apiResErrVodNotFound.
  ///
  /// In en, this message translates to:
  /// **'VOD not found.'**
  String get apiResErrVodNotFound;

  /// No description provided for @apiResErrEndAlreadyEndedLivestream.
  ///
  /// In en, this message translates to:
  /// **'The livestream has already been ended.'**
  String get apiResErrEndAlreadyEndedLivestream;

  /// No description provided for @apiResErrVodCommentNotFound.
  ///
  /// In en, this message translates to:
  /// **'Comment not found.'**
  String get apiResErrVodCommentNotFound;

  /// No description provided for @apiResErrVodCommentCreateFailed.
  ///
  /// In en, this message translates to:
  /// **'Failed to create comment.'**
  String get apiResErrVodCommentCreateFailed;

  /// No description provided for @apiResErrVodCommentAlreadyLiked.
  ///
  /// In en, this message translates to:
  /// **'Comment already liked.'**
  String get apiResErrVodCommentAlreadyLiked;

  /// No description provided for @apiResErrVodCommentNotLiked.
  ///
  /// In en, this message translates to:
  /// **'Comment has not been liked.'**
  String get apiResErrVodCommentNotLiked;

  /// No description provided for @apiResErrVodCommentDeleteFailed.
  ///
  /// In en, this message translates to:
  /// **'Failed to delete comment.'**
  String get apiResErrVodCommentDeleteFailed;

  /// No description provided for @apiResSuccSentVerificationEmail.
  ///
  /// In en, this message translates to:
  /// **'Verification email sent, please check your inbox'**
  String get apiResSuccSentVerificationEmail;

  /// No description provided for @apiResSuccOk.
  ///
  /// In en, this message translates to:
  /// **'Success'**
  String get apiResSuccOk;

  /// No description provided for @apiResSuccLogin.
  ///
  /// In en, this message translates to:
  /// **'Login successfully'**
  String get apiResSuccLogin;

  /// No description provided for @apiResSuccSignUp.
  ///
  /// In en, this message translates to:
  /// **'Sign up successfully'**
  String get apiResSuccSignUp;

  /// No description provided for @apiDefaultError.
  ///
  /// In en, this message translates to:
  /// **'Something went wrong. Please try again.'**
  String get apiDefaultError;

  /// No description provided for @fetchError.
  ///
  /// In en, this message translates to:
  /// **'Failed to fetch data, please try again.'**
  String get fetchError;

  /// No description provided for @themeLight.
  ///
  /// In en, this message translates to:
  /// **'Light'**
  String get themeLight;

  /// No description provided for @themeDark.
  ///
  /// In en, this message translates to:
  /// **'Dark'**
  String get themeDark;

  /// No description provided for @themeSystem.
  ///
  /// In en, this message translates to:
  /// **'System'**
  String get themeSystem;

  /// No description provided for @settingsTitle.
  ///
  /// In en, this message translates to:
  /// **'Settings'**
  String get settingsTitle;

  /// No description provided for @settingsNeedToLogin.
  ///
  /// In en, this message translates to:
  /// **'You need to log in to configure your settings'**
  String get settingsNeedToLogin;

  /// No description provided for @settingsNavProfile.
  ///
  /// In en, this message translates to:
  /// **'Profile'**
  String get settingsNavProfile;

  /// No description provided for @settingsNavSecurity.
  ///
  /// In en, this message translates to:
  /// **'Security'**
  String get settingsNavSecurity;

  /// No description provided for @settingsNavStream.
  ///
  /// In en, this message translates to:
  /// **'Stream'**
  String get settingsNavStream;

  /// No description provided for @settingsNavVods.
  ///
  /// In en, this message translates to:
  /// **'VODs'**
  String get settingsNavVods;

  /// No description provided for @settingsProfileTitle.
  ///
  /// In en, this message translates to:
  /// **'Profile Settings'**
  String get settingsProfileTitle;

  /// No description provided for @settingsProfileDescription.
  ///
  /// In en, this message translates to:
  /// **'Change identifying details for your account'**
  String get settingsProfileDescription;

  /// No description provided for @settingsProfileUsername.
  ///
  /// In en, this message translates to:
  /// **'Username'**
  String get settingsProfileUsername;

  /// No description provided for @settingsProfileDisplayName.
  ///
  /// In en, this message translates to:
  /// **'Display Name'**
  String get settingsProfileDisplayName;

  /// No description provided for @settingsProfileBio.
  ///
  /// In en, this message translates to:
  /// **'Bio'**
  String get settingsProfileBio;

  /// No description provided for @settingsProfileUpdateSuccess.
  ///
  /// In en, this message translates to:
  /// **'Profile information updated successfully!'**
  String get settingsProfileUpdateSuccess;

  /// No description provided for @settingsProfileUpdateBackground.
  ///
  /// In en, this message translates to:
  /// **'Update background'**
  String get settingsProfileUpdateBackground;

  /// No description provided for @settingsProfileUpdateProfilePicture.
  ///
  /// In en, this message translates to:
  /// **'Update profile picture'**
  String get settingsProfileUpdateProfilePicture;

  /// No description provided for @settingsSocialMediaTitle.
  ///
  /// In en, this message translates to:
  /// **'Edit Social Media Links'**
  String get settingsSocialMediaTitle;

  /// No description provided for @settingsSocialMediaDescription.
  ///
  /// In en, this message translates to:
  /// **'Manage links to your social media profiles'**
  String get settingsSocialMediaDescription;

  /// No description provided for @settingsSocialMediaInvalidUrl.
  ///
  /// In en, this message translates to:
  /// **'Invalid link, please try again'**
  String get settingsSocialMediaInvalidUrl;

  /// No description provided for @settingsSocialMediaErrEmptyUrl.
  ///
  /// In en, this message translates to:
  /// **'Link can\'t be empty'**
  String get settingsSocialMediaErrEmptyUrl;

  /// No description provided for @settingsThemesTitle.
  ///
  /// In en, this message translates to:
  /// **'Theme'**
  String get settingsThemesTitle;

  /// No description provided for @settingsThemesDescription.
  ///
  /// In en, this message translates to:
  /// **'Customize the look and feel of the app'**
  String get settingsThemesDescription;

  /// No description provided for @settingsLanguageTitle.
  ///
  /// In en, this message translates to:
  /// **'Language'**
  String get settingsLanguageTitle;

  /// No description provided for @settingsLanguageDescription.
  ///
  /// In en, this message translates to:
  /// **'Select your preferred language'**
  String get settingsLanguageDescription;

  /// No description provided for @settingsSaveChanges.
  ///
  /// In en, this message translates to:
  /// **'Save Changes'**
  String get settingsSaveChanges;

  /// No description provided for @settingsFileSizeExceeds.
  ///
  /// In en, this message translates to:
  /// **'File size exceeds {size} MB'**
  String settingsFileSizeExceeds(int size);

  /// No description provided for @settingsSecurityApiKeyLabel.
  ///
  /// In en, this message translates to:
  /// **'Your API Key'**
  String get settingsSecurityApiKeyLabel;

  /// No description provided for @settingsSecurityApiKeyGenerate.
  ///
  /// In en, this message translates to:
  /// **'Generate New API Key'**
  String get settingsSecurityApiKeyGenerate;

  /// No description provided for @settingsSecurityApiKeyCopied.
  ///
  /// In en, this message translates to:
  /// **'API Key copied to clipboard'**
  String get settingsSecurityApiKeyCopied;

  /// No description provided for @settingsSecurityContactTitle.
  ///
  /// In en, this message translates to:
  /// **'Contact'**
  String get settingsSecurityContactTitle;

  /// No description provided for @settingsSecurityContactDescription.
  ///
  /// In en, this message translates to:
  /// **'Where we send important messages about your account'**
  String get settingsSecurityContactDescription;

  /// No description provided for @settingsSecurityContactEmail.
  ///
  /// In en, this message translates to:
  /// **'Email'**
  String get settingsSecurityContactEmail;

  /// No description provided for @settingsSecurityContactPhone.
  ///
  /// In en, this message translates to:
  /// **'Phone Number'**
  String get settingsSecurityContactPhone;

  /// No description provided for @settingsSecurityContactAddPhone.
  ///
  /// In en, this message translates to:
  /// **'Add a number'**
  String get settingsSecurityContactAddPhone;

  /// No description provided for @settingsSecurityContactPhoneInvalid.
  ///
  /// In en, this message translates to:
  /// **'Invalid phone number, please try again'**
  String get settingsSecurityContactPhoneInvalid;

  /// No description provided for @settingsSecurityTitle.
  ///
  /// In en, this message translates to:
  /// **'Security'**
  String get settingsSecurityTitle;

  /// No description provided for @settingsSecurityDescription.
  ///
  /// In en, this message translates to:
  /// **'Keep your account safe and sound'**
  String get settingsSecurityDescription;

  /// No description provided for @settingsSecurityPasswordTitle.
  ///
  /// In en, this message translates to:
  /// **'Password'**
  String get settingsSecurityPasswordTitle;

  /// No description provided for @settingsSecurityPasswordChange.
  ///
  /// In en, this message translates to:
  /// **'Change password'**
  String get settingsSecurityPasswordChange;

  /// No description provided for @settingsSecurityPasswordDialogTitle.
  ///
  /// In en, this message translates to:
  /// **'Change Password'**
  String get settingsSecurityPasswordDialogTitle;

  /// No description provided for @settingsSecurityPasswordLocalDescription.
  ///
  /// In en, this message translates to:
  /// **'Improve your security with a strong password.'**
  String get settingsSecurityPasswordLocalDescription;

  /// No description provided for @settingsSecurityPasswordUpdatedSuccess.
  ///
  /// In en, this message translates to:
  /// **'Password updated successfully'**
  String get settingsSecurityPasswordUpdatedSuccess;

  /// No description provided for @settingsSecurityPasswordFormCurrentLabel.
  ///
  /// In en, this message translates to:
  /// **'Current Password'**
  String get settingsSecurityPasswordFormCurrentLabel;

  /// No description provided for @settingsSecurityPasswordFormCurrentPlaceholder.
  ///
  /// In en, this message translates to:
  /// **'Enter your current password'**
  String get settingsSecurityPasswordFormCurrentPlaceholder;

  /// No description provided for @settingsSecurityPasswordFormNewLabel.
  ///
  /// In en, this message translates to:
  /// **'New Password'**
  String get settingsSecurityPasswordFormNewLabel;

  /// No description provided for @settingsSecurityPasswordFormNewPlaceholder.
  ///
  /// In en, this message translates to:
  /// **'Enter your new password'**
  String get settingsSecurityPasswordFormNewPlaceholder;

  /// No description provided for @settingsSecurityPasswordFormConfirmLabel.
  ///
  /// In en, this message translates to:
  /// **'Confirm New Password'**
  String get settingsSecurityPasswordFormConfirmLabel;

  /// No description provided for @settingsSecurityPasswordFormConfirmPlaceholder.
  ///
  /// In en, this message translates to:
  /// **'Confirm your new password'**
  String get settingsSecurityPasswordFormConfirmPlaceholder;

  /// No description provided for @settingsSecurityPasswordFormSubmit.
  ///
  /// In en, this message translates to:
  /// **'Confirm'**
  String get settingsSecurityPasswordFormSubmit;

  /// No description provided for @messagesTitle.
  ///
  /// In en, this message translates to:
  /// **'Messages'**
  String get messagesTitle;

  /// No description provided for @messagesSelectConversation.
  ///
  /// In en, this message translates to:
  /// **'Select a conversation'**
  String get messagesSelectConversation;

  /// No description provided for @messagesLoginRequired.
  ///
  /// In en, this message translates to:
  /// **'Please log in to view messages.'**
  String get messagesLoginRequired;

  /// No description provided for @messagesUnknown.
  ///
  /// In en, this message translates to:
  /// **'Unknown'**
  String get messagesUnknown;

  /// No description provided for @messagesGroup.
  ///
  /// In en, this message translates to:
  /// **'Group'**
  String get messagesGroup;

  /// No description provided for @messagesOnline.
  ///
  /// In en, this message translates to:
  /// **'Online'**
  String get messagesOnline;

  /// No description provided for @messagesOffline.
  ///
  /// In en, this message translates to:
  /// **'Offline'**
  String get messagesOffline;

  /// No description provided for @messagesMembersCount.
  ///
  /// In en, this message translates to:
  /// **'{count} members'**
  String messagesMembersCount(int count);

  /// No description provided for @messagesNoConversationsYet.
  ///
  /// In en, this message translates to:
  /// **'No conversations yet'**
  String get messagesNoConversationsYet;

  /// No description provided for @messagesLoadMore.
  ///
  /// In en, this message translates to:
  /// **'Load more'**
  String get messagesLoadMore;

  /// No description provided for @messagesNoMessagesYet.
  ///
  /// In en, this message translates to:
  /// **'No messages yet'**
  String get messagesNoMessagesYet;

  /// No description provided for @messagesNewConversation.
  ///
  /// In en, this message translates to:
  /// **'New Conversation'**
  String get messagesNewConversation;

  /// No description provided for @messagesDirectMessage.
  ///
  /// In en, this message translates to:
  /// **'Direct Message'**
  String get messagesDirectMessage;

  /// No description provided for @messagesGroupNamePlaceholder.
  ///
  /// In en, this message translates to:
  /// **'Group name (optional)'**
  String get messagesGroupNamePlaceholder;

  /// No description provided for @messagesSearchUsersPlaceholder.
  ///
  /// In en, this message translates to:
  /// **'Search users...'**
  String get messagesSearchUsersPlaceholder;

  /// No description provided for @messagesSearching.
  ///
  /// In en, this message translates to:
  /// **'Searching...'**
  String get messagesSearching;

  /// No description provided for @messagesCreating.
  ///
  /// In en, this message translates to:
  /// **'Creating...'**
  String get messagesCreating;

  /// No description provided for @messagesCreate.
  ///
  /// In en, this message translates to:
  /// **'Create'**
  String get messagesCreate;

  /// No description provided for @messagesSentAnImage.
  ///
  /// In en, this message translates to:
  /// **'Sent an image'**
  String get messagesSentAnImage;

  /// No description provided for @messagesSentImagesCount.
  ///
  /// In en, this message translates to:
  /// **'Sent {count} images'**
  String messagesSentImagesCount(int count);

  /// No description provided for @messagesUploadFailed.
  ///
  /// In en, this message translates to:
  /// **'Failed to upload files'**
  String get messagesUploadFailed;

  /// No description provided for @messagesPlaceholderTypeMessage.
  ///
  /// In en, this message translates to:
  /// **'Type a message...'**
  String get messagesPlaceholderTypeMessage;

  /// No description provided for @messagesMessageDeleted.
  ///
  /// In en, this message translates to:
  /// **'This message was deleted'**
  String get messagesMessageDeleted;

  /// No description provided for @messagesBeginningOfConversation.
  ///
  /// In en, this message translates to:
  /// **'Beginning of conversation'**
  String get messagesBeginningOfConversation;

  /// No description provided for @messagesTypingOne.
  ///
  /// In en, this message translates to:
  /// **'{name} is typing...'**
  String messagesTypingOne(String name);

  /// No description provided for @notificationsTitle.
  ///
  /// In en, this message translates to:
  /// **'Notifications'**
  String get notificationsTitle;

  /// No description provided for @notificationsMarkAllAsRead.
  ///
  /// In en, this message translates to:
  /// **'Mark all as read'**
  String get notificationsMarkAllAsRead;

  /// No description provided for @notificationsLoading.
  ///
  /// In en, this message translates to:
  /// **'Loading...'**
  String get notificationsLoading;

  /// No description provided for @notificationsNoNotifications.
  ///
  /// In en, this message translates to:
  /// **'No notifications'**
  String get notificationsNoNotifications;

  /// No description provided for @notificationsViewAll.
  ///
  /// In en, this message translates to:
  /// **'View all notifications'**
  String get notificationsViewAll;

  /// No description provided for @notificationsNoNotificationsYet.
  ///
  /// In en, this message translates to:
  /// **'No notifications yet.'**
  String get notificationsNoNotificationsYet;

  /// No description provided for @notificationsPleaseLogIn.
  ///
  /// In en, this message translates to:
  /// **'Please log in to view notifications.'**
  String get notificationsPleaseLogIn;

  /// No description provided for @notificationsView.
  ///
  /// In en, this message translates to:
  /// **'View'**
  String get notificationsView;

  /// No description provided for @notificationsMarkAsRead.
  ///
  /// In en, this message translates to:
  /// **'Mark as read'**
  String get notificationsMarkAsRead;

  /// No description provided for @notificationsDelete.
  ///
  /// In en, this message translates to:
  /// **'Delete'**
  String get notificationsDelete;

  /// No description provided for @notificationsLoadMore.
  ///
  /// In en, this message translates to:
  /// **'Load more'**
  String get notificationsLoadMore;

  /// No description provided for @commentsTitle.
  ///
  /// In en, this message translates to:
  /// **'Comments'**
  String get commentsTitle;

  /// No description provided for @commentsWriteComment.
  ///
  /// In en, this message translates to:
  /// **'Write a comment...'**
  String get commentsWriteComment;

  /// No description provided for @commentsWriteReply.
  ///
  /// In en, this message translates to:
  /// **'Write a reply...'**
  String get commentsWriteReply;

  /// No description provided for @commentsPost.
  ///
  /// In en, this message translates to:
  /// **'Post'**
  String get commentsPost;

  /// No description provided for @commentsReply.
  ///
  /// In en, this message translates to:
  /// **'Reply'**
  String get commentsReply;

  /// No description provided for @commentsDelete.
  ///
  /// In en, this message translates to:
  /// **'Delete'**
  String get commentsDelete;

  /// No description provided for @commentsDeletedComment.
  ///
  /// In en, this message translates to:
  /// **'This comment has been deleted.'**
  String get commentsDeletedComment;

  /// No description provided for @commentsViewReplies.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, one{View {count} reply} other{View {count} replies}}'**
  String commentsViewReplies(int count);

  /// No description provided for @commentsLoadMoreReplies.
  ///
  /// In en, this message translates to:
  /// **'Load more replies'**
  String get commentsLoadMoreReplies;

  /// No description provided for @commentsNoComments.
  ///
  /// In en, this message translates to:
  /// **'No comments yet. Be the first to comment!'**
  String get commentsNoComments;

  /// No description provided for @commentsLoginToComment.
  ///
  /// In en, this message translates to:
  /// **'Log in to leave a comment.'**
  String get commentsLoginToComment;

  /// No description provided for @commentsOwner.
  ///
  /// In en, this message translates to:
  /// **'Owner'**
  String get commentsOwner;

  /// No description provided for @commentsYou.
  ///
  /// In en, this message translates to:
  /// **'You'**
  String get commentsYou;

  /// No description provided for @commentsDeleteConfirmTitle.
  ///
  /// In en, this message translates to:
  /// **'Delete comment?'**
  String get commentsDeleteConfirmTitle;

  /// No description provided for @commentsDeleteConfirmDescription.
  ///
  /// In en, this message translates to:
  /// **'This action cannot be undone.'**
  String get commentsDeleteConfirmDescription;

  /// No description provided for @commentsLike.
  ///
  /// In en, this message translates to:
  /// **'Like'**
  String get commentsLike;

  /// No description provided for @commentsUnlike.
  ///
  /// In en, this message translates to:
  /// **'Unlike'**
  String get commentsUnlike;

  /// No description provided for @commentsCharRemaining.
  ///
  /// In en, this message translates to:
  /// **'{count} characters remaining'**
  String commentsCharRemaining(int count);

  /// No description provided for @usersChatTitle.
  ///
  /// In en, this message translates to:
  /// **'Chat'**
  String get usersChatTitle;

  /// No description provided for @usersChatJoined.
  ///
  /// In en, this message translates to:
  /// **'joined the chat'**
  String get usersChatJoined;

  /// No description provided for @usersChatLeft.
  ///
  /// In en, this message translates to:
  /// **'left the chat'**
  String get usersChatLeft;

  /// No description provided for @usersChatPlaceholderLogin.
  ///
  /// In en, this message translates to:
  /// **'Login to start messaging'**
  String get usersChatPlaceholderLogin;

  /// No description provided for @usersChatPlaceholderTyping.
  ///
  /// In en, this message translates to:
  /// **'Type a message...'**
  String get usersChatPlaceholderTyping;

  /// No description provided for @usersOffline.
  ///
  /// In en, this message translates to:
  /// **'Offline'**
  String get usersOffline;

  /// No description provided for @usersProfileAbout.
  ///
  /// In en, this message translates to:
  /// **'About'**
  String get usersProfileAbout;

  /// No description provided for @usersProfileFollowers.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, one{follower} other{followers}}'**
  String usersProfileFollowers(int count);

  /// No description provided for @usersProfileJoinedPrefix.
  ///
  /// In en, this message translates to:
  /// **'Joined'**
  String get usersProfileJoinedPrefix;

  /// No description provided for @usersProfileRecentStreams.
  ///
  /// In en, this message translates to:
  /// **'Recent Streams'**
  String get usersProfileRecentStreams;

  /// No description provided for @accessibilityOpenMenu.
  ///
  /// In en, this message translates to:
  /// **'Open menu'**
  String get accessibilityOpenMenu;
}

class _AppLocalizationsDelegate
    extends LocalizationsDelegate<AppLocalizations> {
  const _AppLocalizationsDelegate();

  @override
  Future<AppLocalizations> load(Locale locale) {
    return SynchronousFuture<AppLocalizations>(lookupAppLocalizations(locale));
  }

  @override
  bool isSupported(Locale locale) =>
      <String>['en', 'vi'].contains(locale.languageCode);

  @override
  bool shouldReload(_AppLocalizationsDelegate old) => false;
}

AppLocalizations lookupAppLocalizations(Locale locale) {
  // Lookup logic when only language code is specified.
  switch (locale.languageCode) {
    case 'en':
      return AppLocalizationsEn();
    case 'vi':
      return AppLocalizationsVi();
  }

  throw FlutterError(
    'AppLocalizations.delegate failed to load unsupported locale "$locale". This is likely '
    'an issue with the localizations generation tool. Please file an issue '
    'on GitHub with a reproducible sample app and the gen-l10n configuration '
    'that was used.',
  );
}
