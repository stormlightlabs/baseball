import 'package:bigfly_mobile/colors.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('team enum includes all 30 MLB clubs', () {
    expect(MlbTeam.values.length, 30);
    expect(mlbTeamPrimaryHex.length, 30);
  });

  test('team lookup exposes primary color data', () {
    final cubs = MlbTeam.fromCode('CHC');
    expect(cubs, isNotNull);
    expect(cubs!.primaryHex, '#0E3386');
    expect(teamPrimaryColor('CHC'), isNotNull);
  });

  test('legacy team codes normalize to modern team colors', () {
    expect(normalizeMlbTeamCode('NYA'), 'NYY');
    expect(normalizeMlbTeamCode('CAL'), 'LAA');
    expect(teamPrimaryColor('SFN'), isNotNull);
  });
}
