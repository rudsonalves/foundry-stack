import 'package:flutter_test/flutter_test.dart';
import 'package:foundry_stack_mobile/data/services/apis/auth/dtos/refresh_request_dto.dart';

void main() {
  group('RefreshRequestDto', () {
    test('toMap serializes the request body', () {
      final dto = RefreshRequestDto(refreshToken: 'refresh-token');

      expect(dto.toMap(), {
        'refresh_token': 'refresh-token',
      });
    });
  });
}
