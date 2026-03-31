from rdruid_analyzer.output.table import render_results
from rdruid_analyzer.analysis.pipeline import AnalysisResults


def test_render_results_returns_string():
    results = AnalysisResults(
        total_healing=100000,
        wasted=10000,
        talent_healing={"Soul of the Forest": 15000.0, "Cultivation": 8000.0},
        fight_duration_ms=300000,
    )
    output = render_results(results, fight_name="Mythic Boss", player_name="TestDruid")
    assert "Soul of the Forest" in output
    assert "Cultivation" in output
    assert "Wasted" in output
