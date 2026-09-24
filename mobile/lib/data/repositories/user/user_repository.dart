import '/core/result/result.dart';
import '/domain/common/users/models/user_registration.dart';

abstract class UserRepository {
  AsyncResult<Unit> createUser(UserRegistration registration);
}
