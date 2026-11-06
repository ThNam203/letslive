/// The Profile tab: the signed-in user's own profile as others see it
/// (header, then Home / Videos / About), plus a People tab with the
/// channels the user follows and recommended ones, and wallet, shop and settings
/// buttons in the app bar. The public copy of the user is loaded by id (the
/// session user lacks a few public fields); saving the profile invalidates
/// it.
library;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/router/app_chrome.dart';
import 'package:letslive_mobile/core/router/app_routes.dart';
import 'package:letslive_mobile/core/router/header_actions.dart';
import 'package:letslive_mobile/core/widgets/feed_section.dart';
import 'package:letslive_mobile/features/auth/auth_providers.dart';
import 'package:letslive_mobile/features/profile/presentation/profile_tabs.dart';
import 'package:letslive_mobile/features/user/user_providers.dart';

/// The own-profile tab screen.
class OwnProfileScreen extends ConsumerWidget {
  const OwnProfileScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final myId = ref.watch(sessionProvider).value?.user.id;
    final user = myId == null ? null : ref.watch(publicUserProvider(myId));
    final Widget body;
    if (myId == null || user == null) {
      body = const SizedBox.shrink();
    } else if (user.hasError && !user.isLoading) {
      body = Center(
        child: FeedError(
          error: user.error!,
          onRetry: () => ref.invalidate(publicUserProvider(myId)),
        ),
      );
    } else if (user.value == null) {
      body = const Center(child: CircularProgressIndicator());
    } else {
      body = ProfileTabs(user: user.requireValue, showPeople: true);
    }
    return TabScaffold(
      appBar: AppBar(
        title: Text(l10n.settings_navigation_profile),
        actions: [
          IconButton(
            key: const ValueKey('open-wallet'),
            tooltip: l10n.wallet_page_title,
            icon: const Icon(Icons.account_balance_wallet_outlined),
            onPressed: () => context.push(AppRoutes.wallet),
          ),
          IconButton(
            key: const ValueKey('open-shop'),
            tooltip: l10n.shop_shop_page_title,
            icon: const Icon(Icons.storefront_outlined),
            onPressed: () => context.push(AppRoutes.shop),
          ),
          IconButton(
            key: const ValueKey('open-settings'),
            tooltip: l10n.settings_page_title,
            icon: const Icon(Icons.settings_outlined),
            onPressed: () => context.push(AppRoutes.settings),
          ),
          const HeaderActions(),
        ],
      ),
      body: body,
    );
  }
}
