import 'email_verification_challenge.dart';
import 'email_verification_proof.dart';
import 'onboarding_step.dart';

class OnboardingRegistrationDraft {
  static const currentSchemaVersion = 1;
  static const validity = Duration(hours: 24);

  final int schemaVersion;
  final OnboardingStep step;
  final String? name;
  final String? email;
  final EmailVerificationChallenge? challenge;
  final EmailVerificationProof? proof;
  final DateTime expiresAt;

  const OnboardingRegistrationDraft({
    this.schemaVersion = currentSchemaVersion,
    required this.step,
    this.name,
    this.email,
    this.challenge,
    this.proof,
    required this.expiresAt,
  });

  OnboardingRegistrationDraft copyWith({
    int? schemaVersion,
    OnboardingStep? step,
    String? name,
    String? email,
    EmailVerificationChallenge? challenge,
    EmailVerificationProof? proof,
    DateTime? expiresAt,
    bool clearChallenge = false,
    bool clearProof = false,
  }) {
    return OnboardingRegistrationDraft(
      schemaVersion: schemaVersion ?? this.schemaVersion,
      step: step ?? this.step,
      name: name ?? this.name,
      email: email ?? this.email,
      challenge: clearChallenge ? null : challenge ?? this.challenge,
      proof: clearProof ? null : proof ?? this.proof,
      expiresAt: expiresAt ?? this.expiresAt,
    );
  }

  bool isExpiredAt(DateTime instant) => !instant.isBefore(expiresAt);
}
