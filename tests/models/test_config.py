from flourish.models.config import TalentConfig, load_config


def test_load_config(tmp_path):
    yaml_content = """
mastery:
  base_stacks: 4

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
    assert config.mastery.base_stacks == 4
    assert not config.talents["soul_of_the_forest"].skip
    assert config.talents["soul_of_the_forest"].multiplier == 0.6
    assert config.talents["wild_growth"].skip


def test_missing_talent_uses_defaults():
    tc = TalentConfig()
    assert not tc.skip
    assert tc.multiplier is None
