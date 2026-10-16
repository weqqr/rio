package react

type EffectScope struct {
	effects []*Effect
	scopes  []*EffectScope
	dispose []func()
	parent  *EffectScope
	active  bool
}

var activeScope *EffectScope

func NewEffectScope() *EffectScope {
	s := &EffectScope{active: true}
	if activeScope != nil && activeScope.active {
		s.parent = activeScope
		activeScope.scopes = append(activeScope.scopes, s)
	}
	return s
}

func (s *EffectScope) Run(fn func()) {
	if !s.active {
		return
	}
	previous := activeScope
	activeScope = s
	defer func() { activeScope = previous }()
	fn()
}

func (s *EffectScope) Stop() {
	if !s.active {
		return
	}
	s.active = false
	if s.parent != nil {
		s.parent.forgetScope(s)
		s.parent = nil
	}
	effects, scopes, dispose := s.effects, s.scopes, s.dispose
	s.effects, s.scopes, s.dispose = nil, nil, nil
	for _, e := range effects {
		e.Stop()
	}
	for _, child := range scopes {
		child.Stop()
	}
	for _, fn := range dispose {
		fn()
	}
}

func (s *EffectScope) Active() bool { return s.active }

func (s *EffectScope) forget(e *Effect) {
	for i, item := range s.effects {
		if item == e {
			s.effects = append(s.effects[:i], s.effects[i+1:]...)
			return
		}
	}
}

func (s *EffectScope) forgetScope(child *EffectScope) {
	for i, item := range s.scopes {
		if item == child {
			s.scopes = append(s.scopes[:i], s.scopes[i+1:]...)
			return
		}
	}
}

func OnScopeDispose(fn func()) {
	if activeScope != nil && activeScope.active {
		activeScope.dispose = append(activeScope.dispose, fn)
	}
}

func recordScope(e *Effect) {
	if activeScope == nil || !activeScope.active {
		return
	}
	e.scope = activeScope
	activeScope.effects = append(activeScope.effects, e)
}
