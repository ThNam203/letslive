import 'package:letslive_mobile/core/network/json_fields.dart';

/// A viewer's reaction to a video (`VODReaction` in `web/types/vod.ts`).
enum VodReaction {
  like,
  dislike;

  static VodReaction _fromWire(String value) => switch (value) {
    'like' => like,
    'dislike' => dislike,
    _ => throw FormatException('VodReactionState.reaction: got "$value"'),
  };
}

/// The signed-in viewer's own reaction plus the public like count
/// (`VODReactionState`). Dislikes are never counted publicly.
class VodReactionState {
  const VodReactionState({required this.likeCount, this.reaction});

  factory VodReactionState.fromJson(Map<String, Object?> json) {
    final f = JsonFields('VodReactionState', json);
    final reaction = f.optString('reaction');
    return VodReactionState(
      likeCount: f.integer('likeCount'),
      reaction: reaction == null ? null : VodReaction._fromWire(reaction),
    );
  }

  /// `ParseData<VodReactionState>` for `ApiClient`.
  static VodReactionState parse(Object? data) =>
      parseJsonObject(data, 'VodReactionState', VodReactionState.fromJson);

  final int likeCount;

  /// What the viewer reacted with; null for no reaction.
  final VodReaction? reaction;

  /// The state after the viewer's reaction changed to [next]: web's
  /// optimistic update (`likeDelta`), never below zero likes.
  VodReactionState reactedWith(VodReaction? next) {
    final delta =
        (next == VodReaction.like ? 1 : 0) -
        (reaction == VodReaction.like ? 1 : 0);
    final count = likeCount + delta;
    return VodReactionState(likeCount: count < 0 ? 0 : count, reaction: next);
  }
}
