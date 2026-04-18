import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

@immutable
class AppTypography extends ThemeExtension<AppTypography> {
  const AppTypography({required this.code});

  final TextStyle code;

  static AppTypography fallback(ColorScheme colorScheme) {
    return AppTypography(
      code: GoogleFonts.googleSansCode(
        textStyle: TextStyle(fontSize: 13, fontWeight: FontWeight.w500, color: colorScheme.onSurface),
      ),
    );
  }

  @override
  AppTypography copyWith({TextStyle? code}) {
    return AppTypography(code: code ?? this.code);
  }

  @override
  AppTypography lerp(ThemeExtension<AppTypography>? other, double t) {
    if (other is! AppTypography) {
      return this;
    }
    return AppTypography(code: TextStyle.lerp(code, other.code, t) ?? code);
  }
}

TextTheme buildAppTextTheme(TextTheme base) {
  return base.copyWith(
    displayLarge: GoogleFonts.googleSans(textStyle: base.displayLarge),
    displayMedium: GoogleFonts.googleSans(textStyle: base.displayMedium),
    displaySmall: GoogleFonts.googleSans(textStyle: base.displaySmall),
    headlineLarge: GoogleFonts.googleSans(textStyle: base.headlineLarge),
    headlineMedium: GoogleFonts.googleSans(textStyle: base.headlineMedium),
    headlineSmall: GoogleFonts.googleSans(textStyle: base.headlineSmall),
    titleLarge: GoogleFonts.googleSans(textStyle: base.titleLarge),
    titleMedium: GoogleFonts.googleSans(textStyle: base.titleMedium),
    titleSmall: GoogleFonts.googleSans(textStyle: base.titleSmall),
    bodyLarge: GoogleFonts.inter(textStyle: base.bodyLarge),
    bodyMedium: GoogleFonts.inter(textStyle: base.bodyMedium),
    bodySmall: GoogleFonts.inter(textStyle: base.bodySmall),
    labelLarge: GoogleFonts.inter(textStyle: base.labelLarge),
    labelMedium: GoogleFonts.inter(textStyle: base.labelMedium),
    labelSmall: GoogleFonts.googleSansCode(textStyle: base.labelSmall),
  );
}
