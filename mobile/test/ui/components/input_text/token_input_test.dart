import 'package:foundry_stack_mobile/ui/components/input_text/token_input.dart';
import 'package:material_ui/material_ui.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  Widget subject({
    TextEditingController? controller,
    ValueChanged<String>? onChanged,
    ValueChanged<String>? onCompleted,
  }) {
    return MaterialApp(
      home: Scaffold(
        body: Center(
          child: TokenInput(
            controller: controller,
            onChanged: onChanged,
            onCompleted: onCompleted,
          ),
        ),
      ),
    );
  }

  testWidgets('accepts six digits and preserves leading zero', (tester) async {
    final controller = TextEditingController();
    addTearDown(controller.dispose);

    await tester.pumpWidget(subject(controller: controller));
    await tester.enterText(find.byType(TextField), '012345');
    await tester.pump();

    expect(controller.text, '012345');
    expect(find.text('0'), findsOneWidget);
  });

  testWidgets('filters non-numeric characters when pasting', (tester) async {
    final controller = TextEditingController();
    addTearDown(controller.dispose);
    tester.binding.defaultBinaryMessenger.setMockMethodCallHandler(
      SystemChannels.platform,
      (call) async {
        if (call.method == 'Clipboard.getData') {
          return {'text': 'a01-23b4567'};
        }
        return null;
      },
    );
    addTearDown(
      () => tester.binding.defaultBinaryMessenger.setMockMethodCallHandler(
        SystemChannels.platform,
        null,
      ),
    );

    await tester.pumpWidget(subject(controller: controller));
    await tester.longPress(find.byType(TokenInput));
    await tester.pumpAndSettle();

    expect(controller.text, '012345');
  });

  testWidgets('allows correcting an entered token', (tester) async {
    final controller = TextEditingController();
    addTearDown(controller.dispose);

    await tester.pumpWidget(subject(controller: controller));
    await tester.enterText(find.byType(TextField), '012345');
    await tester.enterText(find.byType(TextField), '012346');
    await tester.pump();

    expect(controller.text, '012346');
    expect(find.text('6'), findsOneWidget);
  });
}
