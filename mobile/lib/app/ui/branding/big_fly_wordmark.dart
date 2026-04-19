import 'package:flutter/material.dart';

class BigFlyWordmark extends StatelessWidget {
  const BigFlyWordmark({
    super.key,
    required this.flyColor,
    this.bigColor = Colors.white,
    this.fontSize = 18,
    this.fontWeight = FontWeight.w700,
  });

  final Color flyColor;
  final Color bigColor;
  final double fontSize;
  final FontWeight fontWeight;

  @override
  Widget build(BuildContext context) {
    final baseStyle = Theme.of(
      context,
    ).textTheme.titleLarge?.copyWith(fontSize: fontSize, fontWeight: fontWeight, letterSpacing: -0.2);

    return Text.rich(
      TextSpan(
        style: baseStyle,
        children: <InlineSpan>[
          TextSpan(
            text: 'Big ',
            style: TextStyle(color: bigColor),
          ),
          TextSpan(
            text: 'Fly',
            style: TextStyle(color: flyColor),
          ),
        ],
      ),
      maxLines: 1,
      overflow: TextOverflow.fade,
      softWrap: false,
    );
  }
}
