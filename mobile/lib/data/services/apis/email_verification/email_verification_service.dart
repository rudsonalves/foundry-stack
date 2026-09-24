import '/core/result/result.dart';
import '/core/services/client_http/client/rest_client.dart';
import '/core/services/client_http/client/rest_client_request.dart';
import '/domain/common/onboarding/models/email_verification_challenge.dart';
import '/domain/common/onboarding/models/email_verification_proof.dart';
import '../core/api_response_parser.dart';
import 'adapters/email_verification_api_adapter.dart';
import 'dtos/confirm_email_verification_response_dto.dart';
import 'dtos/email_verification_response_dto.dart';

class EmailVerificationService {
  final RestClient _client;

  EmailVerificationService(this._client);

  AsyncResult<EmailVerificationChallenge> requestVerification(
    String email,
  ) async {
    final response = await _client.post(
      RestClientRequest(
        '/email-verifications',
        body: EmailVerificationApiAdapter.toRequest(email).toMap(),
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

    final parsed = ApiResponseParser.parse<EmailVerificationResponseDto>(
      result.data,
      EmailVerificationResponseDto.fromMap,
    );

    if (parsed.isFailure) {
      return Failure(parsed.error!);
    }

    return Success(
      EmailVerificationApiAdapter.toChallenge(parsed.value!),
    );
  }

  AsyncResult<EmailVerificationProof> confirmVerification({
    required String verificationId,
    required String code,
  }) async {
    final response = await _client.post(
      RestClientRequest(
        '/email-verifications/confirm',
        body: EmailVerificationApiAdapter.toConfirmRequest(
          verificationId: verificationId,
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

    final parse = ApiResponseParser.parse<ConfirmEmailVerificationResponseDto>(
      result.data,
      ConfirmEmailVerificationResponseDto.fromMap,
    );

    if (parse.isFailure) {
      return Failure(parse.error!);
    }

    return Success(
      EmailVerificationApiAdapter.toProof(parse.value!),
    );
  }
}
