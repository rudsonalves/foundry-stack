import 'package:material_ui/material_ui.dart';
import 'package:flutter/services.dart';

import '/core/extensions/string.dart';

class TokenInput extends StatefulWidget {
  final int length;
  final String? initialValue;
  final ValueChanged<String>? onChanged;
  final ValueChanged<String>? onCompleted;
  final double spacing;
  final double fieldWidth;
  final double fieldHeight;
  final bool autoFocus;
  final FocusNode? focusNode;
  final TextEditingController? controller;
  final bool enabled;

  const TokenInput({
    super.key,
    this.length = 6,
    this.initialValue,
    this.onChanged,
    this.onCompleted,
    this.spacing = 12,
    this.fieldWidth = 53,
    this.fieldHeight = 60,
    this.autoFocus = true,
    this.focusNode,
    this.controller,
    this.enabled = true,
  });

  @override
  State<TokenInput> createState() => _TokenInputState();
}

class _TokenInputState extends State<TokenInput> {
  late final TextEditingController _controller;
  late final FocusNode _focusNode;
  late final bool _ownsFocusNode;
  late final bool _ownsController;

  String get value => _controller.text;

  @override
  void initState() {
    super.initState();

    _ownsController = widget.controller == null;
    _ownsFocusNode = widget.focusNode == null;

    final initialToken =
        (widget.controller?.text ?? widget.initialValue ?? '').onlyNumbers;
    final normalizedToken = initialToken.substring(
      0,
      initialToken.length.clamp(0, widget.length),
    );

    _controller = widget.controller ?? TextEditingController();
    _controller.value = TextEditingValue(
      text: normalizedToken,
      selection: TextSelection.collapsed(
        offset: normalizedToken.length,
      ),
    );

    _focusNode = widget.focusNode ?? FocusNode();

    _controller.addListener(_onTextChanged);
    _focusNode.addListener(_onFocusChanged);

    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (widget.autoFocus) _focusNode.requestFocus();
    });
  }

  @override
  void dispose() {
    _controller.removeListener(_onTextChanged);
    _focusNode.removeListener(_onFocusChanged);

    if (_ownsController) {
      _controller.dispose();
    }

    if (_ownsFocusNode) {
      _focusNode.dispose();
    }

    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      behavior: HitTestBehavior.opaque,
      onTap: widget.enabled ? () => _focusNode.requestFocus() : null,
      onLongPress: widget.enabled ? _pasteFromClipboard : null,
      child: SizedBox(
        height: widget.fieldHeight,
        child: Stack(
          alignment: Alignment.center,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: List.generate(
                widget.length,
                (index) => _TokenCell(
                  char: index < value.length ? value[index] : null,
                  isFocused:
                      _focusNode.hasFocus &&
                      index == value.length.clamp(0, widget.length - 1),
                  width: widget.fieldWidth,
                  height: widget.fieldHeight,
                ),
              ),
            ),
            Positioned.fill(
              child: IgnorePointer(
                child: Opacity(
                  opacity: 0,
                  child: TextField(
                    controller: _controller,
                    focusNode: _focusNode,
                    keyboardType: TextInputType.number,
                    enabled: widget.enabled,
                    autofillHints: const [AutofillHints.oneTimeCode],
                    enableInteractiveSelection: false,
                    showCursor: false,
                    textAlign: TextAlign.center,
                    inputFormatters: [
                      FilteringTextInputFormatter.digitsOnly,
                      LengthLimitingTextInputFormatter(widget.length),
                    ],
                    maxLength: widget.length,
                    decoration: const InputDecoration(
                      border: InputBorder.none,
                      counterText: '',
                      contentPadding: EdgeInsets.zero,
                    ),
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  void _onFocusChanged() => setState(() {});

  void _onTextChanged() {
    setState(() {});

    widget.onChanged?.call(value);

    if (value.length == widget.length) {
      widget.onCompleted?.call(value);
      _focusNode.unfocus();
    }
  }

  Future<void> _pasteFromClipboard() async {
    final data = await Clipboard.getData(Clipboard.kTextPlain);
    final digits = data?.text?.onlyNumbers;

    if (digits == null || digits.isEmpty) return;

    _controller.text = digits.substring(
      0,
      digits.length.clamp(0, widget.length),
    );
    _controller.selection = TextSelection.collapsed(
      offset: _controller.text.length,
    );

    if (_controller.text.length < widget.length) {
      _focusNode.requestFocus();
    }
  }
}

class _TokenCell extends StatelessWidget {
  final String? char;
  final bool isFocused;
  final double width;
  final double height;

  const _TokenCell({
    required this.char,
    required this.isFocused,
    required this.width,
    required this.height,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return SizedBox(
      width: width,
      height: height,
      child: DecoratedBox(
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(8),
          border: Border.all(
            color: isFocused
                ? theme.colorScheme.primary
                : theme.colorScheme.outline,
          ),
        ),
        child: Center(
          child: Text(
            char ?? '',
            style: const TextStyle(fontWeight: FontWeight.w700),
          ),
        ),
      ),
    );
  }
}
