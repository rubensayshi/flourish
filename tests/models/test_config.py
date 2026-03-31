from rdruid_analyzer.models.config import TalentConfig, load_config


def test_load_config(tmp_path):
    yaml_content = """
soul_of_the_forest:
  skip: false
  multiplier: 0.6
wild_growth:
  skip: true
  skip_reason: "always take"
"""
    p = tmp_path / "talents.yaml"
    p.write_text(yaml_content)
    config = load_config(str(p))
    assert not config["soul_of_the_forest"].skip
    assert config["soul_of_the_forest"].multiplier == 0.6
    assert config["wild_growth"].skip


def test_missing_talent_uses_defaults():
    tc = TalentConfig()
    assert not tc.skip
    assert tc.multiplier is None
