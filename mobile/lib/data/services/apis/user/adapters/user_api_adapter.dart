import '/domain/common/users/models/user.dart';
import '/domain/common/users/models/user_registration.dart';
import '../dtos/create_user_request_dto.dart';
import '../dtos/create_user_response_dto.dart';

abstract final class UserApiAdapter {
  static CreateUserRequestDto toCreateUserRequest(
    UserRegistration registration,
  ) {
    return CreateUserRequestDto(
      name: registration.name,
      email: registration.email,
      emailVerificationToken: registration.emailVerificationToken,
      password: registration.password,
    );
  }

  static User toUser(CreateUserResponseDto response) {
    return User(
      id: response.id,
      name: response.name,
      email: response.email,
    );
  }
}
