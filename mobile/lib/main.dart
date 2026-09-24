import 'package:flutter/foundation.dart';
import 'package:material_ui/material_ui.dart';
import 'package:flutter/services.dart';

import 'app/app_widget.dart';
import 'app/dependencies/root_container.dart';

void main() {
  _registerFontLicenses();

  final rootContainer = RootContainer();
  rootContainer.initializeCore();
  rootContainer.initializeApp();
  runApp(AppWidget(rootContainer: rootContainer));
}

void _registerFontLicenses() {
  LicenseRegistry.addLicense(() async* {
    final license = await rootBundle.loadString(
      'assets/fonts/google_sans/OFL.txt',
    );
    yield LicenseEntryWithLineBreaks(['Google Sans'], license);
  });

  LicenseRegistry.addLicense(() async* {
    final license = await rootBundle.loadString(
      'assets/fonts/quicksand/OFL.txt',
    );
    yield LicenseEntryWithLineBreaks(['Quicksand'], license);
  });
}
