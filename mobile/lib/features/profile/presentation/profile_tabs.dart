import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/features/profile/presentation/profile_about_tab.dart';
import 'package:letslive_mobile/features/profile/presentation/profile_header.dart';
import 'package:letslive_mobile/features/profile/presentation/profile_home_tab.dart';
import 'package:letslive_mobile/features/profile/presentation/profile_videos_tab.dart';
import 'package:letslive_mobile/features/user/data/public_user.dart';
import 'package:letslive_mobile/features/user/user_providers.dart';

/// Height of the tab bar (Material's fixed tab height).
const double profileTabBarHeight = 48;

/// A profile (web `ProfileView`): [ProfileHeader], then tabs Home / Videos /
/// About, with [chat] as an extra first tab when given (a channel's live
/// chat; the own profile has none). Fills the height it gets; each tab
/// scrolls on its own.
///
/// [user] is what the parent loaded; the live copy from
/// [publicUserProvider] replaces it, so following changes the count here.
/// With [compact] (the keyboard is open) only [chat] shows, with no header
/// or tab bar, to leave room for typing. The chat keeps its place in the
/// widget tree either way, so its draft and focus survive the switch.
class ProfileTabs extends ConsumerStatefulWidget {
  const ProfileTabs({
    required this.user,
    this.chat,
    this.showBanner = true,
    this.compact = false,
    super.key,
  });

  /// The profile owner as first loaded.
  final PublicUser user;

  /// Content of the extra "Chat" tab, the first and default one.
  final Widget? chat;

  /// See [ProfileHeader.showBanner].
  final bool showBanner;

  /// See the class doc.
  final bool compact;

  @override
  ConsumerState<ProfileTabs> createState() => _ProfileTabsState();
}

class _ProfileTabsState extends ConsumerState<ProfileTabs>
    with SingleTickerProviderStateMixin {
  late final TabController _controller;

  bool get _hasChat => widget.chat != null;

  /// Index of the Videos tab.
  int get _videosIndex => _hasChat ? 2 : 1;

  @override
  void initState() {
    super.initState();
    _controller = TabController(length: _hasChat ? 4 : 3, vsync: this)
      ..addListener(_onTabChanged);
  }

  void _onTabChanged() {
    // The chat stays mounted while hidden; its message box must not keep
    // the focus (and the keyboard) on another tab.
    FocusManager.instance.primaryFocus?.unfocus();
    setState(() {});
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final user =
        ref.watch(publicUserProvider(widget.user.id)).value ?? widget.user;
    final chat = widget.chat;
    final compact = widget.compact && chat != null;
    final tab = _controller.index - (_hasChat ? 1 : 0);
    // Only the selected tab is built, except the chat: it stays mounted
    // (hidden) so its room subscription, scroll position and typed draft
    // survive a visit to another tab. In compact mode the chat shows in the
    // same place in the tree, so opening the keyboard never rebuilds it.
    final other = switch (tab) {
      _ when compact => null,
      < 0 => null,
      0 => ProfileHomeTab(
        user: user,
        onViewAll: () => _controller.animateTo(_videosIndex),
      ),
      1 => ProfileVideosTab(user: user),
      _ => ProfileAboutTab(user: user),
    };
    return Column(
      children: [
        if (!compact) ...[
          ProfileHeader(user: user, showBanner: widget.showBanner),
          TabBar(
            controller: _controller,
            isScrollable: true,
            tabAlignment: TabAlignment.start,
            tabs: [
              if (_hasChat) Tab(text: l10n.chat_title),
              Tab(text: l10n.users_channel_tabs_home),
              Tab(text: l10n.users_channel_tabs_videos),
              Tab(text: l10n.users_channel_tabs_about),
            ],
          ),
        ],
        Expanded(
          // Keyed so it keeps its element when the header and tab bar
          // above it come and go.
          key: const ValueKey('profile-tab-content'),
          child: Stack(
            fit: StackFit.expand,
            children: [
              if (chat != null) Offstage(offstage: other != null, child: chat),
              ?other,
            ],
          ),
        ),
      ],
    );
  }
}
