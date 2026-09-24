class CreateUserResponseDto {
  final String id;
  final String name;
  final String email;

  CreateUserResponseDto({
    required this.id,
    required this.name,
    required this.email,
  });

  factory CreateUserResponseDto.fromMap(Map<String, dynamic> map) {
    if (!map.containsKey('id') ||
        !map.containsKey('name') ||
        !map.containsKey('email') ||
        map['id'] is! String ||
        map['id'].trim().isEmpty ||
        map['name'] is! String ||
        map['name'].trim().isEmpty ||
        map['email'] is! String ||
        map['email'].trim().isEmpty) {
      throw Exception('Invalid map structure for CreateUserResponseDto');
    }

    return CreateUserResponseDto(
      id: map['id'] as String,
      name: map['name'] as String,
      email: map['email'] as String,
    );
  }
}
