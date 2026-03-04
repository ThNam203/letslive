class AppNotification {
  final String id;
  final String userId;
  final String type;
  final String title;
  final String message;
  final String? actionUrl;
  final String? actionLabel;
  final String? referenceId;
  final bool isRead;
  final String createdAt;

  const AppNotification({
    required this.id,
    required this.userId,
    required this.type,
    required this.title,
    required this.message,
    this.actionUrl,
    this.actionLabel,
    this.referenceId,
    this.isRead = false,
    required this.createdAt,
  });

  factory AppNotification.fromJson(Map<String, dynamic> json) {
    return AppNotification(
      id: json['id'] as String,
      userId: json['userId'] as String,
      type: json['type'] as String,
      title: json['title'] as String,
      message: json['message'] as String,
      actionUrl: json['actionUrl'] as String?,
      actionLabel: json['actionLabel'] as String?,
      referenceId: json['referenceId'] as String?,
      isRead: json['isRead'] as bool? ?? false,
      createdAt: json['createdAt'] as String,
    );
  }
}
