import 'package:flutter/material.dart';
import 'package:letslive_mobile/core/chat_parser/emote.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';

/// Opens the emote picker as a bottom sheet; completes with the emote the
/// user tapped, or null when dismissed.
Future<Emote?> showEmotePicker(BuildContext context) =>
    showModalBottomSheet<Emote>(
      context: context,
      isScrollControlled: true,
      showDragHandle: true,
      builder: (_) => const _EmotePickerSheet(),
    );

/// Search box plus the emotes by category, or the matches while searching
/// (web `EmotePicker`).
class _EmotePickerSheet extends StatefulWidget {
  const _EmotePickerSheet();

  @override
  State<_EmotePickerSheet> createState() => _EmotePickerSheetState();
}

class _EmotePickerSheetState extends State<_EmotePickerSheet> {
  String _query = '';

  String _categoryLabel(AppLocalizations l10n, EmoteCategory category) =>
      switch (category) {
        EmoteCategory.smileys => l10n.chat_emote_category_smileys,
        EmoteCategory.gestures => l10n.chat_emote_category_gestures,
        EmoteCategory.hype => l10n.chat_emote_category_hype,
        EmoteCategory.misc => l10n.chat_emote_category_misc,
      };

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final theme = Theme.of(context);
    final matches = searchEmotes(_query);
    final Widget body;
    if (_query.isEmpty) {
      body = ListView(
        padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
        children: [
          for (final category in EmoteCategory.values) ...[
            Padding(
              padding: const EdgeInsets.only(top: 8, bottom: 4),
              child: Text(
                _categoryLabel(l10n, category),
                style: theme.textTheme.labelMedium,
              ),
            ),
            _EmoteWrap([
              for (final e in emotes)
                if (e.category == category) e,
            ]),
          ],
        ],
      );
    } else if (matches.isEmpty) {
      body = Center(child: Text(l10n.chat_emote_empty_state));
    } else {
      body = ListView(
        padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
        children: [_EmoteWrap(matches)],
      );
    }

    final size = MediaQuery.sizeOf(context);
    return Padding(
      padding: EdgeInsets.only(bottom: MediaQuery.viewInsetsOf(context).bottom),
      child: SizedBox(
        height: size.height * 0.5,
        child: Column(
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
              child: TextField(
                autofocus: false,
                decoration: InputDecoration(
                  hintText: l10n.chat_emote_search_placeholder,
                  prefixIcon: const Icon(Icons.search),
                  isDense: true,
                ),
                onChanged: (value) => setState(() => _query = value.trim()),
              ),
            ),
            Expanded(child: body),
          ],
        ),
      ),
    );
  }
}

/// Tappable emoji buttons that wrap to the width.
class _EmoteWrap extends StatelessWidget {
  const _EmoteWrap(this.items);

  final List<Emote> items;

  @override
  Widget build(BuildContext context) => Wrap(
    children: [
      for (final e in items)
        Tooltip(
          message: e.shortcode,
          child: InkWell(
            borderRadius: BorderRadius.circular(8),
            onTap: () => Navigator.of(context).pop(e),
            child: Semantics(
              label: e.name,
              child: SizedBox(
                width: 48,
                height: 48,
                child: Center(
                  child: Text(e.emoji, style: const TextStyle(fontSize: 24)),
                ),
              ),
            ),
          ),
        ),
    ],
  );
}
