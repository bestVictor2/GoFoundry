package session

type BeforeInsertHook interface {
	BeforeInsert(*Session) error
}

type AfterInsertHook interface {
	AfterInsert(*Session) error
}

type BeforeQueryHook interface {
	BeforeQuery(*Session) error
}

type AfterQueryHook interface {
	AfterQuery(*Session) error
}

func callBeforeInsert(model interface{}, s *Session) error {
	if hook, ok := model.(BeforeInsertHook); ok {
		return hook.BeforeInsert(s)
	}
	return nil
}

func callAfterInsert(model interface{}, s *Session) error {
	if hook, ok := model.(AfterInsertHook); ok {
		return hook.AfterInsert(s)
	}
	return nil
}

func callBeforeQuery(model interface{}, s *Session) error {
	if hook, ok := model.(BeforeQueryHook); ok {
		return hook.BeforeQuery(s)
	}
	return nil
}

func callAfterQuery(model interface{}, s *Session) error {
	if hook, ok := model.(AfterQueryHook); ok {
		return hook.AfterQuery(s)
	}
	return nil
}
