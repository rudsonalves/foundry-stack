import 'package:material_ui/material_ui.dart';

class TextHeader extends StatelessWidget {
  final String text;

  const TextHeader(
    this.text, {
    super.key,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.only(top: 24, bottom: 12),
      child: Text(
        text,
      ),
    );
  }
}
