import '/core/result/result.dart';
import '/data/repositories/user/user_repository.dart';
import '/domain/common/users/models/user_registration.dart';

class UsersUsecase {
  final UserRepository _userRepository;

  UsersUsecase({
    required UserRepository userRepository,
  }) : _userRepository = userRepository;

  AsyncResult<Unit> createUser(UserRegistration registration) =>
      _userRepository.createUser(registration);
}
