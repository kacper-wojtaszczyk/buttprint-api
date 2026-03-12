package domain

type ScorerStub struct{}

func (s *ScorerStub) RequiredVariables() []string {
	return []string{"pm2p5", "pm10"}
}

func (s *ScorerStub) Calculate(variableData []VariableData) (Score, error) {
	return Score{
		Composite:  0.5,
		Thickness:  0.5,
		Sweatiness: 0.5,
		Irritation: 0.5,
	}, nil
}

type RendererStub struct{}

func (r *RendererStub) Render(scores Score) (string, error) {
	return `<svg width="200" height="200" viewBox="0 0 220 220" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <radialGradient id="peachGrad" cx="40%" cy="35%" r="70%">
      <stop offset="0%" stop-color="#FFD0D0"/>
      <stop offset="70%" stop-color="#F7B7B7"/>
      <stop offset="100%" stop-color="#F1A8A8"/>
    </radialGradient>
  </defs>

  <path d="M110 45
           C70 45, 35 80, 35 125
           C35 175, 75 195, 110 170
           C145 195, 185 175, 185 125
           C185 80, 150 45, 110 45Z"
        fill="url(#peachGrad)" stroke="#D68F8F" stroke-width="4" stroke-linejoin="round"/>

  <ellipse cx="72" cy="112" rx="16" ry="12" fill="#ffffff" opacity="0.30"/>
  <ellipse cx="148" cy="112" rx="16" ry="12" fill="#ffffff" opacity="0.30"/>

  <ellipse cx="78" cy="140" rx="28" ry="22" fill="#F2A3A3" opacity="0.45"/>
  <ellipse cx="142" cy="140" rx="28" ry="22" fill="#F2A3A3" opacity="0.45"/>

  <path d="M110 70 C104 110 104 138 110 168"
        stroke="#D68F8F" stroke-width="5" fill="none" stroke-linecap="round"/>
</svg>
`, nil
}
