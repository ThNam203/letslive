class Emote {
  final String code;
  final String emoji;
  final String name;
  final String category;

  const Emote({
    required this.code,
    required this.emoji,
    required this.name,
    required this.category,
  });
}

const emotes = <Emote>[
  // Smileys
  Emote(code: 'smile', emoji: '😊', name: 'Smile', category: 'smileys'),
  Emote(code: 'laugh', emoji: '😂', name: 'Laugh', category: 'smileys'),
  Emote(code: 'love', emoji: '😍', name: 'Love', category: 'smileys'),
  Emote(code: 'cool', emoji: '😎', name: 'Cool', category: 'smileys'),
  Emote(code: 'thinking', emoji: '🤔', name: 'Thinking', category: 'smileys'),
  Emote(code: 'cry', emoji: '😢', name: 'Cry', category: 'smileys'),
  Emote(code: 'angry', emoji: '😡', name: 'Angry', category: 'smileys'),
  Emote(code: 'wink', emoji: '😉', name: 'Wink', category: 'smileys'),
  Emote(code: 'sus', emoji: '🤨', name: 'Sus', category: 'smileys'),
  Emote(code: 'skull', emoji: '💀', name: 'Skull', category: 'smileys'),

  // Gestures
  Emote(
    code: 'thumbsup',
    emoji: '👍',
    name: 'Thumbs Up',
    category: 'gestures',
  ),
  Emote(
    code: 'thumbsdown',
    emoji: '👎',
    name: 'Thumbs Down',
    category: 'gestures',
  ),
  Emote(code: 'clap', emoji: '👏', name: 'Clap', category: 'gestures'),
  Emote(code: 'wave', emoji: '👋', name: 'Wave', category: 'gestures'),
  Emote(code: 'pray', emoji: '🙏', name: 'Pray', category: 'gestures'),
  Emote(code: 'salute', emoji: '🫡', name: 'Salute', category: 'gestures'),

  // Hype
  Emote(code: 'fire', emoji: '🔥', name: 'Fire', category: 'hype'),
  Emote(code: 'heart', emoji: '❤️', name: 'Heart', category: 'hype'),
  Emote(code: 'star', emoji: '⭐', name: 'Star', category: 'hype'),
  Emote(code: 'hype', emoji: '🎉', name: 'Hype', category: 'hype'),
  Emote(code: 'gg', emoji: '🏆', name: 'GG', category: 'hype'),
  Emote(code: 'pog', emoji: '😮', name: 'Pog', category: 'hype'),
  Emote(code: 'ez', emoji: '😏', name: 'EZ', category: 'hype'),
  Emote(code: 'goat', emoji: '🐐', name: 'GOAT', category: 'hype'),

  // Misc
  Emote(code: 'eyes', emoji: '👀', name: 'Eyes', category: 'misc'),
  Emote(code: '100', emoji: '💯', name: '100', category: 'misc'),
  Emote(code: 'money', emoji: '💰', name: 'Money', category: 'misc'),
  Emote(code: 'ghost', emoji: '👻', name: 'Ghost', category: 'misc'),
  Emote(code: 'rocket', emoji: '🚀', name: 'Rocket', category: 'misc'),
  Emote(code: 'crown', emoji: '👑', name: 'Crown', category: 'misc'),
];

final emoteMap = Map.fromEntries(emotes.map((e) => MapEntry(e.code, e)));

const emoteCategories = ['smileys', 'gestures', 'hype', 'misc'];
