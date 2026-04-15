import '../../../../models/chat_message.dart';

/// Discriminated union for what shows up in the live-chat scroll area.
sealed class ChatLine {
  const ChatLine();
}

/// A message received from the chat WebSocket / history endpoint.
class RemoteLine extends ChatLine {
  final ChatMessage data;
  const RemoteLine(this.data);
}

/// A locally-generated line visible only to the sender — e.g. `/help` output
/// or "unknown chat command" errors.
class SystemLine extends ChatLine {
  final String text;
  const SystemLine(this.text);
}
