package progress

// Section is a group of tasks rendered together with a shared column layout
// and optional indentation. Sections render top-to-bottom in the order they
// are added.
type Section struct {
	idx       int
	columns   []Column
	indent    int
	taskOrder []TaskID
	progress  *Progress
}

// SectionOption configures a Section.
type SectionOption func(*Section)

// WithSectionColumns sets the columns for this section.
func WithSectionColumns(cols ...Column) SectionOption {
	return func(s *Section) {
		s.columns = cols
	}
}

// WithSectionIndent sets the number of leading space characters prepended
// to each task line in this section. Default is 0 (flush left).
func WithSectionIndent(n int) SectionOption {
	return func(s *Section) {
		s.indent = n
	}
}

// AddTask adds a task to this section and returns its globally-unique TaskID.
func (s *Section) AddTask(description string, total *float64, opts ...TaskOption) TaskID {
	p := s.progress
	p.mu.Lock()
	defer p.mu.Unlock()

	return s.addTaskLocked(p, description, total, opts...)
}

func (s *Section) addTaskLocked(p *Progress, description string, total *float64, opts ...TaskOption) TaskID {
	cfg := TaskConfig{
		Description: description,
		Total:       total,
		Visible:     true,
		Start:       true,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	id := p.nextTaskID
	p.nextTaskID++

	task := NewTask(id, cfg, p.getTime, p.speedEstimatePeriod)
	p.tasks[id] = task
	s.taskOrder = append(s.taskOrder, id)

	return id
}

func (s *Section) removeID(taskID TaskID) {
	for i, id := range s.taskOrder {
		if id == taskID {
			s.taskOrder = append(s.taskOrder[:i], s.taskOrder[i+1:]...)
			return
		}
	}
}
