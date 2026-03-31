import 'package:dio/dio.dart';

import '../../../core/network/api_client.dart';
import '../../../core/network/api_endpoints.dart';
import '../../../core/network/api_response.dart';
import '../../../models/user.dart';

class UserRepository {
  final ApiClient _client;

  UserRepository(this._client);

  Future<ApiResponse<User>> getUser(String id) {
    return _client.get(
      ApiEndpoints.userById(id),
      fromJsonT: (json) => User.fromJson(json as Map<String, dynamic>),
    );
  }

  Future<ApiResponse<List<User>>> searchUsers({
    required String query,
    int page = 0,
    int pageSize = 20,
  }) {
    return _client.get(
      ApiEndpoints.usersSearch,
      queryParameters: {'username': query},
      fromJsonT: (json) => (json as List<dynamic>)
          .map((e) => User.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }

  Future<ApiResponse<List<User>>> getRecommendations({
    int page = 0,
    int pageSize = 20,
  }) {
    return _client.get(
      ApiEndpoints.usersRecommendations,
      queryParameters: {'page': page, 'page_size': pageSize},
      fromJsonT: (json) => (json as List<dynamic>)
          .map((e) => User.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }

  Future<ApiResponse<void>> followUser(String id) {
    return _client.post(ApiEndpoints.userFollow(id));
  }

  Future<ApiResponse<void>> unfollowUser(String id) {
    return _client.delete(ApiEndpoints.userUnfollow(id));
  }

  Future<ApiResponse<List<User>>> getFollowing({
    int page = 0,
    int pageSize = 20,
  }) {
    return _client.get(
      ApiEndpoints.userFollowing,
      queryParameters: {'page': page, 'page_size': pageSize},
      fromJsonT: (json) => (json as List<dynamic>)
          .map((e) => User.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }

  Future<ApiResponse<User>> updateProfile({
    String? displayName,
    String? bio,
    String? phoneNumber,
    SocialMediaLinks? socialMediaLinks,
  }) {
    final data = <String, dynamic>{};
    if (displayName != null) data['displayName'] = displayName;
    if (bio != null) data['bio'] = bio;
    if (phoneNumber != null) data['phoneNumber'] = phoneNumber;
    if (socialMediaLinks != null) {
      data['socialMediaLinks'] = socialMediaLinks.toJson();
    }

    return _client.put(
      ApiEndpoints.userMe,
      data: data,
      fromJsonT: (json) => User.fromJson(json as Map<String, dynamic>),
    );
  }

  Future<ApiResponse<User>> updateProfilePicture(String filePath) async {
    final formData = FormData.fromMap({
      'profile-picture': await MultipartFile.fromFile(filePath),
    });
    return _client.upload(
      ApiEndpoints.userProfilePicture,
      formData: formData,
      fromJsonT: (json) => User.fromJson(json as Map<String, dynamic>),
    );
  }

  Future<ApiResponse<User>> updateBackgroundPicture(String filePath) async {
    final formData = FormData.fromMap({
      'background-picture': await MultipartFile.fromFile(filePath),
    });
    return _client.upload(
      ApiEndpoints.userBackgroundPicture,
      formData: formData,
      fromJsonT: (json) => User.fromJson(json as Map<String, dynamic>),
    );
  }

  Future<ApiResponse<User>> updateLivestreamInformation({
    required String title,
    required String description,
    String? thumbnailFilePath,
    String? thumbnailUrl,
  }) async {
    final map = <String, dynamic>{'title': title, 'description': description};
    if (thumbnailFilePath != null) {
      map['thumbnail'] = await MultipartFile.fromFile(thumbnailFilePath);
    }
    if (thumbnailUrl != null) {
      map['thumbnailUrl'] = thumbnailUrl;
    }
    final formData = FormData.fromMap(map);
    return _client.upload(
      ApiEndpoints.userLivestreamInformation,
      formData: formData,
      fromJsonT: (json) => User.fromJson(json as Map<String, dynamic>),
    );
  }

  Future<ApiResponse<User>> generateApiKey() {
    return _client.patch(
      ApiEndpoints.userApiKey,
      fromJsonT: (json) => User.fromJson(json as Map<String, dynamic>),
    );
  }
}
