/// The Search tab (web `/search`): a search box above the tabs All / Channels /
/// Live / Videos.
///
/// - The box waits [searchDebounce] after the last keystroke, then searches
///   the trimmed text; an empty text shows a prompt and sends nothing, and
///   Enter searches at once. Input is capped at [FieldLimits.searchQueryMax].
/// - Each query reads its own providers (see `search_providers.dart`), so a
///   slow answer to an older query never replaces newer results.
/// - Result rows are not tappable until the profile and player routes exist.
library;

import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/features/home/presentation/feed_section.dart';
import 'package:letslive_mobile/features/search/domain/search_type.dart';
import 'package:letslive_mobile/features/search/presentation/search_all_tab.dart';
import 'package:letslive_mobile/features/search/presentation/search_result_rows.dart';
import 'package:letslive_mobile/features/search/presentation/search_type_tab.dart';
import 'package:letslive_mobile/features/search/search_providers.dart';
import 'package:letslive_mobile/core/validation/field_limits.dart';

/// The Search tab's screen.
class SearchScreen extends StatefulWidget {
  const SearchScreen({super.key});

  @override
  State<SearchScreen> createState() => _SearchScreenState();
}

class _SearchScreenState extends State<SearchScreen>
    with SingleTickerProviderStateMixin {
  final _controller = TextEditingController();
  late final TabController _tabs = TabController(
    length: SearchType.values.length,
    vsync: this,
  );
  Timer? _debounce;

  /// The trimmed text that is searched; empty means "show the prompt".
  String _query = '';
  SearchType _tab = SearchType.all;

  void _onChanged(String text) {
    _debounce?.cancel();
    // Rebuilds the clear button.
    setState(() {});
    if (text.trim().isEmpty) {
      _apply(text);
    } else {
      _debounce = Timer(searchDebounce, () => _apply(text));
    }
  }

  void _apply(String text) {
    _debounce?.cancel();
    final query = text.trim();
    if (query != _query) setState(() => _query = query);
  }

  void _clear() {
    _controller.clear();
    _onChanged('');
  }

  void _selectTab(SearchType tab) => setState(() => _tab = tab);

  @override
  void dispose() {
    _debounce?.cancel();
    _controller.dispose();
    _tabs.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Scaffold(
      body: SafeArea(
        child: Column(
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 12, 16, 4),
              child: TextField(
                key: const ValueKey('search-input'),
                controller: _controller,
                textInputAction: TextInputAction.search,
                inputFormatters: [
                  LengthLimitingTextInputFormatter(FieldLimits.searchQueryMax),
                ],
                onChanged: _onChanged,
                onSubmitted: _apply,
                decoration: InputDecoration(
                  hintText: l10n.common_search_placeholder,
                  border: const OutlineInputBorder(),
                  isDense: true,
                  prefixIcon: const Icon(Icons.search),
                  suffixIcon: _controller.text.isEmpty
                      ? null
                      : IconButton(
                          tooltip: l10n.accessibility_clear_search,
                          icon: const Icon(Icons.close),
                          onPressed: _clear,
                        ),
                ),
              ),
            ),
            TabBar(
              controller: _tabs,
              isScrollable: true,
              tabAlignment: TabAlignment.start,
              onTap: (index) => _selectTab(SearchType.values[index]),
              tabs: [
                Tab(text: l10n.common_search_tabs_all),
                Tab(text: l10n.common_search_tabs_users),
                Tab(text: l10n.common_search_tabs_lives),
                Tab(text: l10n.common_search_tabs_vods),
              ],
            ),
            Expanded(child: _body(l10n)),
          ],
        ),
      ),
    );
  }

  Widget _body(AppLocalizations l10n) {
    if (_query.isEmpty) {
      return ListView(
        children: [
          EmptyNotice(
            icon: Icons.search,
            title: l10n.common_search_empty_prompt,
          ),
        ],
      );
    }
    return switch (_tab) {
      SearchType.all => SearchAllTab(query: _query, onSeeAll: _seeAll),
      SearchType.users => SearchTypeTab(
        provider: usersSearchProvider(_query),
        emptyTitle: l10n.common_search_no_results_type_users,
        rowBuilder: (user) =>
            UserResultRow(key: ValueKey(user.id), user: user, query: _query),
      ),
      SearchType.lives => SearchTypeTab(
        provider: livesSearchProvider(_query),
        emptyTitle: l10n.common_search_no_results_type_lives,
        rowBuilder: (stream) => LiveResultRow(
          key: ValueKey(stream.id),
          stream: stream,
          query: _query,
        ),
      ),
      SearchType.vods => SearchTypeTab(
        provider: vodsSearchProvider(_query),
        emptyTitle: l10n.common_search_no_results_type_vods,
        rowBuilder: (vod) =>
            VodResultRow(key: ValueKey(vod.id), vod: vod, query: _query),
      ),
    };
  }

  /// "See all" of a group: switches to that type's tab, tab bar included.
  void _seeAll(SearchType type) {
    _tabs.animateTo(type.index);
    _selectTab(type);
  }
}
