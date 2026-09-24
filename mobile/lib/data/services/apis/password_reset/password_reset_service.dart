import '/core/result/result.dart';
import '/core/services/client_http/client/rest_client.dart';
import '/core/services/client_http/client/rest_client_request.dart';
import '/domain/common/password_reset/models/password_reset_challenge.dart';
import '../../../../domain/common/password_reset/models/password_reset_proof.dart';
import '../core/api_response_parser.dart';
import 'adapters/password_reset_api_adapter.dart';
import 'dtos/confirm_password_reset_response_dto.dart';
import 'dtos/password_reset_response_dto.dart';

class PasswordResetService {
  final RestClient _client;

  PasswordResetService(this._client);

  AsyncResult<PasswordResetChallenge> requestPasswordReset(String email) async {
    final response = await _client.post(
      RestClientRequest(
        '/password-resets',
        body: PasswordResetApiAdapter.toRequest(email).toMap(),
      ),
    );

    if (response.isFailure) {
      return Failure(response.error!);
    }

    final result = response.value!;

    if (result.statusCode != 202) {
      return Failure(
        AppError(
          statusCode: result.statusCode,
          code: AppErrorCode.invalidData,
          message:
              'Unexpected response status: expected 202, '
              'received ${result.statusCode}',
        ),
      );
    }

    final parse = ApiResponseParser.parse<PasswordResetResponseDto>(
      result.data,
      PasswordResetResponseDto.fromMap,
    );

    if (parse.isFailure) {
      return Failure(parse.error!);
    }

    return Success(
      PasswordResetApiAdapter.toChallenge(
        parse.value!,
        email: email,
      ),
    );
  }

  AsyncResult<PasswordResetProof> confirmPasswordReset({
    required String passwordResetId,
    required String code,
  }) async {
    final response = await _client.post(
      RestClientRequest(
        '/password-resets/confirm',
        body: PasswordResetApiAdapter.toConfirmRequest(
          passwordResetId: passwordResetId,
          code: code,
        ).toMap(),
      ),
    );

    if (response.isFailure) {
      return Failure(response.error!);
    }

    final result = response.value!;

    if (result.statusCode != 200) {
      return Failure(
        AppError(
          statusCode: result.statusCode,
          code: AppErrorCode.invalidData,
          message:
              'Unexpected response status: expected 200, '
              'received ${result.statusCode}',
        ),
      );
    }

    final parsed = ApiResponseParser.parse<ConfirmPasswordResetResponseDto>(
      result.data,
      ConfirmPasswordResetResponseDto.fromMap,
    );

    if (parsed.isFailure) {
      return Failure(parsed.error!);
    }

    return Success(
      PasswordResetApiAdapter.toProof(parsed.value!),
    );
  }

  AsyncResult<Unit> completePasswordReset({
    required String passwordResetId,
    required String passwordResetToken,
    required String newPassword,
  }) async {
    final response = await _client.post(
      RestClientRequest(
        '/password-resets/complete',
        body: PasswordResetApiAdapter.toCompleteRequest(
          passwordResetId: passwordResetId,
          passwordResetToken: passwordResetToken,
          newPassword: newPassword,
        ).toMap(),
      ),
    );

    if (response.isFailure) {
      return Failure(response.error!);
    }

    final result = response.value!;

    if (result.statusCode != 204) {
      return Failure(
        AppError(
          statusCode: result.statusCode,
          code: AppErrorCode.invalidData,
          message:
              'Unexpected response status: expected 204, '
              'received ${result.statusCode}',
        ),
      );
    }

    return Success(unit);
  }
}
