import 'package:flutter/material.dart';
import 'package:forui/forui.dart';

import '../../core/emotes/emote_registry.dart';

/// A bottom-sheet emote picker that inserts shortcodes into a [TextEditingController].
class EmotePicker extends StatefulWidget {
  final void Function(String code) onSelect;

  const EmotePicker({super.key, required this.onSelect});

  @override
  State<EmotePicker> createState() => _EmotePickerState();
}

class _EmotePickerState extends State<EmotePicker> {
  String _search = '';

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;

    final filtered = _search.isEmpty
        ? emotes
        : emotes
              .where(
                (e) =>
                    e.code.contains(_search.toLowerCase()) ||
                    e.name.toLowerCase().contains(_search.toLowerCase()),
              )
              .toList();

    return Container(
      constraints: const BoxConstraints(maxHeight: 320),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          // Search bar
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
            child: TextField(
              autofocus: false,
              decoration: InputDecoration(
                hintText: 'Search emotes...',
                hintStyle: typography.sm.copyWith(
                  color: colors.mutedForeground,
                ),
                prefixIcon: Icon(
                  FIcons.search,
                  size: 16,
                  color: colors.mutedForeground,
                ),
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(12),
                  borderSide: BorderSide(color: colors.border),
                ),
                contentPadding: const EdgeInsets.symmetric(
                  horizontal: 12,
                  vertical: 8,
                ),
                isDense: true,
              ),
              style: typography.sm,
              onChanged: (v) => setState(() => _search = v),
            ),
          ),

          // Emote grid
          Flexible(
            child: _search.isNotEmpty
                ? _buildGrid(filtered)
                : ListView(
                    shrinkWrap: true,
                    children: emoteCategories.map((category) {
                      final categoryEmotes =
                          emotes.where((e) => e.category == category).toList();
                      return Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Padding(
                            padding: const EdgeInsets.fromLTRB(16, 8, 16, 4),
                            child: Text(
                              category.toUpperCase(),
                              style: typography.xs.copyWith(
                                color: colors.mutedForeground,
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                          ),
                          _buildGrid(categoryEmotes),
                        ],
                      );
                    }).toList(),
                  ),
          ),
        ],
      ),
    );
  }

  Widget _buildGrid(List<Emote> items) {
    if (items.isEmpty) {
      return Padding(
        padding: const EdgeInsets.all(16),
        child: Center(
          child: Text(
            'No emotes found',
            style: context.theme.typography.xs.copyWith(
              color: context.theme.colors.mutedForeground,
            ),
          ),
        ),
      );
    }

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 8),
      child: GridView.builder(
        shrinkWrap: true,
        physics: const NeverScrollableScrollPhysics(),
        gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
          crossAxisCount: 8,
          mainAxisSpacing: 2,
          crossAxisSpacing: 2,
        ),
        itemCount: items.length,
        itemBuilder: (context, index) {
          final emote = items[index];
          return Material(
            color: Colors.transparent,
            child: InkWell(
              borderRadius: BorderRadius.circular(8),
              onTap: () => widget.onSelect(':${emote.code}:'),
              child: Center(
                child: Text(emote.emoji, style: const TextStyle(fontSize: 22)),
              ),
            ),
          );
        },
      ),
    );
  }
}

/// Show the emote picker as a bottom sheet and return the selected shortcode.
void showEmotePicker(BuildContext context, void Function(String code) onSelect) {
  showModalBottomSheet(
    context: context,
    builder: (context) => SafeArea(
      child: EmotePicker(
        onSelect: (code) {
          Navigator.pop(context);
          onSelect(code);
        },
      ),
    ),
  );
}
