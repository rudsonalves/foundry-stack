import 'package:flutter_test/flutter_test.dart';
import 'package:foundry_stack_mobile/core/result/result.dart';
import 'package:foundry_stack_mobile/data/repositories/user/user_repository_impl.dart';
import 'package:foundry_stack_mobile/domain/common/users/models/user_registration.dart';
import 'package:foundry_stack_mobile/domain/common/users/models/user.dart';
import 'package:foundry_stack_mobile/data/services/apis/user/user_service.dart';

void main() {
  group('UserRepositoryImpl.createUser', () {
    late _FakeUserService userService;
    late UserRepositoryImpl repository;

    setUp(() {
      userService = _FakeUserService();
      repository = UserRepositoryImpl(userService: userService);
    });

    test('returns Unit after successful registration', () async {
      userService.createUserResult = Success(
        const User(
          id: 'user-id',
          name: 'User name',
          email: 'user@example.com',
        ),
      );
      final request = _createUserRequest();

      final result = await repository.createUser(request);

      expect(result.isSuccess, isTrue);
      expect(result.value, unit);
      expect(identical(userService.lastCreateUserRequest, request), isTrue);
    });

    test('preserves a registration conflict', () async {
      const error = AppError(
        statusCode: 409,
        code: AppErrorCode.conflict,
        message: 'registration conflict',
      );
      userService.createUserResult = const Failure(error);

      final result = await repository.createUser(_createUserRequest());

      expect(result.isFailure, isTrue);
      expect(identical(result.error, error), isTrue);
    });

    test('preserves a registration network failure', () async {
      const error = AppError(
        code: AppErrorCode.networkError,
        message: 'network unavailable',
      );
      userService.createUserResult = const Failure(error);

      final result = await repository.createUser(_createUserRequest());

      expect(result.isFailure, isTrue);
      expect(identical(result.error, error), isTrue);
    });
  });
}

UserRegistration _createUserRequest() {
  return const UserRegistration(
    name: 'User name',
    email: 'user@example.com',
    password: 'password-fixture',
    emailVerificationToken: 'verification-proof',
  );
}

class _FakeUserService implements UserService {
  Result<User>? createUserResult;
  UserRegistration? lastCreateUserRequest;

  @override
  AsyncResult<User> createUser(
    UserRegistration request,
  ) async {
    lastCreateUserRequest = request;
    return createUserResult!;
  }
}
