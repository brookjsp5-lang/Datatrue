package usecases_physical_mysql

type CommandPlan struct {
	Executable string
	Args       []string
	Env        []string
}
