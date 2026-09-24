import '/core/result/result.dart';
import '/core/services/logging/console_log.dart';
import '/domain/common/users/models/user_registration.dart';
import '../../services/apis/user/user_service.dart';
import 'user_repository.dart';

class UserRepositoryImpl implements UserRepository {
  final UserService _userService;

  UserRepositoryImpl({
    required UserService userService,
  }) : _userService = userService;

  final _log = ConsoleLog('UserRepositoryImpl');

  @override
  AsyncResult<Unit> createUser(UserRegistration registration) async {
    final result = await _userService.createUser(registration);

    if (result.isFailure) {
      _log.error('Create user failed: ${result.error}');
      return Failure(result.error!);
    }

    return const Success(unit);
  }
}
