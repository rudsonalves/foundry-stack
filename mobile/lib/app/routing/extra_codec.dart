import 'dart:convert';

class ExtraCodec extends Codec<Object?, String> {
  const ExtraCodec();

  @override
  Converter<Object?, String> get encoder => const _ExtraEncoder();

  @override
  Converter<String, Object?> get decoder => const _ExtraDecoder();
}

class _ExtraEncoder extends Converter<Object?, String> {
  const _ExtraEncoder();

  @override
  String convert(Object? extra) {
    if (extra is String || extra is num || extra is bool) {
      return jsonEncode({'type': 'primitive', 'data': extra});
    }

    if (extra == null) {
      return jsonEncode({'type': 'primitive', 'data': null});
    }

    throw UnsupportedError('Unsupported type: ${extra.runtimeType}');
  }
}

class _ExtraDecoder extends Converter<String, Object?> {
  const _ExtraDecoder();

  @override
  Object? convert(String extra) {
    final decoded = jsonDecode(extra) as Map<String, Object?>;
    final type = decoded['type'] as String;

    switch (type) {
      case 'primitive':
        return decoded['data'];

      default:
        throw UnsupportedError('Unsupported type: $type');
    }
  }
}
