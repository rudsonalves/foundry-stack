import 'dart:convert';

import '/domain/common/onboarding/models/email_verification_challenge.dart';
import '/domain/common/onboarding/models/email_verification_proof.dart';
import '/domain/common/onboarding/models/onboarding_registration_draft.dart';
import '/domain/common/onboarding/models/onboarding_step.dart';

abstract final class OnboardingDraftCodec {
  static String encode(OnboardingRegistrationDraft draft) {
    return jsonEncode({
      'schema_version': draft.schemaVersion,
      'step': _stepToJson(draft.step),
      'name': draft.name,
      'email': draft.email,
      'verification_id': draft.challenge?.verificationId,
      'code_expires_at': draft.challenge?.codeExpiresAt
          .toUtc()
          .toIso8601String(),
      'resend_available_at': draft.challenge?.resendAvailableAt
          .toUtc()
          .toIso8601String(),
      'email_verification_token': draft.proof?.token,
      'proof_expires_at': draft.proof?.expiresAt.toUtc().toIso8601String(),
      'expires_at': draft.expiresAt.toUtc().toIso8601String(),
    });
  }

  static String _stepToJson(OnboardingStep step) {
    return switch (step) {
      OnboardingStep.userName => 'user_name',
      OnboardingStep.email => 'email',
      OnboardingStep.emailVerification => 'email_verification',
      OnboardingStep.password => 'password',
    };
  }

  static OnboardingRegistrationDraft decode(String source) {
    final decoded = jsonDecode(source);

    if (decoded is! Map<String, dynamic>) {
      throw FormatException('Onboarding draft must be a JSON object.');
    }

    final json = decoded;
    final schemaVersion = json['schema_version'];

    if (schemaVersion != OnboardingRegistrationDraft.currentSchemaVersion) {
      throw const FormatException('Unsupported onboarding draft version.');
    }

    const challengeKeys = [
      'verification_id',
      'code_expires_at',
      'resend_available_at',
    ];

    const proofKeys = [
      'email_verification_token',
      'proof_expires_at',
    ];

    final hasAnyChallengeFields = _hasAny(json, challengeKeys);
    final hasAllChallengeFields = _hasAll(json, challengeKeys);
    final hasAnyProofFields = _hasAny(json, proofKeys);
    final hasAllProofFields = _hasAll(json, proofKeys);

    if (hasAnyChallengeFields && !hasAllChallengeFields) {
      throw FormatException('Incomplete email verification challenge.');
    }

    if (hasAnyProofFields && !hasAllProofFields) {
      throw FormatException('Incomplete email verification proof.');
    }

    final name = _optionalString(json, 'name');
    final email = _optionalString(json, 'email');
    final step = _stepFromJson(_requiredString(json, 'step'));

    final proof = hasAllProofFields
        ? EmailVerificationProof(
            token: _requiredString(json, 'email_verification_token'),
            expiresAt: _requiredDateTime(json, 'proof_expires_at'),
          )
        : null;

    final challenge = hasAllChallengeFields
        ? EmailVerificationChallenge(
            verificationId: _requiredString(
              json,
              'verification_id',
            ),
            codeExpiresAt: _requiredDateTime(
              json,
              'code_expires_at',
            ),
            resendAvailableAt: _requiredDateTime(
              json,
              'resend_available_at',
            ),
          )
        : null;

    final expiresAt = _requiredDateTime(json, 'expires_at');

    final draft = OnboardingRegistrationDraft(
      schemaVersion: schemaVersion,
      step: step,
      name: name,
      email: email,
      challenge: challenge,
      proof: proof,
      expiresAt: expiresAt,
    );

    switch (step) {
      case OnboardingStep.userName:
        break;
      case OnboardingStep.email:
        if (name == null) {
          throw const FormatException(
            'Name is required at email step.',
          );
        }
      case OnboardingStep.emailVerification:
        if (name == null || email == null || challenge == null) {
          throw const FormatException(
            'Challenge is required at verification step.',
          );
        }
      case OnboardingStep.password:
        if (name == null || email == null || proof == null) {
          throw const FormatException(
            'Proof is required at password step.',
          );
        }
    }

    return draft;
  }

  static OnboardingStep _stepFromJson(String value) {
    return switch (value) {
      'user_name' => OnboardingStep.userName,
      'email' => OnboardingStep.email,
      'email_verification' => OnboardingStep.emailVerification,
      'password' => OnboardingStep.password,
      _ => throw FormatException('Unknown onboarding step: $value'),
    };
  }

  static String _requiredString(
    Map<String, dynamic> json,
    String key,
  ) {
    final value = json[key];

    if (value is! String || value.trim().isEmpty) {
      throw FormatException('Invalid $key.');
    }

    return value;
  }

  static String? _optionalString(
    Map<String, dynamic> json,
    String key,
  ) {
    final value = json[key];

    if (value == null) return null;

    if (value is! String || value.trim().isEmpty) {
      throw FormatException('Invalid $key.');
    }

    return value;
  }

  static DateTime _requiredDateTime(
    Map<String, dynamic> json,
    String key,
  ) {
    final value = _requiredString(json, key);
    final parsed = DateTime.tryParse(value);

    if (parsed == null) {
      throw FormatException('Invalid $key.');
    }

    return parsed;
  }

  static bool _hasAny(
    Map<String, dynamic> json,
    List<String> keys,
  ) {
    return keys.any((key) => json[key] != null);
  }

  static bool _hasAll(
    Map<String, dynamic> json,
    List<String> keys,
  ) {
    return keys.every((key) => json[key] != null);
  }
}
