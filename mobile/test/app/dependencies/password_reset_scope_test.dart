import 'package:foundry_stack_mobile/app/dependencies/password_reset_dependencies.dart';
import 'package:foundry_stack_mobile/app/dependencies/password_reset_scope.dart';
import 'package:foundry_stack_mobile/core/result/result.dart';
import 'package:foundry_stack_mobile/core/services/client_http/client/rest_client.dart';
import 'package:foundry_stack_mobile/core/services/client_http/client/rest_client_request.dart';
import 'package:foundry_stack_mobile/core/services/client_http/client/rest_client_response.dart';
import 'package:foundry_stack_mobile/core/services/auth_token_store/auth_token_store.dart';
import 'package:foundry_stack_mobile/core/services/secure_storage/local_secure_storage.dart';
import 'package:foundry_stack_mobile/data/repositories/password_reset/password_reset_repository.dart';
import 'package:foundry_stack_mobile/data/services/apis/password_reset/password_reset_service.dart';
import 'package:foundry_stack_mobile/domain/common/password_reset/models/password_reset_step.dart';
import 'package:foundry_stack_mobile/domain/usecases/password_reset/password_reset_usecase.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  late _FakeRestClient restClient;
  final scopes = <PasswordResetScope>[];

  setUp(() {
    restClient = _FakeRestClient();
  });

  tearDown(() {
    for (final scope in scopes) {
      scope.dispose();
    }
    scopes.clear();
  });

  PasswordResetScope createScope({bool initialize = true}) {
    final scope = PasswordResetScope(
      PasswordResetDependencies(restClient: restClient),
    );
    scopes.add(scope);
    if (initialize) scope.initialize();
    return scope;
  }

  test('blocks resolution before initialization', () {
    final scope = createScope(initialize: false);

    expect(() => scope.get<PasswordResetUsecase>(), throwsStateError);
  });

  test('can only be initialized once', () {
    final scope = createScope();

    expect(scope.initialize, throwsStateError);
  });

  test('contains one complete password reset graph', () {
    final scope = createScope();

    expect(identical(scope.get<RestClient>(), restClient), isTrue);
    expect(scope.get<PasswordResetService>(), isA<PasswordResetService>());
    expect(
      scope.get<PasswordResetRepository>(),
      isA<PasswordResetRepository>(),
    );

    final first = scope.get<PasswordResetUsecase>();
    final second = scope.get<PasswordResetUsecase>();
    expect(identical(first, second), isTrue);
  });

  test('does not expose authentication or onboarding storages', () {
    final scope = createScope();

    expect(() => scope.get<AuthTokenStore>(), throwsA(anything));
    expect(() => scope.get<LocalSecureStorage>(), throwsA(anything));
  });

  test('dispose is idempotent and clears the use case state', () async {
    restClient.postResult = const Success(
      RestClientResponse(
        statusCode: 202,
        data: {
          'data': {
            'password_reset_id': 'reset-id',
            'code_expires_at': '2026-09-23T15:15:00Z',
            'resend_available_at': '2026-09-23T15:01:00Z',
          },
        },
      ),
    );
    final scope = createScope();
    final usecase = scope.get<PasswordResetUsecase>();
    await usecase.requestPasswordReset('user@example.com');

    scope.dispose();
    scope.dispose();

    expect(usecase.step, PasswordResetStep.email);
    expect(usecase.email, isNull);
    expect(usecase.challenge, isNull);
    expect(usecase.proof, isNull);
    expect(() => scope.get<PasswordResetUsecase>(), throwsStateError);
  });

  test('isolates use cases between journeys while sharing RestClient', () {
    final firstScope = createScope();
    final secondScope = createScope();

    expect(
      identical(
        firstScope.get<PasswordResetUsecase>(),
        secondScope.get<PasswordResetUsecase>(),
      ),
      isFalse,
    );
    expect(
      identical(firstScope.get<RestClient>(), secondScope.get<RestClient>()),
      isTrue,
    );
  });
}

final class _FakeRestClient implements RestClient {
  Result<RestClientResponse>? postResult;

  @override
  AsyncResult<RestClientResponse> post(RestClientRequest request) async {
    return postResult ??
        const Failure(
          AppError(
            code: AppErrorCode.unexpected,
            message: 'POST response was not configured.',
          ),
        );
  }

  @override
  AsyncResult<RestClientResponse> delete(RestClientRequest request) =>
      throw UnimplementedError();

  @override
  AsyncResult<RestClientResponse> get(RestClientRequest request) =>
      throw UnimplementedError();

  @override
  AsyncResult<RestClientResponse> patch(RestClientRequest request) =>
      throw UnimplementedError();

  @override
  AsyncResult<RestClientResponse> put(RestClientRequest request) =>
      throw UnimplementedError();
}
