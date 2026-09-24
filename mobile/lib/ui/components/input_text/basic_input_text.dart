import 'package:material_ui/material_ui.dart';
import 'package:flutter/services.dart';

typedef InputValidator = String? Function(String value);

class BasicInputText extends StatefulWidget {
  final String initialValue;
  final FocusNode? focusNode;
  final String? labelText;
  final String? hintText;
  final TextInputType? keyboardType;
  final bool enabled;
  final TextStyle style;
  final bool obscureText;
  final Iterable<String>? autofillHints;
  final TextCapitalization textCapitalization;
  final List<TextInputFormatter>? inputFormatters;
  final TextInputAction? textInputAction;
  final InputValidator? validator;
  final bool validateOnChanged;
  final void Function(String value)? onChanged;
  final void Function(String value)? onSubmitted;
  final Widget? prefixIcon;
  final Widget? suffixIcon;
  final TextEditingController? controller;

  const BasicInputText({
    super.key,
    this.initialValue = '',
    this.focusNode,
    this.labelText,
    this.hintText,
    this.keyboardType,
    this.enabled = true,
    this.style = const TextStyle(
      fontSize: 16,
      fontWeight: FontWeight.w700,
    ),
    this.obscureText = false,
    this.autofillHints,
    this.textCapitalization = TextCapitalization.none,
    this.inputFormatters,
    this.textInputAction,
    this.validator,
    this.validateOnChanged = true,
    this.onChanged,
    this.onSubmitted,
    this.prefixIcon,
    this.suffixIcon,
    this.controller,
  });

  @override
  State<BasicInputText> createState() => _BasicInputTextState();
}

class _BasicInputTextState extends State<BasicInputText> {
  late final TextEditingController _controller;
  late final bool _ownsController;

  String? _errorText;
  bool _hasInteracted = false;

  @override
  void initState() {
    super.initState();

    _ownsController = widget.controller == null;
    _controller =
        widget.controller ??
        TextEditingController(
          text: widget.initialValue,
        );
  }

  @override
  void didUpdateWidget(covariant BasicInputText oldWidget) {
    super.didUpdateWidget(oldWidget);

    if (widget.initialValue != oldWidget.initialValue &&
        widget.initialValue != _controller.text) {
      _controller.value = TextEditingValue(
        text: widget.initialValue,
        selection: TextSelection.collapsed(
          offset: widget.initialValue.length,
        ),
      );
    }
  }

  @override
  void dispose() {
    if (_ownsController) _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return TextField(
      controller: _controller,
      focusNode: widget.focusNode,
      keyboardType: widget.keyboardType,
      enabled: widget.enabled,
      style: widget.style,
      obscureText: widget.obscureText,
      autofillHints: widget.autofillHints,
      textCapitalization: widget.textCapitalization,
      inputFormatters: widget.inputFormatters,
      textInputAction: widget.textInputAction,
      onChanged: _handleChanged,
      onSubmitted: _handleSubmitted,
      decoration: InputDecoration(
        labelText: widget.labelText,
        hintText: widget.hintText,
        errorText: _hasInteracted || _errorText != null ? _errorText : null,
        hintStyle: const TextStyle(
          fontSize: 16,
          fontWeight: FontWeight.w400,
          color: Colors.grey,
        ),
        prefixIcon: widget.prefixIcon,
        suffixIcon: widget.suffixIcon,
        filled: true,
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
          borderSide: BorderSide.none,
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
          borderSide: BorderSide.none,
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
          borderSide: BorderSide.none,
        ),
        errorBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
        ),
        focusedErrorBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
        ),
      ),
    );
  }

  bool validate() {
    return _validate(_controller.text);
  }

  bool _validate(String value) {
    final error = widget.validator?.call(value);

    if (_errorText != error) {
      setState(() {
        _errorText = error;
      });
    }

    return error == null;
  }

  void _handleChanged(String value) {
    _hasInteracted = true;

    if (widget.validateOnChanged) {
      _validate(value);
    }

    widget.onChanged?.call(value);
  }

  void _handleSubmitted(String value) {
    _validate(value);
    widget.onSubmitted?.call(value);
  }
}
