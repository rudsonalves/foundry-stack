import 'dart:convert';

import 'package:foundry_stack_mobile/data/repositories/onboarding/codec/onboarding_draft_codec.dart';
import 'package:foundry_stack_mobile/domain/common/onboarding/models/email_verification_challenge.dart';
import 'package:foundry_stack_mobile/domain/common/onboarding/models/email_verification_proof.dart';
import 'package:foundry_stack_mobile/domain/common/onboarding/models/onboarding_registration_draft.dart';
import 'package:foundry_stack_mobile/domain/common/onboarding/models/onboarding_step.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('OnboardingDraftCodec', () {
    test('round-trips a complete registration draft', () {
      final original = _completeDraft();

      final restored = OnboardingDraftCodec.decode(
        OnboardingDraftCodec.encode(original),
      );

      expect(restored.schemaVersion, original.schemaVersion);
      expect(restored.step, original.step);
      expect(restored.name, original.name);
      expect(restored.email, original.email);
      expect(
        restored.challenge?.verificationId,
        original.challenge?.verificationId,
      );
      expect(
        restored.challenge?.codeExpiresAt,
        original.challenge?.codeExpiresAt,
      );
      expect(
        restored.challenge?.resendAvailableAt,
        original.challenge?.resendAvailableAt,
      );
      expect(restored.proof?.token, original.proof?.token);
      expect(restored.proof?.expiresAt, original.proof?.expiresAt);
      expect(restored.expiresAt, original.expiresAt);
    });

    test('does not serialize transient or credential fields', () {
      final encoded = OnboardingDraftCodec.encode(_completeDraft());
      final map = jsonDecode(encoded) as Map<String, dynamic>;

      expect(map, isNot(contains('code')));
      expect(map, isNot(contains('password')));
      expect(map, isNot(contains('password_confirmation')));
      expect(map, isNot(contains('loading')));
      expect(map, isNot(contains('error')));
    });

    test('rejects an incompatible schema version', () {
      expect(
        () => OnboardingDraftCodec.decode(
          jsonEncode({'schema_version': 1}),
        ),
        throwsA(isA<FormatException>()),
      );
    });

    test('rejects a missing draft expiration', () {
      expect(
        () => OnboardingDraftCodec.decode(
          _jsonFor(step: 'user_name', expiresAt: null),
        ),
        throwsA(isA<FormatException>()),
      );
    });

    test('rejects an invalid draft expiration', () {
      expect(
        () => OnboardingDraftCodec.decode(
          _jsonFor(step: 'user_name', expiresAt: 'invalid-date'),
        ),
        throwsA(isA<FormatException>()),
      );
    });

    test('rejects a partial challenge', () {
      expect(
        () => OnboardingDraftCodec.decode(
          _jsonFor(
            step: 'email_verification',
            name: 'Ada',
            email: 'ada@example.com',
            verificationId: 'verification-id',
          ),
        ),
        throwsA(isA<FormatException>()),
      );
    });

    test('rejects a partial proof', () {
      expect(
        () => OnboardingDraftCodec.decode(
          _jsonFor(
            step: 'password',
            name: 'Ada',
            email: 'ada@example.com',
            token: 'opaque-secret',
          ),
        ),
        throwsA(isA<FormatException>()),
      );
    });

    test('rejects a step without its required data', () {
      expect(
        () => OnboardingDraftCodec.decode(
          _jsonFor(step: 'email_verification', name: 'Ada'),
        ),
        throwsA(isA<FormatException>()),
      );
    });
  });
}

OnboardingRegistrationDraft _completeDraft() {
  return OnboardingRegistrationDraft(
    step: OnboardingStep.password,
    name: 'Ada',
    email: 'ada@example.com',
    challenge: EmailVerificationChallenge(
      verificationId: 'verification-id',
      codeExpiresAt: DateTime.utc(2026, 9, 18, 15, 15),
      resendAvailableAt: DateTime.utc(2026, 9, 18, 15, 1),
    ),
    proof: EmailVerificationProof(
      token: 'opaque-secret',
      expiresAt: DateTime.utc(2026, 9, 19, 15),
    ),
    expiresAt: DateTime.utc(2026, 9, 19, 16),
  );
}

String _jsonFor({
  required String step,
  String? name,
  String? email,
  String? verificationId,
  String? codeExpiresAt,
  String? resendAvailableAt,
  String? token,
  String? proofExpiresAt,
  Object? expiresAt = '2026-09-19T16:00:00.000Z',
}) {
  return jsonEncode({
    'schema_version': OnboardingRegistrationDraft.currentSchemaVersion,
    'step': step,
    'name': name,
    'email': email,
    'verification_id': verificationId,
    'code_expires_at': codeExpiresAt,
    'resend_available_at': resendAvailableAt,
    'email_verification_token': token,
    'proof_expires_at': proofExpiresAt,
    'expires_at': expiresAt,
  });
}
