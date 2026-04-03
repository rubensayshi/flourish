package models

import (
	"os"

	"gopkg.in/yaml.v3"
)

type MasteryConfig struct {
	BaseStacks int       `yaml:"base_stacks"`
	DRTable    []float64 `yaml:"dr_table"`
}

type TalentConfig struct {
	Skip       bool     `yaml:"skip"`
	SkipReason string   `yaml:"skip_reason"`
	Multiplier *float64 `yaml:"multiplier"`
}

type Config struct {
	Mastery MasteryConfig
	Talents map[string]TalentConfig
}

var defaultDRTable = []float64{1.0, 1.7, 2.3, 2.8, 3.2}

func DefaultConfig() *Config {
	return &Config{
		Mastery: MasteryConfig{BaseStacks: 2, DRTable: defaultDRTable},
		Talents: map[string]TalentConfig{},
	}
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	config := &Config{
		Mastery: MasteryConfig{
			BaseStacks: 2,
			DRTable:    defaultDRTable,
		},
		Talents: make(map[string]TalentConfig),
	}

	// Parse mastery section
	if masteryRaw, ok := raw["mastery"]; ok {
		masteryBytes, err := yaml.Marshal(masteryRaw)
		if err == nil {
			yaml.Unmarshal(masteryBytes, &config.Mastery)
		}
		delete(raw, "mastery")
	}

	// Remaining keys are talent configs
	for name, v := range raw {
		tc := TalentConfig{}
		if v != nil {
			valBytes, err := yaml.Marshal(v)
			if err == nil {
				yaml.Unmarshal(valBytes, &tc)
			}
		}
		config.Talents[name] = tc
	}

	return config, nil
}
