# Changelog

## 2026/09/24 - start/002

This change refines the API and mobile continuous integration workflows to run only for pull requests targeting the `main` branch. Path filters remain in place so each workflow runs only when its respective module or workflow definition changes.

1. **`.github/workflows/api.yml`**

   * Removed the API workflow trigger for direct pushes.
   * Restricted pull request execution to changes targeting `main`.
   * Preserved path filtering for the `api` module and its workflow definition.

2. **`.github/workflows/mobile.yml`**

   * Removed the mobile workflow trigger for direct pushes.
   * Restricted pull request execution to changes targeting `main`.
   * Preserved path filtering for the `mobile` module and its workflow definition.

### Conclusion

The API and mobile validation workflows now run exclusively for relevant pull requests into `main`, eliminating their previous execution on direct pushes.
## 2026/09/24 - start/001

### Summary

Initial project release.

### Changes

- Established the initial project structure, source code, tests, automation, and documentation.

### Conclusion

The project now has its initial versioned foundation.
