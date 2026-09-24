import '/core/result/result.dart';
import '/core/services/client_http/client/rest_client.dart';
import '/core/services/client_http/client/rest_client_request.dart';
import '/core/services/logging/console_log.dart';
import '/domain/common/users/models/user.dart';
import '/domain/common/users/models/user_registration.dart';
import '../core/api_response_parser.dart';
import 'adapters/user_api_adapter.dart';
import 'dtos/create_user_response_dto.dart';

class UserService {
  final RestClient _client;

  UserService(this._client);

  final _log = ConsoleLog('UserService');

  AsyncResult<User> createUser(
    UserRegistration registration,
  ) async {
    final response = await _client.post(
      RestClientRequest(
        '/users',
        body: UserApiAdapter.toCreateUserRequest(registration).toMap(),
      ),
    );

    if (response.isFailure) {
      _log.error('createUser failed: ${response.error}');
      return Failure(response.error!);
    }

    final resp = response.value!;

    if (resp.statusCode != 201) {
      _log.error(
        'createUser failed: ${resp.statusCode} ${resp.statusMessage}',
      );

      return Failure(
        AppError(
          statusCode: resp.statusCode,
          code: AppErrorCode.invalidData,
          message:
              'Unexpected response status: expected 201, received ${resp.statusCode}',
        ),
      );
    }

    final result = ApiResponseParser.parse<CreateUserResponseDto>(
      resp.data,
      CreateUserResponseDto.fromMap,
    );

    if (result.isFailure) {
      _log.error('createUser response parsing failed: ${result.error}');
    }

    if (result.isFailure) {
      return Failure(result.error!);
    }

    return Success(UserApiAdapter.toUser(result.value!));
  }
}
