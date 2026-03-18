enum AuthProvider {
  google('google'),
  local('local');

  final String value;
  const AuthProvider(this.value);

  factory AuthProvider.fromString(String value) {
    return AuthProvider.values.firstWhere(
      (e) => e.value == value,
      orElse: () => AuthProvider.local,
    );
  }
}

enum UserStatus {
  normal('normal'),
  disabled('disabled');

  final String value;
  const UserStatus(this.value);

  factory UserStatus.fromString(String value) {
    return UserStatus.values.firstWhere(
      (e) => e.value == value,
      orElse: () => UserStatus.normal,
    );
  }
}

class SocialMediaLinks {
  final String? facebook;
  final String? twitter;
  final String? instagram;
  final String? linkedin;
  final String? github;
  final String? youtube;
  final String? website;
  final String? tiktok;

  const SocialMediaLinks({
    this.facebook,
    this.twitter,
    this.instagram,
    this.linkedin,
    this.github,
    this.youtube,
    this.website,
    this.tiktok,
  });

  factory SocialMediaLinks.fromJson(Map<String, dynamic> json) {
    return SocialMediaLinks(
      facebook: json['facebook'] as String?,
      twitter: json['twitter'] as String?,
      instagram: json['instagram'] as String?,
      linkedin: json['linkedin'] as String?,
      github: json['github'] as String?,
      youtube: json['youtube'] as String?,
      website: json['website'] as String?,
      tiktok: json['tiktok'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      if (facebook != null) 'facebook': facebook,
      if (twitter != null) 'twitter': twitter,
      if (instagram != null) 'instagram': instagram,
      if (linkedin != null) 'linkedin': linkedin,
      if (github != null) 'github': github,
      if (youtube != null) 'youtube': youtube,
      if (website != null) 'website': website,
      if (tiktok != null) 'tiktok': tiktok,
    };
  }
}

class LivestreamInformation {
  final String? title;
  final String? description;
  final String? thumbnailUrl;

  const LivestreamInformation({
    this.title,
    this.description,
    this.thumbnailUrl,
  });

  factory LivestreamInformation.fromJson(Map<String, dynamic> json) {
    return LivestreamInformation(
      title: json['title'] as String?,
      description: json['description'] as String?,
      thumbnailUrl: json['thumbnailUrl'] as String?,
    );
  }
}

class User {
  final String id;
  final String username;
  final String email;
  final UserStatus status;
  final AuthProvider authProvider;
  final String createdAt;
  final String? displayName;
  final String? bio;
  final String? backgroundPicture;
  final String? profilePicture;
  final int followerCount;
  final LivestreamInformation livestreamInformation;
  final SocialMediaLinks? socialMediaLinks;

  // PublicUser fields
  final bool? isFollowing;

  // MeUser fields
  final String? phoneNumber;
  final String? streamAPIKey;

  const User({
    required this.id,
    required this.username,
    required this.email,
    required this.status,
    required this.authProvider,
    required this.createdAt,
    this.displayName,
    this.bio,
    this.backgroundPicture,
    this.profilePicture,
    required this.followerCount,
    required this.livestreamInformation,
    this.socialMediaLinks,
    this.isFollowing,
    this.phoneNumber,
    this.streamAPIKey,
  });

  bool get isMe => streamAPIKey != null;

  User copyWith({
    String? id,
    String? username,
    String? email,
    UserStatus? status,
    AuthProvider? authProvider,
    String? createdAt,
    String? displayName,
    String? bio,
    String? backgroundPicture,
    String? profilePicture,
    int? followerCount,
    LivestreamInformation? livestreamInformation,
    SocialMediaLinks? socialMediaLinks,
    bool? isFollowing,
    String? phoneNumber,
    String? streamAPIKey,
  }) {
    return User(
      id: id ?? this.id,
      username: username ?? this.username,
      email: email ?? this.email,
      status: status ?? this.status,
      authProvider: authProvider ?? this.authProvider,
      createdAt: createdAt ?? this.createdAt,
      displayName: displayName ?? this.displayName,
      bio: bio ?? this.bio,
      backgroundPicture: backgroundPicture ?? this.backgroundPicture,
      profilePicture: profilePicture ?? this.profilePicture,
      followerCount: followerCount ?? this.followerCount,
      livestreamInformation:
          livestreamInformation ?? this.livestreamInformation,
      socialMediaLinks: socialMediaLinks ?? this.socialMediaLinks,
      isFollowing: isFollowing ?? this.isFollowing,
      phoneNumber: phoneNumber ?? this.phoneNumber,
      streamAPIKey: streamAPIKey ?? this.streamAPIKey,
    );
  }

  factory User.fromJson(Map<String, dynamic> json) {
    return User(
      id: json['id'] as String,
      username: json['username'] as String,
      email: json['email'] as String? ?? '',
      status: UserStatus.fromString(json['status'] as String? ?? 'normal'),
      authProvider: AuthProvider.fromString(
        json['authProvider'] as String? ?? 'local',
      ),
      createdAt: json['createdAt'] as String,
      displayName: json['displayName'] as String?,
      bio: json['bio'] as String?,
      backgroundPicture: json['backgroundPicture'] as String?,
      profilePicture: json['profilePicture'] as String?,
      followerCount: json['followerCount'] as int? ?? 0,
      livestreamInformation: json['livestreamInformation'] != null
          ? LivestreamInformation.fromJson(
              json['livestreamInformation'] as Map<String, dynamic>,
            )
          : const LivestreamInformation(),
      socialMediaLinks: json['socialMediaLinks'] != null
          ? SocialMediaLinks.fromJson(
              json['socialMediaLinks'] as Map<String, dynamic>,
            )
          : null,
      isFollowing: json['isFollowing'] as bool?,
      phoneNumber: json['phoneNumber'] as String?,
      streamAPIKey: json['streamAPIKey'] as String?,
    );
  }
}
