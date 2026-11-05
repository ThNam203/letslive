/// Edit the own profile (web `/settings/profile`): banner and profile
/// picture, username, bio and social links.
///
/// Like web, nothing is sent until Save. Save uploads the chosen banner,
/// then the chosen picture (with a progress bar), then sends the changed text
/// fields; each step reports its own failure and the later steps still run.
/// Images are checked first with web's size, type and dimension rules.
/// Social links can be added or replaced, not removed (web cannot either).
/// Theme, language and account disabling are not part of this screen yet.
library;

import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/l10n/api_message.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/core/validation/field_limits.dart';
import 'package:letslive_mobile/features/auth/auth_providers.dart';
import 'package:letslive_mobile/features/auth/domain/auth_validators.dart';
import 'package:letslive_mobile/features/auth/presentation/auth_widgets.dart';
import 'package:letslive_mobile/features/settings/domain/image_checks.dart';
import 'package:letslive_mobile/features/settings/settings_providers.dart';
import 'package:letslive_mobile/features/user/data/me_user.dart';
import 'package:letslive_mobile/features/user/data/picked_image.dart';
import 'package:letslive_mobile/features/user/data/social_links.dart';
import 'package:letslive_mobile/features/user/presentation/user_avatar.dart';
import 'package:letslive_mobile/features/user/user_providers.dart';

/// Names shown for the social platforms (brand names, not translated; web
/// hardcodes them too).
const Map<String, String> _platformLabels = {
  'facebook': 'Facebook',
  'twitter': 'X/Twitter',
  'instagram': 'Instagram',
  'linkedin': 'LinkedIn',
  'github': 'GitHub',
  'youtube': 'YouTube',
  'website': 'Website',
  'tiktok': 'TikTok',
};

/// The edit-profile screen.
class ProfileSettingsScreen extends ConsumerStatefulWidget {
  const ProfileSettingsScreen({super.key});

  @override
  ConsumerState<ProfileSettingsScreen> createState() =>
      _ProfileSettingsScreenState();
}

class _ProfileSettingsScreenState extends ConsumerState<ProfileSettingsScreen> {
  late final TextEditingController _username;
  late final TextEditingController _bio;
  late final Map<String, TextEditingController> _links;
  PickedImage? _avatar;
  PickedImage? _background;
  String? _usernameError;
  final Map<String, String> _linkErrors = {};

  /// Upload progress 0..1 of the image being sent; null when none is.
  double? _progress;
  bool _busy = false;

  MeUser get _me => ref.read(sessionProvider).requireValue!.user;

  @override
  void initState() {
    super.initState();
    final me = _me;
    _username = TextEditingController(text: me.username)..addListener(_rebuild);
    _bio = TextEditingController(text: me.bio ?? '')..addListener(_rebuild);
    _links = {
      for (final platform in socialPlatforms)
        platform: TextEditingController(
          text: me.socialMediaLinks[platform] ?? '',
        )..addListener(_rebuild),
    };
  }

  void _rebuild() => setState(() {});

  @override
  void dispose() {
    _username.dispose();
    _bio.dispose();
    for (final controller in _links.values) {
      controller.dispose();
    }
    super.dispose();
  }

  /// The social links that differ from what is saved and are not empty.
  Map<String, String> get _changedLinks {
    final saved = _me.socialMediaLinks;
    return {
      for (final MapEntry(:key, :value) in _links.entries)
        if (value.text.trim().isNotEmpty &&
            value.text.trim() != (saved[key] ?? ''))
          key: value.text.trim(),
    };
  }

  bool get _hasChanges =>
      _username.text != _me.username ||
      _bio.text != (_me.bio ?? '') ||
      _changedLinks.isNotEmpty ||
      _avatar != null ||
      _background != null;

  void _say(String message) {
    if (!mounted) return;
    ScaffoldMessenger.of(context)
        .showSnackBar(SnackBar(content: Text(message)));
  }

  /// Lets the user choose an image and keeps it as pending when it passes
  /// web's checks; otherwise says why not.
  Future<void> _pick({required bool avatar}) async {
    final l10n = AppLocalizations.of(context);
    final image = await ref.read(imagePickProvider)();
    if (image == null || !mounted) return;
    final maxMb = avatar
        ? FieldLimits.avatarMaxFileMb
        : FieldLimits.backgroundMaxFileMb;
    if (checkFileSize(image, maxMb) != null) {
      return _say(l10n.settings_file_size_exceeds(maxMb));
    }
    if (avatar) {
      if (checkAvatarType(image) != null) {
        return _say(l10n.api_response_res_err_invalid_image);
      }
      final size = await ref.read(imageSizeReaderProvider)(image.bytes);
      if (!mounted) return;
      if (size == null) return _say(l10n.settings_profile_image_load_failed);
      if (checkAvatarDimensions(size.width, size.height) != null) {
        return _say(
          // Generated parameters are alphabetical: max, then min.
          l10n.settings_profile_photo_dimensions_out_of_range(
            FieldLimits.avatarSourceMaxDimension,
            FieldLimits.avatarMinDimension,
          ),
        );
      }
    }
    setState(() => avatar ? _avatar = image : _background = image);
  }

  /// Validates the text fields; true when all are fine.
  bool _validate(AppLocalizations l10n) {
    final usernameError = _username.text == _me.username
        ? null
        : validateUsername(_username.text);
    final linkErrors = <String, String>{
      for (final MapEntry(:key, :value) in _changedLinks.entries)
        if (!_isUrl(value)) key: l10n.settings_social_media_links_invalid_url,
    };
    setState(() {
      _usernameError = authFieldErrorText(l10n, usernameError);
      _linkErrors
        ..clear()
        ..addAll(linkErrors);
    });
    return usernameError == null && linkErrors.isEmpty;
  }

  static bool _isUrl(String value) {
    final uri = Uri.tryParse(value);
    return uri != null && uri.hasScheme && uri.hasAuthority;
  }

  Future<void> _save() async {
    final l10n = AppLocalizations.of(context);
    if (!_validate(l10n)) return;
    setState(() => _busy = true);
    final session = ref.read(sessionProvider.notifier);
    final failures = <ApiResponse<Object?>>[];

    Future<void> upload(
      PickedImage? image,
      Future<ApiResponse<String>> Function(
        PickedImage image,
        void Function(int, int) onProgress,
      )
      send,
      VoidCallback onDone,
    ) async {
      if (image == null) return;
      setState(() => _progress = 0);
      final res = await send(image, (sent, total) {
        if (mounted && total > 0) setState(() => _progress = sent / total);
      });
      if (!mounted) return;
      setState(() => _progress = null);
      if (res.success) {
        onDone();
      } else {
        failures.add(res);
      }
    }

    await upload(
      _background,
      (image, onProgress) =>
          session.uploadBackgroundPicture(image, onProgress: onProgress),
      () => setState(() => _background = null),
    );
    if (!mounted) return;
    await upload(
      _avatar,
      (image, onProgress) =>
          session.uploadProfilePicture(image, onProgress: onProgress),
      () => setState(() => _avatar = null),
    );
    if (!mounted) return;

    final usernameChanged = _username.text != _me.username;
    final bioChanged = _bio.text != (_me.bio ?? '');
    final links = _changedLinks;
    if (usernameChanged || bioChanged || links.isNotEmpty) {
      final res = await session.updateProfile(
        username: usernameChanged ? _username.text : null,
        bio: bioChanged ? _bio.text : null,
        socialMediaLinks: links.isEmpty ? null : links,
      );
      if (!res.success) failures.add(res);
    }
    if (!mounted) return;
    ref.invalidate(publicUserProvider(_me.id));
    setState(() => _busy = false);
    _say(
      failures.isEmpty
          ? l10n.settings_profile_update_success
          : apiErrorMessage(l10n, failures.first),
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final theme = Theme.of(context);
    final me = ref.watch(sessionProvider).value?.user;
    if (me == null) return const Scaffold();
    final progress = _progress;
    return Scaffold(
      appBar: AppBar(title: Text(l10n.settings_profile_title)),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          _Banner(pending: _background, url: me.backgroundPicture),
          Wrap(
            spacing: 8,
            children: [
              OutlinedButton.icon(
                key: const ValueKey('pick-background'),
                onPressed: _busy ? null : () => _pick(avatar: false),
                icon: const Icon(Icons.image_outlined),
                label: Text(l10n.settings_profile_update_background),
              ),
            ],
          ),
          const SizedBox(height: 16),
          Row(
            children: [
              _avatarPreview(me),
              const SizedBox(width: 12),
              Expanded(
                child: OutlinedButton.icon(
                  key: const ValueKey('pick-avatar'),
                  onPressed: _busy ? null : () => _pick(avatar: true),
                  icon: const Icon(Icons.photo_camera_outlined),
                  label: Text(l10n.settings_profile_update_profile_picture),
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),
          TextField(
            key: const ValueKey('username'),
            controller: _username,
            enabled: !_busy,
            maxLength: FieldLimits.usernameMax,
            onChanged: (_) => setState(() => _usernameError = null),
            decoration: InputDecoration(
              labelText: l10n.settings_profile_username,
              errorText: _usernameError,
              counterText: '',
            ),
          ),
          const SizedBox(height: 12),
          TextField(
            key: const ValueKey('bio'),
            controller: _bio,
            enabled: !_busy,
            maxLength: FieldLimits.bioMax,
            minLines: 3,
            maxLines: 6,
            decoration: InputDecoration(labelText: l10n.settings_profile_bio),
          ),
          const SizedBox(height: 12),
          Text(
            l10n.settings_social_media_links_title,
            style: theme.textTheme.titleMedium,
          ),
          Text(l10n.settings_social_media_links_description),
          for (final platform in socialPlatforms)
            Padding(
              padding: const EdgeInsets.only(top: 12),
              child: TextField(
                key: ValueKey('social-$platform'),
                controller: _links[platform],
                enabled: !_busy,
                maxLength: FieldLimits.socialUrlMax,
                keyboardType: TextInputType.url,
                onChanged: (_) => setState(() => _linkErrors.remove(platform)),
                decoration: InputDecoration(
                  labelText: _platformLabels[platform],
                  errorText: _linkErrors[platform],
                  counterText: '',
                ),
              ),
            ),
          const SizedBox(height: 16),
          if (progress != null) ...[
            Text(l10n.settings_upload_uploading),
            LinearProgressIndicator(
              key: const ValueKey('upload-progress'),
              value: progress,
            ),
            const SizedBox(height: 16),
          ],
          SubmitButton(
            key: const ValueKey('profile-save'),
            label: l10n.common_save_changes,
            busy: _busy,
            onPressed: _hasChanges ? _save : () {},
          ),
        ],
      ),
    );
  }

  Widget _avatarPreview(MeUser me) {
    final pending = _avatar;
    if (pending == null) {
      return UserAvatar(
        pictureUrl: me.profilePicture,
        name: me.username,
        radius: 32,
      );
    }
    return CircleAvatar(
      radius: 32,
      backgroundImage: MemoryImage(pending.bytes),
    );
  }
}

/// The banner preview: the chosen image, else the saved one, else a flat
/// placeholder.
class _Banner extends StatelessWidget {
  const _Banner({required this.pending, required this.url});

  final PickedImage? pending;
  final String? url;

  @override
  Widget build(BuildContext context) {
    final image = pending;
    final saved = url;
    return AspectRatio(
      aspectRatio: 3,
      child: ClipRRect(
        borderRadius: BorderRadius.circular(8),
        child: ColoredBox(
          color: Theme.of(context).colorScheme.surfaceContainerHighest,
          child: image != null
              ? Image.memory(image.bytes, fit: BoxFit.cover)
              : saved == null
              ? null
              : CachedNetworkImage(
                  imageUrl: saved,
                  fit: BoxFit.cover,
                  errorWidget: (_, _, _) => const SizedBox.shrink(),
                ),
        ),
      ),
    );
  }
}
