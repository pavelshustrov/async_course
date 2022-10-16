package accounts

type TaskRepo interface {
}

type UserRepo interface {
}

type AccountRepo interface {
}

type handler struct {
	taskRepo TaskRepo
	userRepo UserRepo
}

func New() *handler {
	return &handler{}
}

func (h *handler) Fee() {

}

func (h *handler) Award() {

}
