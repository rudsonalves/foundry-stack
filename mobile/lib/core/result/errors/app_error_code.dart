/// Domain-oriented error aligned with API contract
enum AppErrorCode {
  // HTTP
  httpError,
  accountApprovalRequired,
  contactNotVerified,
  timeout,
  networkError,
  parsingError,

  unauthenticated,
  forbidden,
  notFound,
  conflict,
  invalidData,

  // Storage
  storageError,
  storageNotFound,
  storageConflict,
  storageCorrupted,
  storageExpired,

  // Generic
  unexpected,
  unknown,

  // Registration-specific
  cpfAlreadyRegistered,
}
