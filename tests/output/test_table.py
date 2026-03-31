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


def test_render_overlap_disclaimer_when_exceeding_total():
    results = AnalysisResults(
        total_healing=100000,
        wasted=0,
        talent_healing={"Talent A": 80000.0, "Talent B": 50000.0},
        fight_duration_ms=300000,
    )
    output = render_results(results)
    assert "Talents can overlap" in output


def test_render_no_disclaimer_when_within_total():
    results = AnalysisResults(
        total_healing=100000,
        wasted=0,
        talent_healing={"Talent A": 30000.0, "Talent B": 20000.0},
        fight_duration_ms=300000,
    )
    output = render_results(results)
    assert "Talents can overlap" not in output


def test_render_unattributed_dash_when_negative():
    results = AnalysisResults(
        total_healing=100000,
        wasted=0,
        talent_healing={"Talent A": 80000.0, "Talent B": 50000.0},
        fight_duration_ms=300000,
    )
    output = render_results(results)
    # Unattributed should show "—" not a number
    lines = [line for line in output.split("\n") if "Unattributed" in line]
    assert len(lines) == 1
    # Should not contain a healing number for unattributed
    assert "—" in lines[0]
