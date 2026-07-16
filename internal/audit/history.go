package audit

type Recorder struct{}

func NewRecorder() *Recorder {

	return &Recorder{}
}

func (r *Recorder) Record() {

	// later
}